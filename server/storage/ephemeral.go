package storage

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fatih/structs"
	"github.com/tensorland/modelbox/server/storage/artifacts"
	"github.com/vmihailenco/msgpack/v5"
	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	EXPERIMENTS    = []byte("experiments")
	CHECKPOINTS    = []byte("checkpoints")
	MODELS         = []byte("models")
	MODEL_VERSIONS = []byte("model_versions")
	FILES          = []byte("files")
	METADATA       = []byte("metadata")
	MUTATIONS      = []byte("mutations")
	EVENTS         = []byte("events")
	ACTIONS        = []byte("actions")

	ACTION_PARENT_IDX = []byte("action_parent_idx")
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
	db, err := bolt.Open(path, 0o666, nil)
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
		if _, err := tx.CreateBucketIfNotExists(EVENTS); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(MUTATIONS); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(METADATA); err != nil {
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

func (e *EphemeralStorage) CreateExperiment(_ context.Context, experiment *Experiment, metadata SerializableMetadata) (*CreateExperimentResult, error) {
	id := experiment.Hash()
	event := &ChangeEvent{
		ObjectId:   experiment.Id,
		ObjectType: "experiment",
		Action:     "create",
		Time:       time.Now(),
		Payload:    structs.Map(experiment),
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
		if exp == nil {
			return fmt.Errorf("unable to find experiment with id: %v", id)
		}
		return msgpack.Unmarshal(exp, &experiment)
	})
	return &experiment, err
}

func (e *EphemeralStorage) ListExperiments(_ context.Context, namespace string) ([]*Experiment, error) {
	experiments := make([]*Experiment, 0)
	err := e.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(EXPERIMENTS)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var experiment Experiment
			if err := msgpack.Unmarshal(v, &experiment); err != nil {
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

func (e *EphemeralStorage) CreateCheckpoint(ctx context.Context, checkpoint *Checkpoint, metadata SerializableMetadata) (*CreateCheckpointResult, error) {
	if err := e.writeBytes(checkpoint, checkpoint.Id, CHECKPOINTS, nil); err != nil {
		return nil, err
	}

	if err := e.UpdateMetadata(ctx, checkpoint.Id, metadata); err != nil {
		return nil, err
	}

	return &CreateCheckpointResult{CheckpointId: checkpoint.Id}, nil
}

func (e *EphemeralStorage) ListCheckpoints(_ context.Context, experimentId string) ([]*Checkpoint, error) {
	checkpoints := make([]*Checkpoint, 0)
	err := e.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(CHECKPOINTS)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var checkpoint Checkpoint
			if err := msgpack.Unmarshal(v, &checkpoint); err != nil {
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
	encodedBytes, err := msgpack.Marshal(obj)
	if err != nil {
		return fmt.Errorf("unable to encode object to msgpack: %v", err)
	}
	return e.db.Update(func(tx *bolt.Tx) error {
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
}

func (e *EphemeralStorage) writeJsonBytes(obj interface{}, id string, bucket []byte, mutation *ChangeEvent) error {
	encodedBytes, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("unable to encode object in json: %v", err)
	}
	return e.db.Update(func(tx *bolt.Tx) error {
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
}

func (e *EphemeralStorage) GetCheckpoint(_ context.Context, id string) (*Checkpoint, error) {
	var checkpoint Checkpoint
	err := e.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(CHECKPOINTS)
		cp := b.Get([]byte(id))
		if cp == nil {
			return fmt.Errorf("not found")
		}
		return msgpack.Unmarshal(cp, &checkpoint)
	})
	if err != nil {
		return nil, err
	}
	return &checkpoint, nil
}

func (e *EphemeralStorage) CreateModel(ctx context.Context, model *Model, metadata SerializableMetadata) (*CreateModelResult, error) {
	event := &ChangeEvent{
		ObjectId:   model.Id,
		ObjectType: "model",
		Action:     "create",
		Time:       time.Now(),
		Payload:    structs.Map(model),
		Namespace:  model.Namespace,
	}

	result, err := &CreateModelResult{ModelId: model.Id}, e.writeBytes(model, model.Id, MODELS, event)
	if err != nil {
		return nil, err
	}

	if err := e.UpdateMetadata(ctx, model.Id, metadata); err != nil {
		return nil, err
	}
	return result, nil
}

func (e *EphemeralStorage) GetModel(_ context.Context, id string) (*Model, error) {
	var model Model
	err := e.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(MODELS)
		exp := b.Get([]byte(id))
		if exp == nil {
			return fmt.Errorf("no model found with key %v", id)
		}
		return msgpack.Unmarshal(exp, &model)
	})
	return &model, err
}

func (e *EphemeralStorage) ListModels(_ context.Context, namespace string) ([]*Model, error) {
	models := make([]*Model, 0)
	err := e.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(MODELS)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var model Model
			if err := msgpack.Unmarshal(v, &model); err != nil {
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

func (e *EphemeralStorage) ListModelVersions(_ context.Context, modelId string) ([]*ModelVersion, error) {
	modelVersions := []*ModelVersion{}
	err := e.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(MODEL_VERSIONS)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var mv ModelVersion
			if err := msgpack.Unmarshal(v, &mv); err != nil {
				return err
			}
			if mv.ModelId == modelId {
				modelVersions = append(modelVersions, &mv)
			}
		}
		return nil
	})
	return modelVersions, err
}

func (e *EphemeralStorage) CreateModelVersion(ctx context.Context, modelVersion *ModelVersion, metadata SerializableMetadata) (*CreateModelVersionResult, error) {
	result, err := &CreateModelVersionResult{ModelVersionId: modelVersion.Id}, e.writeBytes(modelVersion, modelVersion.Id, MODEL_VERSIONS, nil)
	if err != nil {
		return nil, err
	}
	if err := e.UpdateMetadata(ctx, modelVersion.Id, metadata); err != nil {
		return nil, err
	}
	return result, nil
}

func (e *EphemeralStorage) GetModelVersion(ctx context.Context, id string) (*ModelVersion, error) {
	var modelVersion ModelVersion
	err := e.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(MODEL_VERSIONS)
		exp := b.Get([]byte(id))
		if exp == nil {
			return fmt.Errorf("unable to find model version with id: %v", id)
		}
		return msgpack.Unmarshal(exp, &modelVersion)
	})
	return &modelVersion, err
}

