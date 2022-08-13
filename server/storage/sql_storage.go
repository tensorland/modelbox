package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/fatih/structs"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	EXPERIMENTS_LIST = "SELECT id, name, owner, namespace, external_id, ml_framework, created_at, updated_at from experiments where namespace = ?"

	CHECKPOINTS_LIST = `select id, experiment, epoch, metrics, created_at, updated_at from checkpoints 
	                      where experiment = ?`

	MODEL_GET = "select id, name, owner, namespace, task, description, created_at, updated_at from models where id = ?"

	MODELS_NS_LIST = "select id, name, owner, namespace, task, description, created_at, updated_at from models where namespace = ?"

	MODEL_VERSION_CREATE = `insert into model_versions(id, name, model_id, version, description, ml_framework, unique_tags, created_at, updated_at) values(:id, :name, :model_id, :version, :description, :ml_framework, :unique_tags, :created_at, :updated_at)`

	MODEL_VERSION_GET = "select name, model_id, version, description, ml_framework, unique_tags, created_at, updated_at from model_versions where id = ?"

	BLOB_MULTI_WRITE = "insert into blobs(id, parent_id, metadata) VALUES "

	BLOBSET_GET = "select id, parent_id, metadata from blobs where parent_id=?"

	BLOB_GET = "select id, parent_id, metadata from blobs where id=?"

	METADATA_UPDATE = "INSERT INTO metadata (id, parent_id, metadata) VALUES(:id, :parent_id, :metadata) ON DUPLICATE KEY UPDATE id=VALUES(`id`), parent_id=VALUES(`parent_id`), metadata=VALUES(`metadata`)"

	METADATA_LIST = "select id, parent_id, metadata from metadata where parent_id=?"

	EVENT_CREATE = "insert into mutation_events (mutation_time, action, object_id, object_type, parent_id, namespace, payload) VALUES(:mutation_time, :action, :object_id, :object_type, :parent_id, :namespace, :payload)"

	EVENT_NS_LIST = "select mutation_id, mutation_time, action, object_id, object_type, parent_id, namespace, payload from mutation_events where namespace =? and mutation_time>=?"
)

type driverUtils interface {
	isDuplicate(err error) bool

	updateMetadata() string

	createExperiment() string

	createCheckpoint() string

	createModel() string
}

type SQLStorage struct {
	driverUtils
	db *sqlx.DB

	logger *zap.Logger
}

func NewSQLStorage(db *sqlx.DB, driverUtils driverUtils, logger *zap.Logger) *SQLStorage {
	return &SQLStorage{db: db, driverUtils: driverUtils, logger: logger}
}

func (s *SQLStorage) CreateExperiment(
	ctx context.Context,
	experiment *Experiment,
	metadata SerializableMetadata,
) (*CreateExperimentResult, error) {
	result := &CreateExperimentResult{}
	err := s.transact(ctx, func(tx *sqlx.Tx) error {
		schema := FromExperimentToSchema(experiment)
		_, err := tx.NamedExec(
			s.createExperiment(),
			schema,
		)
		if err != nil {
			if s.driverUtils.isDuplicate(err) {
				result.Exists = true
				result.ExperimentId = experiment.Id
				return nil
			}
			return fmt.Errorf("unable to write experiment to db: %v", err)
		}
		result.ExperimentId = experiment.Id
		if err := s.writeMetadata(tx, result.ExperimentId, metadata); err != nil {
			return fmt.Errorf("can't write metadata: %v", err)
		}
		event := &MutationEventSchema{
			MutationTime: uint64(time.Now().Unix()),
			Action:       "create",
			ObjectType:   "experiment",
			ObjectId:     experiment.Id,
			Namespace:    experiment.Namespace,
			Payload:      structs.Map(experiment),
		}
		if err := s.createMutationEvent(ctx, tx, event); err != nil {
			return fmt.Errorf("unable to create mutation for experiment: %v", err)
		}
		return nil
	})
	return result, err
}

func (s *SQLStorage) CreateCheckpoint(
	ctx context.Context,
	c *Checkpoint,
	metadata SerializableMetadata,
) (*CreateCheckpointResult, error) {
	err := s.transact(ctx, func(tx *sqlx.Tx) error {
		cs := ToCheckpointSchema(c)
		_, err := tx.NamedExec(s.createCheckpoint(), cs)
		if err != nil {
			if s.driverUtils.isDuplicate(err) {
				return nil
			}
			return fmt.Errorf("unable to write checkpoint: %v", err)
		}
		if err := s.writeMetadata(tx, c.Id, metadata); err != nil {
			return fmt.Errorf("can't write metadata: %v", err)
		}
		return s.writeFileSet(tx, c.Files)
	})
	return &CreateCheckpointResult{CheckpointId: c.Id}, err
}

func (s *SQLStorage) ListExperiments(
	ctx context.Context,
	namespace string,
) ([]*Experiment, error) {
	experiments := make([]*Experiment, 0)
	err := s.transact(ctx, func(tx *sqlx.Tx) error {
		rows := []ExperimentSchema{}
		if err := tx.Select(&rows, s.db.Rebind(EXPERIMENTS_LIST), namespace); err != nil {
			return err
		}
		for _, row := range rows {
			experiments = append(experiments, row.ToExperiment())
		}
		return nil
	})
	return experiments, err
}

