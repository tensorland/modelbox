package storage

import (
	"encoding/json"
	"fmt"
)

type ModelSchema struct {
	Id        string
	Name      string
	Owner     string
	Namespace string
	Task      string
	Meta      []uint8 `db:"metadata"`
	Desc      string  `db:"description"`
	CreatedAt int64   `db:"created_at"`
	UpdatedAt int64   `db:"updated_at"`
}

func (m *ModelSchema) ToModel(blobs BlobSet) (*Model, error) {
	meta := make(map[string]string)
	if err := json.Unmarshal(m.Meta, &meta); err != nil {
		return nil, err
	}
	model := &Model{
		Id:          m.Id,
		Name:        m.Name,
		Owner:       m.Owner,
		Namespace:   m.Namespace,
		Task:        m.Task,
		Meta:        meta,
		Description: m.Desc,
		Blobs:       []*BlobInfo{},
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
	if blobs != nil {
		model.Blobs = blobs
	}
	return model, nil
}

type ModelVersionSchema struct {
	Id         string
	Name       string
	Model      string `db:"model_id"`
	Version    string
	Desc       string  `db:"description"`
	Framework  int8    `db:"ml_framework"`
	Meta       []uint8 `db:"metadata"`
	UniqueTags []uint8 `db:"unique_tags"`
	Tags       []uint8 `db:"tags"`
	CreatedAt  int64   `db:"created_at"`
	UpdatedAt  int64   `db:"updated_at"`
}

func (m *ModelVersionSchema) ToModelVersion(blobs BlobSet) (*ModelVersion, error) {
	meta := make(map[string]string)
	if err := json.Unmarshal(m.Meta, &meta); err != nil {
		return nil, err
	}
	uniqueTags, err := SerializableTagsFromBytes(m.UniqueTags)
	if err != nil {
		return nil, err
	}
	modelVersion := &ModelVersion{
		Id:          m.Id,
		Name:        m.Name,
		ModelId:     m.Model,
		Version:     m.Version,
		Description: m.Desc,
		Framework:   MLFramework(m.Framework),
		Meta:        meta,
		Blobs:       blobs,
		UniqueTags:  uniqueTags,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
	return modelVersion, nil
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
	ParentId string  `db:"parent_id"`
	Meta     []uint8 `db:"metadata"`
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
	Framework uint8   `db:"ml_framework"`
	Meta      []uint8 `db:"metadata"`
	CreatedAt int64   `db:"created_at"`
	UpdatedAt int64   `db:"updated_at"`
}

func (e *ExperimentSchema) ToExperiment() (*Experiment, error) {
	meta := make(map[string]string)
	if err := json.Unmarshal(e.Meta, &meta); err != nil {
		return nil, err
	}
	experiment := &Experiment{
		Id:         e.Id,
		Name:       e.Name,
		Owner:      e.Owner,
		Namespace:  e.Namespace,
		ExternalId: e.ExternId,
		Framework:  MLFramework(e.Framework),
		Meta:       meta,
		Exists:     false,
		CreatedAt:  e.CreatedAt,
		UpdatedAt:  e.UpdatedAt,
	}
	return experiment, nil
}

type CheckpointSchema struct {
	Id         string
	Experiment string
	Epoch      uint64
	Path       string
	Checksum   string
	State      int
	Metrics    []uint8
	Meta       []uint8 `db:"metadata"`
	CreatedAt  int64   `db:"created_at"`
	UpdatedAt  int64   `db:"updated_at"`
}

func (c *CheckpointSchema) ToCheckpoint(blobs BlobSet) (*Checkpoint, error) {
	meta := make(map[string]string)
	if err := json.Unmarshal(c.Meta, &meta); err != nil {
		return nil, fmt.Errorf("can't unmarshall metadata: %v", err)
	}
	metrics := make(map[string]float32)
	if err := json.Unmarshal(c.Metrics, &metrics); err != nil {
		return nil, fmt.Errorf("can't unmarshall metrics: %v", err)
	}
	return &Checkpoint{
		Id:           c.Id,
		ExperimentId: c.Experiment,
		Epoch:        c.Epoch,
		Blobs:        blobs,
		Meta:         meta,
		Metrics:      metrics,
		CreatedAt:    c.CreatedAt,
		UpdtedAt:     c.UpdatedAt,
	}, nil
}
