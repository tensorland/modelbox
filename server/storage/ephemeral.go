package storage

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ugorji/go/codec"
	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"
)

var (
	EXPERIMENTS    = []byte("experiments")
	CHECKPOINTS    = []byte("checkpoints")
	MODELS         = []byte("models")
	MODEL_VERSIONS = []byte("model_versions")
	FILES          = []byte("files")
	METADATA       = []byte("metadata")
	MUTATIONS      = []byte("mutations")
)

// itob returns an 8-byte big endian representation of v.
func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

type EphemeralStorage struct {
	db *bolt.DB

	logger *zap.Logger
}

func NewEphemeralStorage(path string, logger *zap.Logger) (*EphemeralStorage, error) {
	db, err := bolt.Open(path, 0666, nil)
	if err != err {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(EXPERIMENTS); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(CHECKPOINTS); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(MODELS); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(MODEL_VERSIONS); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(FILES); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &EphemeralStorage{db: db, logger: logger}, nil
}

func (e *EphemeralStorage) Close() error {
	return e.db.Close()
}

func (e *EphemeralStorage) CreateExperiment(_ context.Context, experiment *Experiment) (*CreateExperimentResult, error) {
	id := experiment.Hash()
	event := &ChangeEvent{
		ObjectId:   experiment.Id,
		ObjectType: "experiment",
		Action:     "create",
		Time:       time.Now(),
		Payload:    experiment,
		Namespace:  experiment.Namespace,
	}
	if err := e.writeBytes(experiment, id, EXPERIMENTS, event); err != nil {
		return nil, err
	}
	result := CreateExperimentResult{
		ExperimentId: id,
	}
	return &result, nil
}

func (e *EphemeralStorage) GetExperiment(_ context.Context, id string) (*Experiment, error) {
	var experiment Experiment
	err := e.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(EXPERIMENTS)
		exp := b.Get([]byte(id))
		decoder := codec.NewDecoderBytes(exp, new(codec.MsgpackHandle))
		return decoder.Decode(&experiment)
	})
	return &experiment, err
}

func (e *EphemeralStorage) ListExperiments(_ context.Context, namespace string) ([]*Experiment, error) {
	experiments := make([]*Experiment, 0)
	err := e.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(EXPERIMENTS)
		c := b.Cursor()
		handle := new(codec.MsgpackHandle)
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var experiment Experiment
			decoder := codec.NewDecoderBytes(v, handle)
			if err := decoder.Decode(&experiment); err != nil {
				return err
			}
			if experiment.Namespace != namespace {
				continue
			}
			experiments = append(experiments, &experiment)
		}
		return nil
	})
	return experiments, err
}

func (e *EphemeralStorage) CreateCheckpoint(_ context.Context, checkpoint *Checkpoint) (*CreateCheckpointResult, error) {
	if err := e.writeBytes(checkpoint, checkpoint.Id, CHECKPOINTS, nil); err != nil {
		return nil, err
	}

	return &CreateCheckpointResult{CheckpointId: checkpoint.Id}, nil
}

func (e *EphemeralStorage) ListCheckpoints(_ context.Context, experimentId string) ([]*Checkpoint, error) {
	checkpoints := make([]*Checkpoint, 0)
	err := e.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(CHECKPOINTS)
		c := b.Cursor()
		handle := new(codec.MsgpackHandle)
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var checkpoint Checkpoint
			decoder := codec.NewDecoderBytes(v, handle)
			if err := decoder.Decode(&checkpoint); err != nil {
				return err
			}
			if checkpoint.ExperimentId != experimentId {
				continue
			}
			checkpoints = append(checkpoints, &checkpoint)
		}
		return nil
	})
	return checkpoints, err
}

func (e *EphemeralStorage) writeBytes(obj interface{}, id string, bucket []byte, mutation *ChangeEvent) error {
	var encodedBytes []byte
	encoder := codec.NewEncoderBytes(&encodedBytes, new(codec.MsgpackHandle))
	if err := encoder.Encode(obj); err != nil {
		return err
	}
	err := e.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(bucket)
		if err != nil {
			return err
		}
		bucket.Put([]byte(id), encodedBytes)
		b, err := tx.CreateBucketIfNotExists([]byte(MUTATIONS))
		if err != nil {
			return err
		}
		if mutation != nil {
			mutationId, err := b.NextSequence()
			if err != nil {
				return err
			}
			val, err := mutation.json()
			if err != nil {
				return fmt.Errorf("unable to convert change event to json: %v", err)
			}
			b.Put(itob(int(mutationId)), val)
		}
		return nil
	})
	return err
}

func (e *EphemeralStorage) GetCheckpoint(_ context.Context, id string) (*Checkpoint, error) {
	var checkpoint Checkpoint
	err := e.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(CHECKPOINTS)
		cp := b.Get([]byte(id))
		if cp == nil {
			return fmt.Errorf("not found")
		}
		decoder := codec.NewDecoderBytes(cp, new(codec.MsgpackHandle))
		return decoder.Decode(&checkpoint)
	})
	if err != nil {
		return nil, err
	}
	return &checkpoint, nil
}

