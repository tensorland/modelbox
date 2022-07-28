package storage

import (
	"encoding/json"

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

func (m *ModelSchema) ToModel(blobs BlobSet) *Model {
	model := &Model{
		Id:          m.Id,
		Name:        m.Name,
		Owner:       m.Owner,
		Namespace:   m.Namespace,
		Task:        m.Task,
		Meta:        m.Meta,
		Description: m.Desc,
		Blobs:       []*BlobInfo{},
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
	if blobs != nil {
		model.Blobs = blobs
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

func (m *ModelVersionSchema) ToModelVersion(blobs BlobSet) *ModelVersion {
	modelVersion := &ModelVersion{
		Id:          m.Id,
		Name:        m.Name,
		ModelId:     m.Model,
		Version:     m.Version,
		Description: m.Desc,
		Framework:   MLFramework(m.Framework),
		Meta:        m.Meta,
		Blobs:       blobs,
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

func ToBlobSet(rows []BlobSchema) ([]*BlobInfo, error) {
	blobSet := make([]*BlobInfo, len(rows))
	for i, row := range rows {
		blob, err := row.ToBlob()
		if err != nil {
			return nil, err
		}
		blobSet[i] = blob
	}
	return blobSet, nil
}

type BlobSchema struct {
	Id       string
	ParentId string         `db:"parent_id"`
	Meta     types.JSONText `db:"metadata"`
}

func (b *BlobSchema) ToBlob() (*BlobInfo, error) {
	type BlobMeta struct {
		Type      BlobType
		Path      string
		Checksum  string
		CreatedAt int64
		UpdatedAt int64
	}
	meta := BlobMeta{}
	if err := json.Unmarshal(b.Meta, &meta); err != nil {
		return nil, err
	}

	return &BlobInfo{
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

func (c *CheckpointSchema) ToCheckpoint(blobs BlobSet) *Checkpoint {
	return &Checkpoint{
		Id:           c.Id,
		ExperimentId: c.Experiment,
		Epoch:        c.Epoch,
		Blobs:        blobs,
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
