package storage

import (
	"crypto/sha1"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx/types"
	"github.com/tensorland/modelbox/server/storage/artifacts"
	"github.com/tensorland/modelbox/server/utils"
	"google.golang.org/protobuf/types/known/structpb"
)

type ModelSchema struct {
	Id        string
	Name      string
	Owner     string
	Namespace string
	Task      string
	Desc      string `db:"description"`
	CreatedAt int64  `db:"created_at"`
	UpdatedAt int64  `db:"updated_at"`
}

func (m *ModelSchema) ToModel(files artifacts.FileSet) *Model {
	model := &Model{
		Id:          m.Id,
		Name:        m.Name,
		Owner:       m.Owner,
		Namespace:   m.Namespace,
		Task:        m.Task,
		Description: m.Desc,
		Files:       []*artifacts.FileMetadata{},
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
	if files != nil {
		model.Files = files
	}
	return model
}

func (m *ModelSchema) mutationSchema(model *Model) *MutationEventSchema {
	return &MutationEventSchema{
		MutationTime: uint64(time.Now().Unix()),
		EventType:    uint8(EventTypeModelCreated),
		ObjectType:   uint8(EventObjectTypeModel),
		ObjectId:     m.Id,
		Namespace:    m.Namespace,
		ModelPayload: model,
	}
}

func ModelToSchema(m *Model) *ModelSchema {
	return &ModelSchema{
		Id:        m.Id,
		Name:      m.Name,
		Owner:     m.Owner,
		Namespace: m.Namespace,
		Task:      m.Task,
		Desc:      m.Description,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

type ModelVersionSchema struct {
	Id         string
	Name       string
	Model      string `db:"model_id"`
	Version    string
	Desc       string           `db:"description"`
	Framework  int8             `db:"ml_framework"`
	UniqueTags SerializableTags `db:"unique_tags"`
	Tags       SerializableTags `db:"tags"`
	CreatedAt  int64            `db:"created_at"`
	UpdatedAt  int64            `db:"updated_at"`
}

func (m *ModelVersionSchema) ToModelVersion(files artifacts.FileSet) *ModelVersion {
	modelVersion := &ModelVersion{
		Id:          m.Id,
		Name:        m.Name,
		ModelId:     m.Model,
		Version:     m.Version,
		Description: m.Desc,
		Framework:   MLFramework(m.Framework),
		Files:       files,
		UniqueTags:  m.UniqueTags,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
	return modelVersion
}

func ModelVersionToSchema(mv *ModelVersion) *ModelVersionSchema {
	return &ModelVersionSchema{
		Id:         mv.Id,
		Name:       mv.Name,
		Model:      mv.ModelId,
		Version:    mv.Version,
		Desc:       mv.Description,
		Framework:  int8(mv.Framework),
		UniqueTags: mv.UniqueTags,
		CreatedAt:  mv.CreatedAt,
		UpdatedAt:  mv.UpdatedAt,
	}
}

func (m *ModelVersionSchema) mutationSchema(mv *ModelVersion) *MutationEventSchema {
	return &MutationEventSchema{
		MutationTime:        uint64(time.Now().Unix()),
		EventType:           uint8(EventTypeModelVersionCreated),
		ObjectType:          uint8(EventObjectTypeModelVersion),
		ObjectId:            m.Id,
		Namespace:           "",
		ModelVersionPayload: mv,
	}
}

func ToFileSet(rows []FileSchema) ([]*artifacts.FileMetadata, error) {
	fileSet := make([]*artifacts.FileMetadata, len(rows))
	for i, row := range rows {
		file, err := row.ToFile()
		if err != nil {
			return nil, err
		}
		fileSet[i] = file
	}
	return fileSet, nil
}

type FileSchema struct {
	Id       string
	ParentId string         `db:"parent_id"`
	Meta     types.JSONText `db:"metadata"`
}

func (b *FileSchema) ToFile() (*artifacts.FileMetadata, error) {
	type BlobMeta struct {
		Type      artifacts.FileMIMEType
		Path      string
		Checksum  string
		CreatedAt int64
		UpdatedAt int64
	}
	meta := BlobMeta{}
	if err := json.Unmarshal(b.Meta, &meta); err != nil {
		return nil, err
	}

	return &artifacts.FileMetadata{
		Id:        b.Id,
		ParentId:  b.ParentId,
		Type:      meta.Type,
		Path:      meta.Path,
		Checksum:  meta.Checksum,
		CreatedAt: meta.CreatedAt,
		UpdatedAt: meta.UpdatedAt,
	}, nil
}

type ExperimentSchema struct {
	Id        string
	ExternId  string `db:"external_id"`
	Name      string
	Owner     string
	Namespace string
	Framework uint8 `db:"ml_framework"`
	CreatedAt int64 `db:"created_at"`
	UpdatedAt int64 `db:"updated_at"`
}

func (e *ExperimentSchema) ToExperiment() *Experiment {
	return &Experiment{
		Id:         e.Id,
		Name:       e.Name,
		Owner:      e.Owner,
		Namespace:  e.Namespace,
		ExternalId: e.ExternId,
		Framework:  MLFramework(e.Framework),
		Exists:     false,
		CreatedAt:  e.CreatedAt,
		UpdatedAt:  e.UpdatedAt,
	}
}

func (e *ExperimentSchema) mutationSchema(ex *Experiment) *MutationEventSchema {
	return &MutationEventSchema{
		MutationTime:      uint64(time.Now().Unix()),
		EventType:         uint8(EventTypeExperimentCreated),
		ObjectType:        uint8(EventObjectTypeExperiment),
		ObjectId:          e.Id,
		Namespace:         e.Namespace,
		ExperimentPayload: ex,
	}
}

func FromExperimentToSchema(experiment *Experiment) *ExperimentSchema {
	return &ExperimentSchema{
		Id:        experiment.Id,
		ExternId:  experiment.ExternalId,
		Name:      experiment.Name,
		Owner:     experiment.Owner,
		Namespace: experiment.Namespace,
		Framework: uint8(experiment.Framework),
		CreatedAt: experiment.CreatedAt,
		UpdatedAt: experiment.UpdatedAt,
	}
}

type CheckpointSchema struct {
	Id         string
	Experiment string
	Epoch      uint64
	Metrics    SerializableMetrics
	CreatedAt  int64 `db:"created_at"`
	UpdatedAt  int64 `db:"updated_at"`
}

func (c *CheckpointSchema) ToCheckpoint(files artifacts.FileSet) *Checkpoint {
	return &Checkpoint{
		Id:           c.Id,
		ExperimentId: c.Experiment,
		Epoch:        c.Epoch,
		Files:        files,
		Metrics:      c.Metrics,
		CreatedAt:    c.CreatedAt,
		UpdtedAt:     c.UpdatedAt,
	}
}

func ToCheckpointSchema(c *Checkpoint) *CheckpointSchema {
	return &CheckpointSchema{
		Id:         c.Id,
		Experiment: c.ExperimentId,
		Epoch:      c.Epoch,
		Metrics:    c.Metrics,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdtedAt,
	}
}

type MetadataSchema struct {
	Id       string
	ParentId string `db:"parent_id"`
	Metadata SerializableMetadata
}

func toMetadata(rows []*MetadataSchema) map[string]*structpb.Value {
	metadata := make(map[string]*structpb.Value)
	for _, m := range rows {
		for k, v := range m.Metadata {
			metadata[k] = v
		}
	}
	return metadata
}

func toMetadataSchema(parentId string, metadata map[string]*structpb.Value) []*MetadataSchema {
	rows := []*MetadataSchema{}
	for k, v := range metadata {
		h := sha1.New()
		utils.HashString(h, parentId)
		utils.HashString(h, k)
		id := fmt.Sprintf("%x", h.Sum(nil))
		m := &MetadataSchema{
			Id:       id,
			ParentId: parentId,
			Metadata: map[string]*structpb.Value{k: v},
		}
		rows = append(rows, m)
	}
	return rows
}

type SerializablePayload map[string]interface{}

func (s SerializablePayload) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *SerializablePayload) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &s)
}

type MutationEventSchema struct {
	MutationId            uint64 `db:"mutation_id"`
	MutationTime          uint64 `db:"mutation_time"`
	EventType             uint8  `db:"event_type"`
	ObjectId              string `db:"object_id"`
	ObjectType            uint8  `db:"object_type"`
	ParentId              string `db:"parent_id"`
	Namespace             string
	ProcessedAt           uint64          `db:"processed_at"`
	ExperimentPayload     *Experiment     `db:"experiment_payload"`
	ModelPayload          *Model          `db:"model_payload"`
	ModelVersionPayload   *ModelVersion   `db:"model_version_payload"`
	ActionPayload         *Action         `db:"action_payload"`
	ActionInstancePayload *ActionInstance `db:"action_instance_payload"`
}

func newMutationEventSchema(event *ChangeEvent) *MutationEventSchema {
	return &MutationEventSchema{
		MutationId:          event.Id,
		MutationTime:        event.CreatedAt,
		EventType:           uint8(event.EventType),
		ObjectId:            event.ObjectId,
		ObjectType:          uint8(event.ObjectType),
		Namespace:           event.Namespace,
		ProcessedAt:         event.ProcessedAt,
		ExperimentPayload:   event.Experiment,
		ModelPayload:        event.Model,
		ModelVersionPayload: event.ModelVersion,
		ActionPayload:       event.Action,
	}
}

func (m *MutationEventSchema) toChangeEvent() *ChangeEvent {
	return &ChangeEvent{
		Id:           m.MutationId,
		ObjectId:     m.ObjectId,
		EventType:    EventType(m.EventType),
		ObjectType:   EventObjectType(m.ObjectType),
		Namespace:    m.Namespace,
		ProcessedAt:  m.ProcessedAt,
		CreatedAt:    m.MutationTime,
		Experiment:   m.ExperimentPayload,
		Model:        m.ModelPayload,
		ModelVersion: m.ModelVersionPayload,
		Action:       m.ActionPayload,
	}
}

type EventSchema struct {
	Id        string
	ParentId  string `db:"parent_id"`
	Name      string
	Source    string `db:"source_name"`
	Wallclock uint64
	Metadata  SerializableMetadata `db:"metadata"`
}

type ActionSchema struct {
	Id          string
	ParentId    string `db:"parent_id"`
	Name        string
	Arch        string
	Params      SerializableMetadata `db:"params"`
	Trigger     string               `db:"trigger_predicate"`
	TriggerType int8                 `db:"trigger_type"`
	CreatedAt   uint64               `db:"created_at"`
	UpdatedAt   uint64               `db:"updated_at"`
	FinishedAt  uint64               `db:"finished_at"`
}

func (a *ActionSchema) toAction() *Action {
	return &Action{
		Id:         a.Id,
		ParentId:   a.ParentId,
		Name:       a.Name,
		Params:     a.Params,
		Arch:       a.Arch,
		Trigger:    NewTrigger(a.Trigger, TriggerType(a.TriggerType)),
		CreatedAt:  int64(a.CreatedAt),
		UpdatedAt:  int64(a.UpdatedAt),
		FinishedAt: int64(a.FinishedAt),
	}
}

func (a *ActionSchema) mutationSchema(action *Action) *MutationEventSchema {
	return &MutationEventSchema{
		MutationTime:  uint64(time.Now().Unix()),
		EventType:     uint8(EventTypeActionCreated),
		ObjectId:      a.Id,
		ObjectType:    uint8(EventObjectTypeAction),
		ParentId:      a.ParentId,
		Namespace:     "",
		ProcessedAt:   0,
		ActionPayload: action,
	}
}

func newActionSchema(action *Action) *ActionSchema {
	return &ActionSchema{
		Id:         action.Id,
		ParentId:   action.ParentId,
		Name:       action.Name,
		Arch:       action.Arch,
		Params:     action.Params,
		CreatedAt:  uint64(action.CreatedAt),
		UpdatedAt:  uint64(action.UpdatedAt),
		FinishedAt: uint64(action.FinishedAt),
	}
}

type ActionInstanceSchema struct {
	Id            string
	ActionId      string `db:"action_id"`
	Attempt       uint64
	Status        uint8
	Outcome       uint8
	OutcomeReason string `db:"outcome_reason"`
	CreatedAt     int64  `db:"created_at"`
	UpdatedAt     int64  `db:"updated_at"`
	FinishedAt    int64  `db:"finished_at"`
}

func newActionInstanceSchema(ai *ActionInstance) *ActionInstanceSchema {
	return &ActionInstanceSchema{
		Id:            ai.Id,
		ActionId:      ai.ActionId,
		Attempt:       uint64(ai.Attempt),
		Status:        uint8(ai.Status),
		Outcome:       uint8(ai.Outcome),
		OutcomeReason: ai.OutcomeReason,
		CreatedAt:     ai.CreatedAt,
		UpdatedAt:     ai.UpdatedAt,
		FinishedAt:    ai.FinishedAt,
	}
}

func (a *ActionInstanceSchema) mutationEvent(ai *ActionInstance, eventType EventType) *MutationEventSchema {
	return &MutationEventSchema{
		MutationTime:          uint64(time.Now().Unix()),
		EventType:             uint8(eventType),
		ObjectId:              ai.Id,
		ObjectType:            uint8(EventObjectTypeActionInstance),
		ParentId:              ai.ActionId,
		Namespace:             "",
		ProcessedAt:           0,
		ActionInstancePayload: ai,
	}
}

func (a *ActionInstanceSchema) toActionInstance() *ActionInstance {
	return &ActionInstance{
		Id:            a.Id,
		ActionId:      a.ActionId,
		Attempt:       uint(a.Attempt),
		Status:        ActionStatus(a.Status),
		Outcome:       ActionOutcome(a.Outcome),
		OutcomeReason: a.OutcomeReason,
		CreatedAt:     a.CreatedAt,
		UpdatedAt:     a.UpdatedAt,
		FinishedAt:    a.FinishedAt,
	}
}

type AgentSchema struct {
	NodeId        string `db:"node_id"`
	Info          *Agent `db:"info"`
	HeartbeatTime uint64 `db:"heartbeat_time"`
}

func NewAgentSchema(a *Agent, hbTime uint64) *AgentSchema {
	return &AgentSchema{
		NodeId:        a.AgentId(),
		Info:          a,
		HeartbeatTime: hbTime,
	}
}
