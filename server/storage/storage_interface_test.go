package storage

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tensorland/modelbox/server/storage/artifacts"
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
	since := time.Now()
	ctx := context.Background()
	e := NewExperiment(MODEL_NAME, OWNER, NAMESPACE, "xyz", Pytorch)
	metaVal, _ := structpb.NewValue(map[string]interface{}{"/foo": 5})
	metaData := map[string]*structpb.Value{"foo": metaVal}
	_, err := s.storageIf.CreateExperiment(context.Background(), e, metaData)
	assert.Nil(s.t, err)
	experiments, err := s.storageIf.ListExperiments(context.Background(), e.Namespace)
	assert.Nil(s.t, err)
	assert.Equal(s.t, 1, len(experiments))
	assert.Equal(s.t, MODEL_NAME, experiments[0].Name)
	assert.Equal(s.t, OWNER, experiments[0].Owner)
	assert.Equal(s.t, NAMESPACE, experiments[0].Namespace)
	assert.Equal(s.t, "xyz", experiments[0].ExternalId)

	// Check for mutation events
	changes, err := s.storageIf.ListChanges(ctx, NAMESPACE, since)
	assert.Nil(s.t, err)
	assert.Equal(s.t, len(changes), 1)
}

func (s *StorageInterfaceTestSuite) TestCreateCheckpoint() {
	metrics := SerializableMetrics(map[string]float32{"val_loss": 0.041, "train_accu": 98.01})
	e := NewExperiment("quartznet-lid", "owner@email", "langtech", "xyz", Pytorch)
	c := NewCheckpoint(e.Id, 45, metrics)
	chk, err := s.storageIf.CreateCheckpoint(context.Background(), c, nil)
	assert.Nil(s.t, err)
	assert.Equal(s.t, c.Id, chk.CheckpointId)
	checkpoints, err := s.storageIf.ListCheckpoints(context.Background(), e.Id)
	assert.Nil(s.t, err)
	assert.Equal(s.t, 1, len(checkpoints))
	assert.Equal(s.t, chk.CheckpointId, checkpoints[0].Id)
	assert.Equal(s.t, e.Id, checkpoints[0].ExperimentId)
	assert.Equal(s.t, uint64(45), checkpoints[0].Epoch)
	assert.Equal(s.t, metrics, checkpoints[0].Metrics)
}

func (s *StorageInterfaceTestSuite) TestObjectCreateIdempotency() {
	ctx := context.Background()
	metrics := SerializableMetrics(map[string]float32{"val_loss": 0.041, "train_accu": 98.01})
	e := NewExperiment("quartznet-lid", "owner@email", "langtech", "xyz", Pytorch)
	result1, err := s.storageIf.CreateExperiment(ctx, e, nil)
	assert.Nil(s.t, err)
	result2, err := s.storageIf.CreateExperiment(ctx, e, nil)
	assert.Nil(s.t, err)
	assert.Equal(s.t, result1.ExperimentId, result2.ExperimentId)

	c := NewCheckpoint(e.Id, 45, metrics)
	chk1, err := s.storageIf.CreateCheckpoint(ctx, c, nil)
	assert.Nil(s.t, err)
	chk2, err := s.storageIf.CreateCheckpoint(ctx, c, nil)
	assert.Nil(s.t, err)
	assert.Equal(s.t, chk1.CheckpointId, chk2.CheckpointId)

	m1 := NewModel(MODEL_NAME, e.Owner, NAMESPACE, TASK, "description")
	resp1, err := s.storageIf.CreateModel(ctx, m1, nil)
	assert.Nil(s.t, err)
	m2 := NewModel(MODEL_NAME, e.Owner, NAMESPACE, TASK, "description")
	resp2, err := s.storageIf.CreateModel(ctx, m2, nil)
	assert.Nil(s.t, err)
	assert.Equal(s.t, resp1.ModelId, resp2.ModelId)
}

func (s *StorageInterfaceTestSuite) TestCreateModel() {
	since := time.Now()
	description := "a large translation model based on gpt3"
	m := NewModel("blender", OWNER, NAMESPACE, TASK, description)
	blob1 := artifacts.NewFileMetadata(m.Id, "/foo/bar", "checksum123", artifacts.TextFile, 0, 0)
	blob2 := artifacts.NewFileMetadata(m.Id, "/foo/pipe", "checksum345", artifacts.ModelFile, 0, 0)
	m.SetFiles([]*artifacts.FileMetadata{blob1, blob2})
	metaVal, _ := structpb.NewValue(map[string]interface{}{"/foo": 5})
	metaData := map[string]*structpb.Value{"foo": metaVal}
	ctx := context.Background()
	_, err := s.storageIf.CreateModel(ctx, m, metaData)
	assert.Nil(s.t, err)

	m1, err := s.storageIf.GetModel(ctx, m.Id)
	assert.Nil(s.t, err)
	assert.Equal(s.t, description, m1.Description)
	assert.Equal(s.t, NAMESPACE, m1.Namespace)

	// Check for mutation events
	changes, err := s.storageIf.ListChanges(ctx, NAMESPACE, since)
	assert.Nil(s.t, err)
	// TODO Make this better by being able to filter events only for a specific object
	hasEvent := false
	for _, change := range changes {
		if change.ObjectId == m.Id {
			hasEvent = true
		}
	}
	assert.True(s.t, hasEvent)
}