func (s *SQLStorage) ListCheckpoints(
	ctx context.Context,
	experimentId string,
) ([]*Checkpoint, error) {
	checkpoints := make([]*Checkpoint, 0)
	err := s.transact(ctx, func(tx *sqlx.Tx) error {
		rows := []CheckpointSchema{}
		if err := s.db.Select(&rows, s.db.Rebind(CHECKPOINTS_LIST), experimentId); err != nil {
			return err
		}
		for _, row := range rows {
			files, err := s.getFileSetForParent(tx, row.Id)
			if err != nil {
				return err
			}
			checkpoints = append(checkpoints, row.ToCheckpoint(files))
		}
		return nil
	})

	return checkpoints, err
}

func (s *SQLStorage) GetCheckpoint(
	ctx context.Context,
	checkpointId string,
) (*Checkpoint, error) {
	var checkpoint *Checkpoint
	err := s.transact(ctx, func(tx *sqlx.Tx) error {
		var checkpointSchema CheckpointSchema
		if err := tx.Select(&checkpointSchema, CHECKPOINTS_LIST, checkpointId); err != nil {
			return err
		}
		rows := []FileSchema{}
		if err := tx.Select(&rows, s.db.Rebind(BLOBSET_GET), checkpointSchema.Id); err != nil {
			return err
		}
		files, err := ToFileSet(rows)
		if err != nil {
			return err
		}
		checkpoint = checkpointSchema.ToCheckpoint(files)
		return nil
	})
	return checkpoint, err
}

func (s *SQLStorage) CreateModel(ctx context.Context, model *Model, metadata SerializableMetadata) (*CreateModelResult, error) {
	err := s.transact(ctx, func(tx *sqlx.Tx) error {
		schema := ModelToSchema(model)
		if _, err := tx.NamedExec(s.createModel(), schema); err != nil {
			if s.driverUtils.isDuplicate(err) {
				return nil
			}
			return fmt.Errorf("unable to create model: %v", err)
		}
		if err := s.writeMetadata(tx, model.Id, metadata); err != nil {
			return fmt.Errorf("can't write metadata: %v", err)
		}
		return s.writeFileSet(tx, model.Files)
	})
	return &CreateModelResult{ModelId: model.Id}, err
}

func (s *SQLStorage) GetModel(ctx context.Context, id string) (*Model, error) {
	var model *Model
	err := s.transact(ctx, func(tx *sqlx.Tx) error {
		modelSchema := ModelSchema{}
		if err := tx.Get(&modelSchema, s.db.Rebind(MODEL_GET), id); err != nil {
			return err
		}
		fileSet, err := s.getFileSetForParent(tx, id)
		if err != nil {
			return fmt.Errorf("unable to get query fileset: %v", err)
		}
		model = modelSchema.ToModel(fileSet)
		return nil
	})
	return model, err
}

func (s *SQLStorage) ListModels(ctx context.Context, namespace string) ([]*Model, error) {
	models := make([]*Model, 0)
	err := s.transact(ctx, func(tx *sqlx.Tx) error {
		modelRows := []ModelSchema{}
		if err := tx.Select(&modelRows, s.db.Rebind(MODELS_NS_LIST), namespace); err != nil {
			return fmt.Errorf("can't query: %v", err)
		}
		for _, modelRow := range modelRows {
			fileSet, err := s.getFileSetForParent(tx, modelRow.Id)
			if err != nil {
				return err
			}
			models = append(models, modelRow.ToModel(fileSet))
		}
		return nil
	})
	return models, err
}

func (s *SQLStorage) CreateModelVersion(
	ctx context.Context,
	modelVersion *ModelVersion,
	metadata SerializableMetadata,
) (*CreateModelVersionResult, error) {
	err := s.transact(ctx, func(tx *sqlx.Tx) error {
		schema := ModelVersionToSchema(modelVersion)
		if _, err := tx.NamedExec(
			MODEL_VERSION_CREATE,
			schema,
		); err != nil {
			if s.driverUtils.isDuplicate(err) {
				return nil
			}
			return fmt.Errorf("unable to create model version: %v", err)
		}
		if err := s.writeMetadata(tx, modelVersion.Id, metadata); err != nil {
			return fmt.Errorf("can't write metadata: %v", err)
		}
		return s.writeFileSet(tx, modelVersion.Files)
	})
	if err != nil {
		return nil, err
	}
	return &CreateModelVersionResult{ModelVersionId: modelVersion.Id}, err
}

func (s *SQLStorage) GetModelVersion(ctx context.Context, id string) (*ModelVersion, error) {
	var modelVersion *ModelVersion
	err := s.transact(ctx, func(tx *sqlx.Tx) error {
		var modelVersionSchema ModelVersionSchema
		if err := tx.Get(&modelVersionSchema, s.db.Rebind(MODEL_VERSION_GET), id); err != nil {
			return err
		}
		fileSet, err := s.getFileSetForParent(tx, id)
		if err != nil {
			return err
		}
		modelVersion = modelVersionSchema.ToModelVersion(fileSet)
		return err
	})
	return modelVersion, err
}