func (e *EphemeralStorage) WriteFiles(_ context.Context, files artifacts.FileSet) error {
	for _, file := range files {
		if err := e.writeBytes(file, file.Id, FILES, nil); err != nil {
			return err
		}
	}
	return nil
}

func (e *EphemeralStorage) GetFiles(ctx context.Context, parentId string) (artifacts.FileSet, error) {
	files := []*artifacts.FileMetadata{}
	err := e.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(FILES)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var file artifacts.FileMetadata
			if err := msgpack.Unmarshal(v, &file); err != nil {
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

func (e *EphemeralStorage) GetFile(ctx context.Context, id string) (*artifacts.FileMetadata, error) {
	file := artifacts.FileMetadata{}
	err := e.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(FILES)
		out := b.Get([]byte(id))
		if out == nil {
			return fmt.Errorf("blob with id: %v not present", id)
		}
		return msgpack.Unmarshal(out, &file)
	})
	return &file, err
}

func (e *EphemeralStorage) UpdateMetadata(_ context.Context, parentId string, metadata map[string]*structpb.Value) error {
	for k, v := range metadata {
		key := fmt.Sprintf("%s-%s", parentId, k)
		if err := e.writeJsonBytes(map[string]*structpb.Value{k: v}, key, METADATA, nil); err != nil {
			return err
		}
	}
	return nil
}

func (e *EphemeralStorage) ListMetadata(ctx context.Context, parentId string) (map[string]*structpb.Value, error) {
	metadata := map[string]*structpb.Value{}
	err := e.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(METADATA).Cursor()
		parentIdBytes := []byte(parentId)
		for k, v := c.Seek(parentIdBytes); k != nil && bytes.HasPrefix(k, parentIdBytes); k, v = c.Next() {
			var mB map[string]interface{}
			if err := json.Unmarshal(v, &mB); err != nil {
				return fmt.Errorf("can't unmarshal json: %v", err)
			}
			for mK, mV := range mB {
				m, err := structpb.NewValue(mV)
				if err != nil {
					return fmt.Errorf("cant create value: %v", err)
				}
				metadata[mK] = m
			}
		}
		return nil
	})
	return metadata, err
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

func (e *EphemeralStorage) LogEvent(ctx context.Context, parentId string, event *Event) error {
	id := fmt.Sprintf("%s-%s", parentId, event.Id)
	encodedBytes, err := msgpack.Marshal(event)
	if err != nil {
		return fmt.Errorf("unable to encode object to msgpack: %v", err)
	}
	return e.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(EVENTS)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(id), encodedBytes)
	})
}

func (e *EphemeralStorage) ListEvents(ctx context.Context, parentId string) ([]*Event, error) {
	events := []*Event{}
	err := e.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(EVENTS).Cursor()
		byteId := []byte(parentId)
		for k, v := c.First(); k != nil && bytes.HasPrefix(k, byteId); k, v = c.Next() {
			var event Event
			if err := json.Unmarshal(v, &event); err != nil {
				return fmt.Errorf("unable to decode event from json : %v", err)
			}
			events = append(events, &event)
		}
		return nil
	})
	return events, err
}

func (e *EphemeralStorage) ListActions(ctx context.Context, parentId string) ([]*Action, error) {
	actions := []*Action{}
	err := e.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(ACTION_PARENT_IDX).Cursor()
		actionBucket := tx.Bucket(ACTIONS)
		parentBytes := []byte(parentId)
		for k, v := c.First(); k != nil && bytes.HasPrefix(k, parentBytes); k, v = c.Next() {
			var actionId string
			if err := json.Unmarshal(v, &actionId); err != nil {
				return fmt.Errorf("unable to decode action id from json : %v", err)
			}

			actionBytes := actionBucket.Get([]byte(actionId))
			if actionBytes == nil {
				continue
			}
			var action Action
			if err := json.Unmarshal(actionBytes, &action); err != nil {
				return fmt.Errorf("unable to decode action from bytes: %v", err)
			}
			actions = append(actions, &action)
		}
		return nil
	})
	return actions, err
}

func (e *EphemeralStorage) GetAction(ctx context.Context, id string) (*ActionState, error) {
	var state ActionState

	err := e.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(ACTIONS)
		actionBytes := b.Get([]byte(id))
		if actionBytes == nil {
			return fmt.Errorf("unknown action")
		}

		var action Action

		if err := json.Unmarshal(actionBytes, &action); err != nil {
			return err
		}
		state.Action = &action

		// TODO Get action instances

		return nil
	})
	if err != nil {
		return nil, err
	}
	return &state, nil
}

func (e *EphemeralStorage) CreateAction(ctx context.Context, action *Action) error {
	// TODO Create a changeEvent
	// TODO Put the two writes under one tx
	if err := e.writeJsonBytes(action.Id, fmt.Sprintf("%s-%s", action.ParentId, action.Id), ACTION_PARENT_IDX, nil); err != nil {
		return err
	}
	return e.writeJsonBytes(action, action.Id, ACTIONS, nil)
}
