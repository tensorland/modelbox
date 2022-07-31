package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	MODEL_NAME = "gpt3"
	OWNER      = "unicorn@modelbox.io"
	TASK       = "translate"
	NAMESPACE  = "ai/langtech/translation"
)

type StorageInterfaceTestSuite struct {
	t *testing.T

	storageIf MetadataStorage
}

func (s *StorageInterfaceTestSuite) TestCreateExperiment() {
	meta := SerializableMeta(map[string]string{"foo": "bar"})
	e := NewExperiment(MODEL_NAME, OWNER, NAMESPACE, "xyz", Pytorch, meta)
	_, err := s.storageIf.CreateExperiment(context.Background(), e)
	assert.Nil(s.t, err)
	experiments, err := s.storageIf.ListExperiments(context.Background(), e.Namespace)
	assert.Nil(s.t, err)
	assert.Equal(s.t, 1, len(experiments))
	assert.Equal(s.t, MODEL_NAME, experiments[0].Name)
	assert.Equal(s.t, OWNER, experiments[0].Owner)
	assert.Equal(s.t, NAMESPACE, experiments[0].Namespace)
	assert.Equal(s.t, "xyz", experiments[0].ExternalId)
}

func (s *StorageInterfaceTestSuite) TestCreateCheckpoint() {
	meta := SerializableMeta(map[string]string{"foo": "bar"})
	metrics := SerializableMetrics(map[string]float32{"val_loss": 0.041, "train_accu": 98.01})
	e := NewExperiment("quartznet-lid", "owner@email", "langtech", "xyz", Pytorch, meta)
	c := NewCheckpoint(e.Id, 45, meta, metrics)
	chk, err := s.storageIf.CreateCheckpoint(context.Background(), c)
	assert.Nil(s.t, err)
	assert.Equal(s.t, c.Id, chk.CheckpointId)
	checkpoints, err := s.storageIf.ListCheckpoints(context.Background(), e.Id)
	assert.Nil(s.t, err)
	assert.Equal(s.t, 1, len(checkpoints))
	assert.Equal(s.t, chk.CheckpointId, checkpoints[0].Id)
	assert.Equal(s.t, e.Id, checkpoints[0].ExperimentId)
	assert.Equal(s.t, uint64(45), checkpoints[0].Epoch)
	assert.Equal(s.t, meta, checkpoints[0].Meta)
	assert.Equal(s.t, metrics, checkpoints[0].Metrics)
}

func (s *StorageInterfaceTestSuite) TestObjectCreateIdempotency() {
	ctx := context.Background()
	meta := SerializableMeta(map[string]string{"foo": "bar"})
	metrics := SerializableMetrics(map[string]float32{"val_loss": 0.041, "train_accu": 98.01})
	e := NewExperiment("quartznet-lid", "owner@email", "langtech", "xyz", Pytorch, meta)
	result1, err := s.storageIf.CreateExperiment(ctx, e)
	assert.Nil(s.t, err)
	result2, err := s.storageIf.CreateExperiment(ctx, e)
	assert.Nil(s.t, err)
	assert.Equal(s.t, result1.ExperimentId, result2.ExperimentId)

	c := NewCheckpoint(e.Id, 45, meta, metrics)
	chk1, err := s.storageIf.CreateCheckpoint(ctx, c)
	assert.Nil(s.t, err)
	chk2, err := s.storageIf.CreateCheckpoint(ctx, c)
	assert.Nil(s.t, err)
	assert.Equal(s.t, chk1.CheckpointId, chk2.CheckpointId)

	m1 := NewModel(MODEL_NAME, e.Owner, NAMESPACE, TASK, "description", meta)
	resp1, err := s.storageIf.CreateModel(ctx, m1)
	assert.Nil(s.t, err)
	m2 := NewModel(MODEL_NAME, e.Owner, NAMESPACE, TASK, "description", meta)
	resp2, err := s.storageIf.CreateModel(ctx, m2)
	assert.Nil(s.t, err)
	assert.Equal(s.t, resp1.ModelId, resp2.ModelId)
}

func (s *StorageInterfaceTestSuite) TestCreateModel() {
	meta := map[string]string{"model": "gpt3"}
	description := "a large translation model based on gpt3"
	m := NewModel("blender", OWNER, NAMESPACE, TASK, description, meta)
	blob1 := NewBlobInfo(m.Id, "/foo/bar", "checksum123", File, 0, 0)
	blob2 := NewBlobInfo(m.Id, "/foo/pipe", "checksum345", ModelBlob, 0, 0)
	m.SetBlobs([]*BlobInfo{blob1, blob2})
	ctx := context.Background()
	_, err := s.storageIf.CreateModel(ctx, m)
	assert.Nil(s.t, err)

	m1, err := s.storageIf.GetModel(ctx, m.Id)
	assert.Nil(s.t, err)
	assert.Equal(s.t, description, m1.Description)
	assert.Equal(s.t, NAMESPACE, m1.Namespace)
}