func (s *SQLStorage) ListModelVersions(
	ctx context.Context,
	model string,
) ([]*ModelVersion, error) {
	return nil, nil
}

func (e *SQLStorage) WriteFiles(ctx context.Context, blobs FileSet) error {
	return e.transact(ctx, func(tx *sqlx.Tx) error {
		return e.writeFileSet(tx, blobs)
	})
}

func (e *SQLStorage) GetFiles(ctx context.Context, parentId string) (FileSet, error) {
	var blobs FileSet
	err := e.transact(ctx, func(tx *sqlx.Tx) error {
		blobSet, err := e.getFileSetForParent(tx, parentId)
		blobs = blobSet
		return err
	})
	return blobs, err
}

func (s *SQLStorage) GetFile(ctx context.Context, id string) (*FileMetadata, error) {
	var blob FileMetadata
	err := s.transact(ctx, func(tx *sqlx.Tx) error {
		var blobRow FileSchema
		if err := tx.Get(&blobRow, s.db.Rebind(BLOB_GET), id); err != nil {
			return fmt.Errorf("unable to get query blobs: %v", err)
		}

		b, err := blobRow.ToFile()
		if err != nil {
			return fmt.Errorf("unable to convert blobschema to blob: %v", err)
		}
		blob = *b
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to retreieve blob: %v", err)
	}
	return &blob, nil
}

func (s *SQLStorage) writeMetadata(tx *sqlx.Tx, parentId string, metadata map[string]*structpb.Value) error {
	rows := toMetadataSchema(parentId, metadata)
	for _, row := range rows {
		if _, err := tx.NamedExec(s.updateMetadata(), row); err != nil {
			return fmt.Errorf("unable to write to db: %v", err)
		}
	}
	return nil
}

func (s *SQLStorage) UpdateMetadata(ctx context.Context, parentId string, metadata map[string]*structpb.Value) error {
	return s.transact(ctx, func(tx *sqlx.Tx) error {
		return s.writeMetadata(tx, parentId, metadata)
	})
}

func (s *SQLStorage) ListMetadata(ctx context.Context, parentId string) (map[string]*structpb.Value, error) {
	meta := map[string]*structpb.Value{}
	err := s.transact(ctx, func(tx *sqlx.Tx) error {
		rows := []*MetadataSchema{}
		if err := s.db.Select(&rows, s.db.Rebind(METADATA_LIST), parentId); err != nil {
			return err
		}
		meta = toMetadata(rows)
		return nil
	})
	return meta, err
}

func (s *SQLStorage) createMutationEvent(ctx context.Context, tx *sqlx.Tx, event *MutationEventSchema) error {
	_, err := tx.NamedExec(EVENT_CREATE, event)
	return err
}

func (s *SQLStorage) ListChanges(ctx context.Context, namespace string, since time.Time) ([]*ChangeEvent, error) {
	rows := []MutationEventSchema{}
	if err := s.db.Select(&rows, s.db.Rebind(EVENT_NS_LIST), namespace, since.Unix()); err != nil {
		return nil, fmt.Errorf("unable to list mutation events: %v", err)
	}

	result := make([]*ChangeEvent, len(rows))
	for i, row := range rows {
		result[i] = &ChangeEvent{
			ObjectId:   row.ObjectId,
			ObjectType: row.ObjectType,
			Action:     row.Action,
			Payload:    row.Payload,
			Time:       time.Unix(int64(row.MutationTime), 0),
			Namespace:  row.Namespace,
		}
	}
	return result, nil
}

func (s *SQLStorage) getFileSetForParent(tx *sqlx.Tx, parentId string) (FileSet, error) {
	blobRows := []FileSchema{}
	if err := tx.Select(&blobRows, s.db.Rebind(BLOBSET_GET), parentId); err != nil {
		return nil, fmt.Errorf("unable to get query blobset: %v", err)
	}
	blobSet, err := ToFileSet(blobRows)
	if err != nil {
		return nil, err
	}
	return blobSet, nil
}

func (s *SQLStorage) writeFileSet(tx *sqlx.Tx, files FileSet) error {
	if files == nil {
		return nil
	}
	vals := []interface{}{}
	sqlStr := BLOB_MULTI_WRITE
	for _, file := range files {
		bJson, err := file.ToJson()
		if err != nil {
			return fmt.Errorf("can't serialize blob to json :%v", err)
		}
		sqlStr += "(?, ?, ?),"
		vals = append(vals, file.Id, file.ParentId, bJson)
	}
	sqlStr = sqlStr[0 : len(sqlStr)-1]
	if len(files) > 0 {
		if _, err := tx.Exec(s.db.Rebind(sqlStr), vals...); err != nil {
			return fmt.Errorf("unable to create blobs for model: %v", err)
		}
	}

	return nil
}

func (s *SQLStorage) transact(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}