func (s *StorageInterfaceTestSuite) TestListModels() {
	description := "a large translation model based on gpt3"
	namespace := "namespace-x"
	m := NewModel("blender", OWNER, namespace, TASK, description)
	blob1 := artifacts.NewFileMetadata(m.Id, "/foo/bar", "checksum123", artifacts.TextFile, 0, 0)
	blob2 := artifacts.NewFileMetadata(m.Id, "/foo/pipe", "checksum345", artifacts.ModelFile, 0, 0)
	m.SetFiles([]*artifacts.FileMetadata{blob1, blob2})
	ctx := context.Background()
	_, err := s.storageIf.CreateModel(ctx, m, nil)
	assert.Nil(s.t, err)

	models, err := s.storageIf.ListModels(ctx, namespace)
	assert.Nil(s.t, err)
	assert.Equal(s.t, 1, len(models))
	assert.Equal(s.t, 2, len(models[0].Files))
}

func (s *StorageInterfaceTestSuite) TestCreateModelVersion() {
	blobs := []*artifacts.FileMetadata{}
	mvName := "test-version"
	version := "1"
	description := "testing"
	uniqueTags := SerializableTags([]string{"foo", "bar"})
	metaVal, _ := structpb.NewValue(map[string]interface{}{"/foo": 5})
	metaData := map[string]*structpb.Value{"foo": metaVal}
	mv := NewModelVersion(
		mvName,
		MODEL_NAME,
		version,
		description,
		Pytorch,
		blobs,
		uniqueTags,
	)
	_, err := s.storageIf.CreateModelVersion(context.Background(), mv, metaData)
	assert.Nil(s.t, err)

	mv1, err := s.storageIf.GetModelVersion(context.Background(), mv.Id)
	assert.Nil(s.t, err)
	assert.Equal(s.t, mvName, mv1.Name)
	assert.Equal(s.t, version, mv1.Version)
	assert.Equal(s.t, description, mv1.Description)
	assert.Equal(s.t, uniqueTags, mv1.UniqueTags)
}

func (s *StorageInterfaceTestSuite) TestListModelVersions() {
	modelId := "some-reandom-model"
	description := "testing"
	uniqueTags := SerializableTags([]string{"foo", "bar"})
	metaVal, _ := structpb.NewValue(map[string]interface{}{"/foo": 5})
	metaData := map[string]*structpb.Value{"foo": metaVal}
	mv := NewModelVersion(
		"test-version1",
		modelId,
		"1",
		description,
		Pytorch,
		[]*artifacts.FileMetadata{},
		uniqueTags,
	)
	_, err := s.storageIf.CreateModelVersion(context.Background(), mv, metaData)
	assert.Nil(s.t, err)
	mv1 := NewModelVersion(
		"test-version-2",
		modelId,
		"2",
		description,
		Pytorch,
		[]*artifacts.FileMetadata{},
		uniqueTags,
	)
	_, err = s.storageIf.CreateModelVersion(context.Background(), mv1, metaData)
	assert.Nil(s.t, err)

	mvs, err := s.storageIf.ListModelVersions(context.Background(), modelId)
	assert.Nil(s.t, err)
	assert.Equal(s.t, 2, len(mvs))
}

func (s *StorageInterfaceTestSuite) TestWriteBlobs() {
	ctx := context.Background()
	blob1 := artifacts.NewFileMetadata(MODEL_NAME, "/foo/bar", "checksum123", artifacts.TextFile, 0, 0)
	blob2 := artifacts.NewFileMetadata(MODEL_NAME, "/foo/pipe", "checksum345", artifacts.ModelFile, 0, 0)
	blobs := []*artifacts.FileMetadata{blob1, blob2}
	err := s.storageIf.WriteFiles(ctx, blobs)
	assert.Nil(s.t, err)

	// Test Get Blobs for ParentID
	blobsOut, err := s.storageIf.GetFiles(ctx, MODEL_NAME)
	assert.Nil(s.t, err)
	assert.Equal(s.t, 2, len(blobsOut))
	assert.Equal(s.t, "/foo/bar", blobsOut[0].Path)
	assert.Equal(s.t, "/foo/pipe", blobsOut[1].Path)

	// Test Get Blob with ID
	blob3, err := s.storageIf.GetFile(ctx, blob1.Id)
	assert.Nil(s.t, err)
	assert.Equal(s.t, blob1.Id, blob3.Id)
}

