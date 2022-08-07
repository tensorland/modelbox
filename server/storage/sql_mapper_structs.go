package storage

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx/types"
)

type ModelSchema struct {
	Id        string
	Name      string
	Owner     string
	Namespace string
	Task      string
	Meta      SerializableMeta `db:"metadata"`
	Desc      string           `db:"description"`
	CreatedAt int64            `db:"created_at"`
	UpdatedAt int64            `db:"updated_at"`
}

func (m *ModelSchema) ToModel(files FileSet) *Model {
	model := &Model{
		Id:          m.Id,
		Name:        m.Name,
		Owner:       m.Owner,
		Namespace:   m.Namespace,
		Task:        m.Task,
		Meta:        m.Meta,
		Description: m.Desc,
		Files:       []*FileMetadata{},
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
	if files != nil {
		model.Files = files
	}
	return model
}

func ModelToSchema(m *Model) *ModelSchema {
	return &ModelSchema{
		Id:        m.Id,
		Name:      m.Name,
		Owner:     m.Owner,
		Namespace: m.Namespace,
		Task:      m.Task,
		Meta:      m.Meta,
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
	Meta       SerializableMeta `db:"metadata"`
	UniqueTags SerializableTags `db:"unique_tags"`
	Tags       SerializableTags `db:"tags"`
	CreatedAt  int64            `db:"created_at"`
	UpdatedAt  int64            `db:"updated_at"`
}

func (m *ModelVersionSchema) ToModelVersion(files FileSet) *ModelVersion {
	modelVersion := &ModelVersion{
		Id:          m.Id,
		Name:        m.Name,
		ModelId:     m.Model,
		Version:     m.Version,
		Description: m.Desc,
		Framework:   MLFramework(m.Framework),
		Meta:        m.Meta,
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
		Meta:       mv.Meta,
		UniqueTags: mv.UniqueTags,
		CreatedAt:  mv.CreatedAt,
		UpdatedAt:  mv.UpdatedAt,
	}
}

func ToFileSet(rows []FileSchema) ([]*FileMetadata, error) {
	fileSet := make([]*FileMetadata, len(rows))
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

func (b *FileSchema) ToFile() (*FileMetadata, error) {
	type BlobMeta struct {
		Type      FileMIMEType
		Path      string
		Checksum  string
		CreatedAt int64
		UpdatedAt int64
	}
	meta := BlobMeta{}
	if err := json.Unmarshal(b.Meta, &meta); err != nil {
		return nil, err
	}

	return &FileMetadata{
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
	Framework uint8            `db:"ml_framework"`
	Meta      SerializableMeta `db:"metadata"`
	CreatedAt int64            `db:"created_at"`
	UpdatedAt int64            `db:"updated_at"`
}

func (e *ExperimentSchema) ToExperiment() *Experiment {
	return &Experiment{
		Id:         e.Id,
		Name:       e.Name,
		Owner:      e.Owner,
		Namespace:  e.Namespace,
		ExternalId: e.ExternId,
		Framework:  MLFramework(e.Framework),
		Meta:       e.Meta,
		Exists:     false,
		CreatedAt:  e.CreatedAt,
		UpdatedAt:  e.UpdatedAt,
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
		Meta:      experiment.Meta,
		CreatedAt: experiment.CreatedAt,
		UpdatedAt: experiment.UpdatedAt,
	}
}

type CheckpointSchema struct {
	Id         string
	Experiment string
	Epoch      uint64
	Metrics    SerializableMetrics
	Meta       SerializableMeta `db:"metadata"`
	CreatedAt  int64            `db:"created_at"`
	UpdatedAt  int64            `db:"updated_at"`
}

func (c *CheckpointSchema) ToCheckpoint(files FileSet) *Checkpoint {
	return &Checkpoint{
		Id:           c.Id,
		ExperimentId: c.Experiment,
		Epoch:        c.Epoch,
		Files:        files,
		Meta:         c.Meta,
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
		Meta:       c.Meta,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdtedAt,
	}
}

type MetadataSchema struct {
	Id       string
	ParentId string `db:"parent_id"`
	Metadata types.JSONText
}

func (c *MetadataSchema) toMetadata() (*Metadata, error) {
	var m map[string]interface{}
	json.Unmarshal(c.Metadata, &m)
	k, _ := m["key"].(string)
	v := m["value"]
	return &Metadata{
		Id:       c.Id,
		ParentId: c.ParentId,
		Key:      k,
		Value:    v,
	}, nil
}

func toMetadataSchema(m *Metadata) (*MetadataSchema, error) {
	meta := map[string]interface{}{"key": m.Key, "value": m.Value}
	b, err := json.Marshal(meta)
	if err != nil {
		return nil, fmt.Errorf("unable to convert to json: %v", err)
	}

	return &MetadataSchema{
		Id:       m.Id,
		ParentId: m.ParentId,
		Metadata: b,
	}, nil
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
	MutationId   uint64 `db:"mutation_id"`
	MutationTime uint64 `db:"mutation_time"`
	Action       string
	ObjectId     string `db:"object_id"`
	ObjectType   string `db:"object_type"`
	ParentId     string `db:"parent_id"`
	Namespace    string
	Payload      SerializablePayload
}
