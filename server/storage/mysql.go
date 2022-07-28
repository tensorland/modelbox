package storage

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/VividCortex/mysqlerr"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"go.uber.org/zap"
)

const (
	MYSQL_DRIVER = "mysql"

	EXPERIMENT_CREATE = "insert into experiments(id, name, owner, namespace, external_id, ml_framework, metadata, created_at, updated_at) values(:id, :name, :owner, :namespace, :external_id, :ml_framework, :metadata, :created_at, :updated_at)"

	EXPERIMENTS_LIST = "SELECT id, name, owner, namespace, external_id, ml_framework, metadata, created_at, updated_at from experiments where namespace = ?"

	EXPERIMENT_GET = "SELECT name, owner, namespace, external_id, ml_framework, metadata, created_at, updated_at from experiments where id = ?"

	EXPERIMENTS_DELETE = "delete from experiments where id=?"

	CHECKPOINTS_CREATE = "insert into checkpoints(id, experiment, epoch, metrics, metadata, created_at, updated_at) values(:id, :experiment, :epoch, :metrics, :metadata, :created_at, :updated_at)"

	CHECKPOINTS_LIST = `select id, experiment, epoch, metrics, metadata, created_at, updated_at from checkpoints 
	                      where experiment = ?`

	CHECKPOINT_UPDATE_PATH = "update checkpoints set path = ? where id = ?"

	MODEL_CREATE = "insert into models(id, name, owner, namespace, task, metadata, description, created_at, updated_at) values(:id, :name, :owner, :namespace, :task, :metadata, :description, :created_at, :updated_at)"

	MODEL_GET = "select id, name, owner, namespace, task, metadata, description, created_at, updated_at from models where id = ?"

	MODELS_NS_LIST = "select id, name, owner, namespace, task, metadata, description, created_at, updated_at from models where namespace = ?"

	MODEL_VERSION_CREATE = `insert into model_versions(id, name, model_id, version, description, ml_framework, metadata, unique_tags, created_at, updated_at) values(:id, :name, :model_id, :version, :description, :ml_framework, :metadata, :unique_tags, :created_at, :updated_at)`

	MODEL_VERSION_GET = "select name, model_id, version, description, ml_framework, metadata, unique_tags, created_at, updated_at from model_versions where id = ?"

	BLOB_MULTI_WRITE = "insert into blobs(id, parent_id, metadata) VALUES "

	BLOBSET_GET = "select id, parent_id, metadata from blobs where parent_id=?"
)

type MySqlStorage struct {
	db     *sqlx.DB
	config *MySqlStorageConfig

	logger *zap.Logger
}

func NewMySqlStorage(config *MySqlStorageConfig, logger *zap.Logger) (*MySqlStorage, error) {
	db, err := sqlx.Open(MYSQL_DRIVER, config.DataSource())
	if err != nil {
		return nil, fmt.Errorf("unable to connect to mysql %v", err)
	}
	return &MySqlStorage{db: db, config: config, logger: logger}, nil
}

func (m *MySqlStorage) CreateExperiment(
	ctx context.Context,
	experiment *Experiment,
) (*CreateExperimentResult, error) {
	schema := FromExperimentToSchema(experiment)
	_, err := m.db.NamedExec(
		EXPERIMENT_CREATE,
		schema,
	)
	if err != nil {
		if m.isDuplicateError(err) {
			return &CreateExperimentResult{ExperimentId: experiment.Id, Exists: true}, nil
		}
		return nil, fmt.Errorf("unable to write experiment to db: %v", err)
	}
	return &CreateExperimentResult{ExperimentId: experiment.Id, Exists: false}, nil
}

func (m *MySqlStorage) CreateCheckpoint(
	ctx context.Context,
	c *Checkpoint,
) (*CreateCheckpointResult, error) {
	err := m.transact(ctx, func(tx *sqlx.Tx) error {
		cs := ToCheckpointSchema(c)
		_, err := tx.NamedExec(CHECKPOINTS_CREATE, cs)
		if err != nil {
			if m.isDuplicateError(err) {
				return nil
			}
			return fmt.Errorf("unable to write checkpoint: %v", err)
		}

		return m.writeBlobSet(tx, c.Blobs)
	})
	return &CreateCheckpointResult{CheckpointId: c.Id}, err
}

func (m *MySqlStorage) ListExperiments(
	ctx context.Context,
	namespace string,
) ([]*Experiment, error) {
	experiments := make([]*Experiment, 0)
	err := m.transact(ctx, func(tx *sqlx.Tx) error {
		rows := []ExperimentSchema{}
		if err := tx.Select(&rows, EXPERIMENTS_LIST, namespace); err != nil {
			return err
		}
		for _, row := range rows {
			experiments = append(experiments, row.ToExperiment())
		}
		return nil
	})
	return experiments, err
}