func (s *StorageInterfaceTestSuite) TestUpdateMetadata() {
	ctx := context.Background()

	// Write Metadata
	val1, _ := structpb.NewValue(1)
	meta1 := map[string]*structpb.Value{"/tmp/foo": val1}
	err := s.storageIf.UpdateMetadata(ctx, "parent-id1", meta1)
	assert.Nil(s.t, err)

	val2, _ := structpb.NewValue(map[string]interface{}{"name1": "val1", "name2": 5})
	complexVal := map[string]*structpb.Value{"/tmp/hola": val2}
	err = s.storageIf.UpdateMetadata(ctx, "parent-id2", complexVal)
	assert.Nil(s.t, err)

	// Get Metadata
	meta3, err := s.storageIf.ListMetadata(ctx, "parent-id1")
	assert.Nil(s.t, err)
	assert.Equal(s.t, 1, len(meta3))

	meta4, err := s.storageIf.ListMetadata(ctx, "parent-id2")
	assert.Nil(s.t, err)
	assert.Equal(s.t, 1, len(meta4))
}

func (s *StorageInterfaceTestSuite) TestCreateEvent() {
	val, _ := structpb.NewValue(map[string]interface{}{"name1": "val1", "name2": 5})
	meta := map[string]*structpb.Value{"/tmp/hola": val}
	eventName := "write-chk-step1"
	parentId := "parent1"
	event := NewEvent(parentId, "trainer-1", eventName, time.Now(), meta)
	err := s.storageIf.LogEvent(context.Background(), parentId, event)
	assert.Nil(s.t, err)

	events, err := s.storageIf.ListEvents(context.Background(), parentId)
	assert.Nil(s.t, err)
	assert.Equal(s.t, 1, len(events))
}

func (s *StorageInterfaceTestSuite) TestGetExperiment() {
	ctx := context.Background()
	e := NewExperiment(MODEL_NAME, OWNER, NAMESPACE, "xyz", Pytorch)
	metaVal, _ := structpb.NewValue(map[string]interface{}{"/foo": 5})
	metaData := map[string]*structpb.Value{"foo": metaVal}
	_, err := s.storageIf.CreateExperiment(context.Background(), e, metaData)
	assert.Nil(s.t, err)
	experiment, err := s.storageIf.GetExperiment(ctx, e.Id)
	assert.Nil(s.t, err)
	assert.Equal(s.t, MODEL_NAME, experiment.Name)
	assert.Equal(s.t, OWNER, experiment.Owner)
	assert.Equal(s.t, NAMESPACE, experiment.Namespace)
	assert.Equal(s.t, "xyz", experiment.ExternalId)
}

func (s *StorageInterfaceTestSuite) TestCreateActions() {
	ctx := context.Background()
	a1 := NewAction("quantize", "x86", "parent1", "trigger", s.createMetadata())
	err := s.storageIf.CreateAction(ctx, a1)
	assert.Nil(s.t, err)

	// Ensure that the action is available
	actions, err := s.storageIf.ListActions(ctx, "parent1")
	assert.Nil(s.t, err)
	assert.Equal(s.t, 1, len(actions))

	// Get Action
	actionState, err := s.storageIf.GetAction(ctx, a1.Id)
	assert.Nil(s.t, err)
	assert.Equal(s.t, actionState.Action.Id, a1.Id)
	assert.Equal(s.t, actionState.Action.Arch, a1.Arch)

	// Get Action Eval
	actionEvals, err := s.filterEvalByActionId(a1.Id)
	assert.Nil(s.t, err)
	require.Equal(s.t, 1, len(actionEvals))
	// Ensure eval matches expectaction
	eval := actionEvals[0]
	assert.NotEmpty(s.t, eval.Id)
	assert.Equal(s.t, eval.ParentId, a1.Id)
	assert.Equal(s.t, eval.ParentType, EvalParentAction)
	assert.Equal(s.t, eval.Type, EvalTypeActionCreated)
	assert.Equal(s.t, eval.ProcessedAt, int64(0))
}