func (s *StorageInterfaceTestSuite) TestListModels() {
	meta := map[string]string{"model": "gpt3"}
	description := "a large translation model based on gpt3"
	namespace := "namespace-x"
	m := NewModel("blender", OWNER, namespace, TASK, description, meta)
	blob1 := NewBlobInfo(m.Id, "/foo/bar", "checksum123", File, 0, 0)
	blob2 := NewBlobInfo(m.Id, "/foo/pipe", "checksum345", ModelBlob, 0, 0)
	m.SetBlobs([]*BlobInfo{blob1, blob2})
	ctx := context.Background()
	_, err := s.storageIf.CreateModel(ctx, m)
	assert.Nil(s.t, err)

	models, err := s.storageIf.ListModels(ctx, namespace)
	assert.Nil(s.t, err)
	assert.Equal(s.t, 1, len(models))
	assert.Equal(s.t, 2, len(models[0].Blobs))
}

func (s *StorageInterfaceTestSuite) TestCreateModelVersion() {
	meta := map[string]string{"bar": "foo"}
	blobs := []*BlobInfo{}
	mvName := "test-version"
	version := "1"
	description := "testing"
	uniqueTags := SerializableTags([]string{"foo", "bar"})
	mv := NewModelVersion(
		mvName,
		MODEL_NAME,
		version,
		description,
		Pytorch,
		meta,
		blobs,
		uniqueTags,
	)
	_, err := s.storageIf.CreateModelVersion(context.Background(), mv)
	assert.Nil(s.t, err)

	mv1, err := s.storageIf.GetModelVersion(context.Background(), mv.Id)
	assert.Nil(s.t, err)
	assert.Equal(s.t, mvName, mv1.Name)
	assert.Equal(s.t, version, mv1.Version)
	assert.Equal(s.t, description, mv1.Description)
	assert.Equal(s.t, uniqueTags, mv1.UniqueTags)
}

func (s *StorageInterfaceTestSuite) TestWriteBlobs() {
	ctx := context.Background()
	blob1 := NewBlobInfo(MODEL_NAME, "/foo/bar", "checksum123", File, 0, 0)
	blob2 := NewBlobInfo(MODEL_NAME, "/foo/pipe", "checksum345", ModelBlob, 0, 0)
	blobs := []*BlobInfo{blob1, blob2}
	err := s.storageIf.WriteBlobs(ctx, blobs)
	assert.Nil(s.t, err)

	// Test Get Blobs for ParentID
	blobsOut, err := s.storageIf.GetBlobs(ctx, MODEL_NAME)
	assert.Nil(s.t, err)
	assert.Equal(s.t, 2, len(blobsOut))
	assert.Equal(s.t, "/foo/bar", blobsOut[0].Path)
	assert.Equal(s.t, "/foo/pipe", blobsOut[1].Path)

	// Test Get Blob with ID
	blob3, err := s.storageIf.GetBlob(ctx, blob1.Id)
	assert.Nil(s.t, err)
	assert.Equal(s.t, blob1.Id, blob3.Id)
}

func (s *StorageInterfaceTestSuite) TestUpdateMetadata() {
	ctx := context.Background()

	// Write Metadata
	scalerVal, err := structpb.NewValue(map[string]interface{}{"/foo/bar": interface{}(1)})
	assert.Nil(s.t, err)
	m := NewMetadata("parent-id1", "/foo/bar", scalerVal)
	err = s.storageIf.UpdateMetadata(ctx, []*Metadata{m})
	assert.Nil(s.t, err)

	complexVal, err1 := structpb.NewValue(map[string]interface{}{"/tmp/hola": map[string]interface{}{"name1": "val1", "name2": 5}})
	assert.Nil(s.t, err1)
	m2 := NewMetadata("parent-id1", "/foo/bar1", complexVal)
	err = s.storageIf.UpdateMetadata(ctx, []*Metadata{m2})
	assert.Nil(s.t, err)

	// Get Metadata
	meta, err := s.storageIf.ListMetadata(ctx, "parent-id1")
	assert.Nil(s.t, err)
	assert.Equal(s.t, 2, len(meta))
}
