package storage

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"strconv"
	"time"

	"github.com/diptanu/modelbox/proto"
	"github.com/diptanu/modelbox/server/config"
	"go.uber.org/zap"
)

type MLFramework uint16

const (
	Unknown MLFramework = iota
	Pytorch
	Keras
)

type BlobType uint8

const (
	UnknownBlob BlobType = iota
	CheckpointBlob
	ModelBlob
	File
)

type SerializableMeta map[string]string

func (s *SerializableMeta) SerializeToJson() (string, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func NewSerializableMeta(bytes []byte) (SerializableMeta, error) {
	meta := make(map[string]string)
	if err := json.Unmarshal(bytes, &meta); err != nil {
		return nil, fmt.Errorf("unable to convert json to meta: %v", err)
	}
	return meta, nil
}

type SerializableTags []string

func (s *SerializableTags) SerializeToJson() (string, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func SerializableTagsFromBytes(bytes []uint8) (SerializableTags, error) {
	b := []string{}
	if err := json.Unmarshal(bytes, &b); err != nil {
		return nil, fmt.Errorf("unable to convert json to tags: %v", err)
	}
	return b, nil
}

type BlobSet []*BlobInfo

func (b *BlobSet) ToJson() ([]byte, error) {
	bytes, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func NewBlobSetFromBytes(bytes []byte) (BlobSet, error) {
	b := make([]*BlobInfo, 0)
	if err := json.Unmarshal(bytes, &b); err != nil {
		return nil, fmt.Errorf("unable to convert json to blobset: %v", err)
	}
	return b, nil
}

func NewBlobFromProto(parent string, pbBlob *proto.BlobMetadata) *BlobInfo {
	return NewBlobInfo(
		parent,
		pbBlob.Path,
		pbBlob.Checksum,
		BlobType(pbBlob.BlobType),
		pbBlob.CreatedAt.AsTime().Unix(),
		pbBlob.UpdatedAt.AsTime().Unix(),
	)
}

func NewBlobSetFromProto(parent string, pb []*proto.BlobMetadata) BlobSet {
	blobs := make([]*BlobInfo, len(pb))
	for i, b := range pb {
		blobs[i] = NewBlobFromProto(parent, b)
	}
	return blobs
}

/*
 * BlobInfo are metadata about files and other blobs such as models.
 * They can be associated with any modelbox object.
 */
type BlobInfo struct {
	Id        string
	ParentId  string
	Type      BlobType
	Path      string
	Checksum  string
	CreatedAt int64
	UpdatedAt int64
}

func (b *BlobInfo) CreateId() {
	h := sha1.New()
	hashString(h, b.ParentId)
	hashInt(h, int(b.Type))
	hashString(h, b.Checksum)
	b.Id = fmt.Sprintf("%x", h.Sum(nil))
}

func (b *BlobInfo) ToJson() ([]byte, error) {
	bytes, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

type BackendInfo struct {
	Name string
}

func (b BackendInfo) String() string {
	return b.Name
}

func NewBlobInfo(
	parent, path, checksum string,
	blobType BlobType,
	createdAt, updatedAt int64,
) *BlobInfo {
	currentTime := time.Now().Unix()
	if createdAt == 0 {
		createdAt = currentTime
	}
	if updatedAt == 0 {
		updatedAt = currentTime
	}
	blob := &BlobInfo{
		ParentId:  parent,
		Path:      path,
		Checksum:  checksum,
		Type:      blobType,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
	blob.CreateId()
	return blob
}

func MLFrameworkFromProto(fwk proto.MLFramework) MLFramework {
	switch fwk {
	case proto.MLFramework_PYTORCH:
		return Pytorch
	case proto.MLFramework_KERAS:
		return Keras
	}
	return Unknown
}

func MLFrameworkToProto(fwk MLFramework) proto.MLFramework {
	switch fwk {
	case Pytorch:
		return proto.MLFramework_PYTORCH
	case Keras:
		return proto.MLFramework_KERAS
	}
	return proto.MLFramework_UNKNOWN
}

func BlobTypeFromProto(t proto.BlobType) BlobType {
	switch t {
	case proto.BlobType_CHECKPOINT:
		return CheckpointBlob
	case proto.BlobType_MODEL:
		return ModelBlob
	}
	return UnknownBlob
}

func BlobTypeToProto(t BlobType) proto.BlobType {
	switch t {
	case CheckpointBlob:
		return proto.BlobType_CHECKPOINT
	case ModelBlob:
		return proto.BlobType_MODEL
	}
	return proto.BlobType_UNDEFINED
}

type BlobOpenMode uint8

const (
	Read BlobOpenMode = iota
	Write
)

type CheckpointState int

const (
	CheckpointInitalized CheckpointState = iota
	CheckpointReady
)

type ModelVersionState uint8

const (
	ModelVersionInitialized ModelVersionState = iota
	ModelVersionBlobsCommitted
)

func hashString(h hash.Hash, s string) {
	_, _ = io.WriteString(h, s)
}

func hashUint64(h hash.Hash, i uint64) {
	_, _ = io.WriteString(h, strconv.FormatUint(i, 10))
}

func hashInt(h hash.Hash, i int) {
	_, _ = io.WriteString(h, strconv.Itoa(i))
}

type Experiment struct {
	Id         string
	Name       string
	Owner      string
	Namespace  string
	ExternalId string
	Framework  MLFramework
	Meta       SerializableMeta
	Exists     bool
	CreatedAt  int64
	UpdatedAt  int64
}

func NewExperiment(
	name, owner, namespace, externId string,
	fwk MLFramework,
	meta map[string]string,
) *Experiment {
	currentTime := time.Now().Unix()
	experiment := &Experiment{
		Name:       name,
		Owner:      owner,
		ExternalId: externId,
		Namespace:  namespace,
		Framework:  fwk,
		Meta:       meta,
		CreatedAt:  currentTime,
		UpdatedAt:  currentTime,
	}
	experiment.Id = experiment.Hash()
	return experiment
}

func (e *Experiment) SerialializeMeta() (string, error) {
	b, err := json.Marshal(e.Meta)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
func (e *Experiment) Hash() string {
	h := sha1.New()
	hashString(h, e.Name)
	hashString(h, e.Namespace)
	return fmt.Sprintf("%x", h.Sum(nil))
}

type Checkpoint struct {
	Id           string
	ExperimentId string
	Epoch        uint64
	Blobs        BlobSet
	Meta         map[string]string
	Metrics      map[string]float32
	CreatedAt    int64
	UpdtedAt     int64
}

func NewCheckpoint(
	experimentId string,
	epoch uint64,
	blobs BlobSet,
	meta map[string]string,
	metrics map[string]float32) *Checkpoint {
	currentTime := time.Now().Unix()
	chk := &Checkpoint{
		ExperimentId: experimentId,
		Epoch:        epoch,
		Blobs:        blobs,
		Meta:         meta,
		Metrics:      metrics,
		CreatedAt:    currentTime,
		UpdtedAt:     currentTime,
	}
	chk.CreateId()
	return chk
}

func (c *Checkpoint) CreateId() {
	h := sha1.New()
	hashString(h, c.ExperimentId)
	hashUint64(h, c.Epoch)
	c.Id = fmt.Sprintf("%x", h.Sum(nil))
}

func (c *Checkpoint) SerializeMetrics() (string, error) {
	b, err := json.Marshal(c.Metrics)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c *Checkpoint) SerialializeMeta() (string, error) {
	b, err := json.Marshal(c.Meta)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

type Model struct {
	Id          string
	Name        string
	Owner       string
	Namespace   string
	Task        string
	Meta        SerializableMeta
	Description string
	Blobs       BlobSet
	CreatedAt   int64
	UpdatedAt   int64
}

func NewModel(name, owner, namespace, task, description string,
	meta map[string]string) *Model {
	currentTime := time.Now().Unix()
	model := &Model{
		Name:        name,
		Owner:       owner,
		Namespace:   namespace,
		Task:        task,
		Description: description,
		Meta:        meta,
		CreatedAt:   currentTime,
		UpdatedAt:   currentTime,
	}
	model.CreateId()
	return model
}

func (m *Model) CreateId() {
	h := sha1.New()
	hashString(h, m.Name)
	hashString(h, m.Namespace)
	m.Id = fmt.Sprintf("%x", h.Sum(nil))
}

func (m *Model) SetBlobs(blobs BlobSet) {
	m.Blobs = blobs
}

type ModelVersion struct {
	Id          string
	Name        string
	ModelId     string
	Version     string
	Description string
	Framework   MLFramework
	Meta        SerializableMeta
	Blobs       BlobSet
	UniqueTags  SerializableTags
	CreatedAt   int64
	UpdatedAt   int64
}

func NewModelVersion(name, model, version, description string,
	framework MLFramework,
	meta map[string]string,
	blobs []*BlobInfo,
	uniqueTags []string) *ModelVersion {
	currentTime := time.Now().Unix()
	mv := &ModelVersion{
		Name:        name,
		ModelId:     model,
		Version:     version,
		Description: description,
		Framework:   framework,
		Meta:        meta,
		Blobs:       blobs,
		UniqueTags:  uniqueTags,
		CreatedAt:   currentTime,
		UpdatedAt:   currentTime,
	}
	mv.CreateId()
	return mv
}

func (m *ModelVersion) CreateId() {
	h := sha1.New()
	hashString(h, m.ModelId)
	hashString(h, m.Version)
	hashString(h, m.Name)
	m.Id = fmt.Sprintf("%x", h.Sum(nil))
}

type CreateErr struct {
	Exists bool
	what   string
}

func NewCreateErr(what string, exists bool) CreateErr {
	return CreateErr{Exists: exists, what: what}
}

func (e *CreateErr) Error() string {
	return e.what
}

type CreateExperimentResult struct {
	ExperimentId string
	Exists       bool
}

type CreateCheckpointResult struct {
	CheckpointId string
	CreatedAt    int64
	UpdatedAt    int64
}

type CreateModelResult struct {
	ModelId   string
	CreatedAt int64
	UpdatedAt int64
}

type CreateModelVersionResult struct {
	ModelVersionId string
	CreatedAt      int64
	UpdatedAt      int64
}

type MetadataStorage interface {
	CreateExperiment(ctx context.Context, experiment *Experiment) (*CreateExperimentResult, error)

	CreateCheckpoint(ctx context.Context, checkpoint *Checkpoint) (*CreateCheckpointResult, error)

	ListExperiments(ctx context.Context, namespace string) ([]*Experiment, error)

	ListCheckpoints(ctx context.Context, experimentId string) ([]*Checkpoint, error)

	GetCheckpoint(ctx context.Context, checkpointId string) (*Checkpoint, error)

	CreateModel(ctx context.Context, model *Model) (*CreateModelResult, error)

	GetModel(ctx context.Context, id string) (*Model, error)

	CreateModelVersion(ctx context.Context, modelVersion *ModelVersion) (*CreateModelVersionResult, error)

	GetModelVersion(ctx context.Context, id string) (*ModelVersion, error)

	ListModels(ctx context.Context, namespace string) ([]*Model, error)

	ListModelVersions(ctx context.Context, modelId string) ([]*ModelVersion, error)

	Ping() error

	CreateSchema(schema string) error

	Backend() *BackendInfo

	WriteBlobs(context.Context, BlobSet) error

	GetBlobs(ctx context.Context, parentId string) (BlobSet, error)

	UpdateBlobPath(ctx context.Context, path string, parentId string, t BlobType) error

	DeleteExperiment(ctx context.Context, id string) error

	Close() error
}

func NewMetadataStorage(
	svrConfig *config.ServerConfig,
	logger *zap.Logger,
) (MetadataStorage, error) {
	switch svrConfig.MetadataBackend {
	case config.METADATA_BACKEND_INTEGRATED:
		return NewEphemeralStorage(svrConfig.IntegratedStorage.Path, logger)
	case config.METADATA_BACKEND_MYSQL:
		mysqlConfig := svrConfig.MySQLConfig
		if mysqlConfig == nil {
			return nil, fmt.Errorf("mysql config is not set up")
		}
		return NewMySqlStorage(&MySqlStorageConfig{
			Host:     mysqlConfig.Host,
			Port:     mysqlConfig.Port,
			Password: mysqlConfig.Password,
			UserName: mysqlConfig.User,
			DbName:   mysqlConfig.DbName,
		}, logger)
	}
	return nil, fmt.Errorf("unknown metadata backend: %v", svrConfig.MetadataBackend)
}

func NewBlobStorageBuilder(
	svrConfig *config.ServerConfig,
	logger *zap.Logger,
) (BlobStorageBuilder, error) {
	switch svrConfig.StorageBackend {
	case config.BLOB_STORAGE_BACKEND_FS:
		return NewFileBlobStorageBuilder(svrConfig.FileStorage.BaseDir, logger)
	}
	return nil, fmt.Errorf("unknown blob storage backend: %v", svrConfig.StorageBackend)
}

type BlobStorage interface {
	Open(id string, mode BlobOpenMode) error

	GetPath() (string, error)

	io.ReadWriteCloser
}

type BlobStorageBuilder interface {
	Build() BlobStorage
}