func (m *MySqlStorage) ListCheckpoints(
	ctx context.Context,
	experimentId string,
) ([]*Checkpoint, error) {
	checkpoints := make([]*Checkpoint, 0)
	err := m.transact(ctx, func(tx *sqlx.Tx) error {
		rows := []CheckpointSchema{}
		if err := m.db.Select(&rows, CHECKPOINTS_LIST, experimentId); err != nil {
			return err
		}
		for _, row := range rows {
			blobs, err := m.getBlobSetForParent(tx, row.Id)
			if err != nil {
				return err
			}
			checkpoints = append(checkpoints, row.ToCheckpoint(blobs))
		}
		return nil
	})

	return checkpoints, err
}

func (m *MySqlStorage) GetCheckpoint(
	ctx context.Context,
	checkpointId string,
) (*Checkpoint, error) {
	var checkpoint *Checkpoint
	err := m.transact(ctx, func(tx *sqlx.Tx) error {
		var checkpointSchema CheckpointSchema
		if err := tx.Select(&checkpointSchema, CHECKPOINTS_LIST, checkpointId); err != nil {
			return err
		}
		rows := []BlobSchema{}
		if err := tx.Select(&rows, BLOBSET_GET, checkpointSchema.Id); err != nil {
			return err
		}
		blobs, err := ToBlobSet(rows)
		if err != nil {
			return err
		}
		checkpoint = checkpointSchema.ToCheckpoint(blobs)
		return nil
	})
	return checkpoint, err
}

func (m *MySqlStorage) CreateModel(ctx context.Context, model *Model) (*CreateModelResult, error) {
	err := m.transact(ctx, func(tx *sqlx.Tx) error {
		schema := ModelToSchema(model)
		if _, err := tx.NamedExec(MODEL_CREATE, schema); err != nil {
			if m.isDuplicateError(err) {
				return nil
			}
			return fmt.Errorf("unable to create model: %v", err)
		}
		return m.writeBlobSet(tx, model.Blobs)
	})
	return &CreateModelResult{ModelId: model.Id}, err
}

func (m *MySqlStorage) GetModel(ctx context.Context, id string) (*Model, error) {
	var model *Model
	err := m.transact(ctx, func(tx *sqlx.Tx) error {
		modelSchema := ModelSchema{}
		if err := tx.Get(&modelSchema, MODEL_GET, id); err != nil {
			return err
		}
		blobSet, err := m.getBlobSetForParent(tx, id)
		if err != nil {
			return fmt.Errorf("unable to get query blobset: %v", err)
		}
		model = modelSchema.ToModel(blobSet)
		return nil
	})
	return model, err
}

func (m *MySqlStorage) getBlobSetForParent(tx *sqlx.Tx, parentId string) (BlobSet, error) {
	blobRows := []BlobSchema{}
	if err := tx.Select(&blobRows, BLOBSET_GET, parentId); err != nil {
		return nil, fmt.Errorf("unable to get query blobset: %v", err)
	}
	blobSet, err := ToBlobSet(blobRows)
	if err != nil {
		return nil, err
	}
	return blobSet, nil
}

func (m *MySqlStorage) ListModels(ctx context.Context, namespace string) ([]*Model, error) {
	models := make([]*Model, 0)
	err := m.transact(ctx, func(tx *sqlx.Tx) error {
		modelRows := []ModelSchema{}
		if err := tx.Select(&modelRows, MODELS_NS_LIST, namespace); err != nil {
			return fmt.Errorf("can't query: %v", err)
		}
		for _, modelRow := range modelRows {
			blobSet, err := m.getBlobSetForParent(tx, modelRow.Id)
			if err != nil {
				return err
			}
			models = append(models, modelRow.ToModel(blobSet))
		}
		return nil
	})
	return models, err
}

func (m *MySqlStorage) Ping() error {
	return m.db.Ping()
}

func (m *MySqlStorage) CreateModelVersion(
	ctx context.Context,
	modelVersion *ModelVersion,
) (*CreateModelVersionResult, error) {
	err := m.transact(ctx, func(tx *sqlx.Tx) error {
		schema := ModelVersionToSchema(modelVersion)
		if _, err := tx.NamedExec(
			MODEL_VERSION_CREATE,
			schema,
		); err != nil {
			if m.isDuplicateError(err) {
				return nil
			}
			return fmt.Errorf("unable to create model version: %v", err)
		}
		return m.writeBlobSet(tx, modelVersion.Blobs)
	})
	return &CreateModelVersionResult{ModelVersionId: modelVersion.Id}, err
}