func (e *EphemeralStorage) CreateModel(_ context.Context, model *Model) (*CreateModelResult, error) {
	return &CreateModelResult{ModelId: model.Id}, e.writeBytes(model, model.Id, MODELS, nil)
}

func (e *EphemeralStorage) GetModel(_ context.Context, id string) (*Model, error) {
	var model Model
	err := e.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(MODELS)
		exp := b.Get([]byte(id))
		decoder := codec.NewDecoderBytes(exp, new(codec.MsgpackHandle))
		return decoder.Decode(&model)
	})
	return &model, err
}

func (e *EphemeralStorage) ListModelVersions(_ context.Context, modelId string) ([]*ModelVersion, error) {
	return nil, nil
}

func (e *EphemeralStorage) ListModels(_ context.Context, namespace string) ([]*Model, error) {
	models := make([]*Model, 0)
	err := e.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(MODELS)
		c := b.Cursor()
		handle := new(codec.MsgpackHandle)
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var model Model
			decoder := codec.NewDecoderBytes(v, handle)
			if err := decoder.Decode(&model); err != nil {
				return err
			}
			if model.Namespace == namespace {
				models = append(models, &model)
			}
		}
		return nil
	})
	return models, err
}

func (e *EphemeralStorage) CreateModelVersion(_ context.Context, modelVersion *ModelVersion) (*CreateModelVersionResult, error) {
	return &CreateModelVersionResult{ModelVersionId: modelVersion.Id}, e.writeBytes(modelVersion, modelVersion.Id, MODEL_VERSIONS, nil)
}

func (e *EphemeralStorage) GetModelVersion(ctx context.Context, id string) (*ModelVersion, error) {
	var modelVersion ModelVersion
	err := e.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(MODEL_VERSIONS)
		exp := b.Get([]byte(id))
		decoder := codec.NewDecoderBytes(exp, new(codec.MsgpackHandle))
		return decoder.Decode(&modelVersion)
	})
	return &modelVersion, err
}

func (e *EphemeralStorage) WriteFiles(_ context.Context, files FileSet) error {
	for _, file := range files {
		if err := e.writeBytes(file, file.Id, FILES, nil); err != nil {
			return err
		}
	}
	return nil
}

func (e *EphemeralStorage) GetFiles(ctx context.Context, parentId string) (FileSet, error) {
	files := []*FileMetadata{}
	err := e.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(FILES)
		c := b.Cursor()
		handle := new(codec.MsgpackHandle)
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var file FileMetadata
			decoder := codec.NewDecoderBytes(v, handle)
			if err := decoder.Decode(&file); err != nil {
				return err
			}
			if file.ParentId == parentId {
				files = append(files, &file)
			}
		}
		return nil
	})
	return files, err
}

func (e *EphemeralStorage) Ping() error {
	return nil
}

func (e *EphemeralStorage) CreateSchema(path string) error {
	return nil
}

func (e *EphemeralStorage) Backend() *BackendInfo {
	return &BackendInfo{Name: "boltdb"}
}

func (e *EphemeralStorage) UpdateBlobPath(_ context.Context, path string, parentId string, t FileMIMEType) error {
	return nil
}

func (e *EphemeralStorage) DeleteExperiment(_ context.Context, id string) error {
	return nil
}

func (e *EphemeralStorage) GetFile(ctx context.Context, id string) (*FileMetadata, error) {
	file := FileMetadata{}
	err := e.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(FILES)
		out := b.Get([]byte(id))
		if out == nil {
			return fmt.Errorf("blob with id: %v not present", id)
		}
		decoder := codec.NewDecoderBytes(out, new(codec.MsgpackHandle))
		return decoder.Decode(&file)
	})
	return &file, err
}

func (e *EphemeralStorage) UpdateMetadata(_ context.Context, metadataList []*Metadata) error {
	for _, m := range metadataList {
		if err := e.writeBytes(m, m.Id, METADATA, nil); err != nil {
			return err
		}
	}
	return nil
}

func (e *EphemeralStorage) ListMetadata(ctx context.Context, parentId string) ([]*Metadata, error) {
	metadataList := []*Metadata{}
	err := e.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(METADATA)
		c := b.Cursor()
		handle := new(codec.MsgpackHandle)
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var metadata Metadata
			decoder := codec.NewDecoderBytes(v, handle)
			if err := decoder.Decode(&metadata); err != nil {
				return err
			}
			if metadata.ParentId == parentId {
				metadataList = append(metadataList, &metadata)
			}
		}
		return nil
	})
	return metadataList, err
}

func (e *EphemeralStorage) ListChanges(ctx context.Context, namespace string, since time.Time) ([]*ChangeEvent, error) {
	events := []*ChangeEvent{}
	err := e.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(MUTATIONS)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var changeEvent ChangeEvent
			if err := json.Unmarshal(v, &changeEvent); err != nil {
				return fmt.Errorf("unable to decode change event: %v", err)
			}
			if changeEvent.Namespace == namespace && changeEvent.Time.After(since) {
				events = append(events, &changeEvent)
			}
		}
		return nil
	})
	return events, err
}