func (s *StorageInterfaceTestSuite) TestCreateActionInstance() {
	ctx := context.Background()
	a1 := NewAction("quantize", "x86", "parent2", "trigger", s.createMetadata())
	err := s.storageIf.CreateAction(ctx, a1)
	assert.Nil(s.t, err)
	actionEvals, err := s.filterEvalByActionId(a1.Id)
	assert.Nil(s.t, err)

	instance := NewActionInstance(a1.Id, 0)
	err = s.storageIf.CreateActionInstance(ctx, instance, actionEvals[0])
	assert.Nil(s.t, err)

	actionState, err := s.storageIf.GetAction(ctx, a1.Id)
	assert.Nil(s.t, err)
	assert.Equal(s.t, 1, len(actionState.Instances))
	assert.Equal(s.t, instance, actionState.Instances[0])

	// Ensure that the eval is marked as processed
	actionEval, err := s.storageIf.GetActionEvalById(ctx, actionEvals[0].Id)
	require.Nil(s.t, err)
	assert.NotEmpty(s.t, actionEval.ProcessedAt)
}

func (s *StorageInterfaceTestSuite) TestUpdateActionInstanceSuccess() {
	// 1. Create a action and action instance
	// 2. Update the action instance to finished
	// 3. Assert that the action was updated
	ctx := context.Background()
	a1 := NewAction("quantize", "x86", "parent3", "trigger", s.createMetadata())
	err := s.storageIf.CreateAction(ctx, a1)
	require.Nil(s.t, err)
	actionEvals, err := s.filterEvalByActionId(a1.Id)
	require.Nil(s.t, err)
	ai1 := NewActionInstance(a1.Id, 0)
	err = s.storageIf.CreateActionInstance(ctx, ai1, actionEvals[0])
	require.Nil(s.t, err)

	finishedTime := uint64(ai1.CreatedAt) + uint64(2*time.Minute)
	_, eval1 := ai1.Update(NewActionInstanceUpdate(ai1.Id, StatusFinished, OutcomeSuccess, "", finishedTime))
	err = s.storageIf.UpdateActionInstance(ctx, ai1, eval1)
	require.Nil(s.t, err)

	ai2, err := s.storageIf.GetActionInstance(ctx, ai1.Id)
	require.Nil(s.t, err)
	assert.Equal(s.t, OutcomeSuccess, ai2.Outcome)
	assert.Equal(s.t, StatusFinished, ai2.Status)
	assert.Equal(s.t, finishedTime, uint64(ai2.FinishedAt))
}

func (s *StorageInterfaceTestSuite) TestUpdateActionInstanceFail() {
	// 1. Create a action and action instance
	// 2. Update the action instance to finished
	// 3. Assert that the action was updated
	// 4. Assert that the eval is processed.
	ctx := context.Background()
	a1 := NewAction("quantize", "x86", "parent4", "trigger", s.createMetadata())
	err := s.storageIf.CreateAction(ctx, a1)
	require.Nil(s.t, err)
	actionEvals, err := s.filterEvalByActionId(a1.Id)
	require.Nil(s.t, err)
	ai1 := NewActionInstance(a1.Id, 0)
	err = s.storageIf.CreateActionInstance(ctx, ai1, actionEvals[0])
	require.Nil(s.t, err)

	finishedTime := uint64(ai1.CreatedAt) + uint64(2*time.Minute)
	_, eval1 := ai1.Update(NewActionInstanceUpdate(ai1.Id, StatusFinished, OutcomeFailure, "", finishedTime))
	assert.NotNil(s.t, eval1)
	err = s.storageIf.UpdateActionInstance(ctx, ai1, eval1)
	require.Nil(s.t, err)

	ai2, err := s.storageIf.GetActionInstance(ctx, ai1.Id)
	require.Nil(s.t, err)
	assert.Equal(s.t, OutcomeFailure, ai2.Outcome)
	assert.Equal(s.t, StatusFinished, ai2.Status)
	assert.Equal(s.t, finishedTime, uint64(ai2.FinishedAt))
	eval2, err := s.storageIf.GetActionEvalById(ctx, eval1.Id)
	require.Nil(s.t, err)
	assert.NotEmpty(s.t, eval2.ProcessedAt)
}

func (s *StorageInterfaceTestSuite) filterEvalByActionId(id string) ([]*ActionEval, error) {
	var evals []*ActionEval
	actionEvals, err := s.storageIf.GetActionEvals(context.Background())
	if err != nil {
		return nil, err
	}
	for _, ev := range actionEvals {
		if ev.ParentId == id {
			evals = append(evals, ev)
		}
	}
	return evals, nil
}

func (s *StorageInterfaceTestSuite) createMetadata() map[string]*structpb.Value {
	metaVal, _ := structpb.NewValue(map[string]interface{}{"/foo": 5})
	return map[string]*structpb.Value{"foo": metaVal}
}