func (m *MySqlStorage) GetModelVersion(ctx context.Context, id string) (*ModelVersion, error) {
	var modelVersion *ModelVersion
	err := m.transact(ctx, func(tx *sqlx.Tx) error {
		var modelVersionSchema ModelVersionSchema
		if err := tx.Get(&modelVersionSchema, MODEL_VERSION_GET, id); err != nil {
			return err
		}
		blobSet, err := m.getBlobSetForParent(tx, id)
		if err != nil {
			return err
		}
		modelVersion = modelVersionSchema.ToModelVersion(blobSet)
		return err
	})
	return modelVersion, err
}

func (m *MySqlStorage) ListModelVersions(
	ctx context.Context,
	model string,
) ([]*ModelVersion, error) {

	return nil, nil
}

func (m *MySqlStorage) Close() error {
	return m.db.Close()
}

func (e *MySqlStorage) Backend() *BackendInfo {
	return &BackendInfo{Name: "mysql"}
}

func (e *MySqlStorage) UpdateBlobPath(
	_ context.Context,
	path string,
	parentId string,
	t BlobType,
) error {
	switch t {
	case CheckpointBlob:
		if _, err := e.db.Exec(CHECKPOINT_UPDATE_PATH, path, parentId); err != nil {
			e.logger.Sugar().Errorf("unable to updat path for blobinfo %v :%v", path, err)
			return err
		}
	case ModelBlob:
		return fmt.Errorf("model path update not implemented yet")

	}
	return nil
}

func (e *MySqlStorage) DeleteExperiment(_ context.Context, id string) error {
	_, err := e.db.Exec(EXPERIMENTS_DELETE, id)
	if err != nil {
		e.logger.Sugar().Errorf("unable to delete experiment: %v %v", id, err.Error())
	}
	return err
}

func (m *MySqlStorage) writeBlobSet(tx *sqlx.Tx, blobs BlobSet) error {
	if blobs == nil {
		return nil
	}
	vals := []interface{}{}
	sqlStr := BLOB_MULTI_WRITE
	for _, blob := range blobs {
		bJson, err := blob.ToJson()
		if err != nil {
			return fmt.Errorf("can't serialize blob to json :%v", err)
		}
		sqlStr += "(?, ?, ?),"
		vals = append(vals, blob.Id, blob.ParentId, bJson)
	}
	sqlStr = sqlStr[0 : len(sqlStr)-1]
	if len(blobs) > 0 {
		if _, err := tx.Exec(sqlStr, vals...); err != nil {
			return fmt.Errorf("unable to create blobs for model: %v", err)
		}
	}

	return nil
}

func (e *MySqlStorage) WriteBlobs(ctx context.Context, blobs BlobSet) error {
	return e.transact(ctx, func(tx *sqlx.Tx) error {
		return e.writeBlobSet(tx, blobs)
	})
}

func (e *MySqlStorage) GetBlobs(ctx context.Context, parentId string) (BlobSet, error) {
	var blobs BlobSet
	err := e.transact(ctx, func(tx *sqlx.Tx) error {
		blobSet, err := e.getBlobSetForParent(tx, parentId)
		blobs = blobSet
		return err
	})
	return blobs, err
}

func (m *MySqlStorage) transact(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func (e *MySqlStorage) DropDb() error {
	db, err := sqlx.Open(MYSQL_DRIVER, e.config.dsnAdmin())
	if err != nil {
		return fmt.Errorf("unable to connec to db: %v", err)
	}
	defer db.Close()
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE %s", e.config.DbName))
	return err
}

func (e *MySqlStorage) CreateSchema(path string) error {
	db, err := sqlx.Open(MYSQL_DRIVER, e.config.dsnAdmin())
	if err != nil {
		return fmt.Errorf("unable to connec to db: %v", err)
	}
	defer db.Close()
	queries := []string{
		fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", e.config.DbName),
		fmt.Sprintf("USE %s", e.config.DbName),
	}
	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("unable to execute query: %v err: %v", query, err)
		}
	}
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("unable to read schema: %v", err)
	}
	queries = strings.Split(string(buf), ";")
	for _, query := range queries {
		if strings.TrimSpace(query) == "" {
			continue
		}
		if _, err := e.db.Exec(query); err != nil {
			return fmt.Errorf("unable to execute query: %v", err)
		}
	}
	return nil
}

func (s *MySqlStorage) isDuplicateError(err error) bool {
	if driverErr, ok := err.(*mysql.MySQLError); ok {
		if driverErr.Number == mysqlerr.ER_DUP_ENTRY {
			return true
		}
	}
	return false
}
