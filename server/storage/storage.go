package storage

import (
	"context"
	"crypto/sha1"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/tensorland/modelbox/sdk-go/proto"
	"github.com/tensorland/modelbox/server/config"
	"github.com/tensorland/modelbox/server/storage/artifacts"
	storageconfig "github.com/tensorland/modelbox/server/storage/config"
	"github.com/tensorland/modelbox/server/storage/logging"
	"github.com/tensorland/modelbox/server/utils"
	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

type SerializableMetadata map[string]*structpb.Value

func (s SerializableMetadata) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *SerializableMetadata) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &s)
}

// Status of an action associated with a model or experiment
type ActionStatus uint8

const (
	StatusPending ActionStatus = iota
	StatusRunning
	StatusFinished
)

// Outcome of the action
type ActionOutcome uint8

const (
	OutcomeUnknown ActionOutcome = iota
	OutcomeSuccess
	OutcomeFailure
)

type ActionEval struct {
	Id          string
	ParentId    string
	ParentType  string
	CreatedAt   int64
	ProcessedAt int64
}

// Action represents work associated with a model or an experiment
type Action struct {
	Id         string
	ParentId   string
	Name       string
	Command    string
	Params     map[string]*structpb.Value
	Arch       string
	CreatedAt  int64
	UpdatedAt  int64
	FinishedAt int64
}

func NewAction(name, arch, parent string, params map[string]*structpb.Value) *Action {
	h := sha1.New()
	utils.HashString(h, name)
	utils.HashString(h, parent)
	utils.HashString(h, arch)
	utils.HashMeta(h, params)
	id := fmt.Sprintf("%x", h.Sum(nil))
	currentTime := time.Now().Unix()
	return &Action{
		Id:        id,
		ParentId:  parent,
		Name:      name,
		Arch:      arch,
		Params:    params,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	}
}

func (a *Action) actionEval() *ActionEval {
	h := sha1.New()
	utils.HashString(h, a.Id)
	utils.HashString(h, "action-create")
	utils.HashUint64(h, uint64(time.Now().Unix()))
	id := fmt.Sprintf("%x", h.Sum(nil))
	return &ActionEval{
		Id:         id,
		ParentId:   a.Id,
		ParentType: "action",
		CreatedAt:  time.Now().Unix(),
	}
}

type ActionInstance struct {
	Id            string
	Attempt       uint
	Status        ActionStatus
	Outcome       ActionOutcome
	OutcomeReason string
	CreatedTime   int64
	FinishedTime  int64
}

type ActionState struct {
	Action    *Action
	Instances []*ActionInstance
}

type Event struct {
	Id              string
	ParentId        string
	Name            string
	Source          string
	SourceWallclock uint64
	Metadata        map[string]*structpb.Value
}

func NewEvent(parentId, source, name string, wallclock time.Time, metadata map[string]*structpb.Value) *Event {
	h := sha1.New()
	utils.HashString(h, parentId)
	utils.HashString(h, name)
	utils.HashString(h, source)
	utils.HashUint64(h, uint64(wallclock.Unix()))
	utils.HashMeta(h, metadata)
	id := fmt.Sprintf("%x", h.Sum(nil))
	return &Event{
		Id:              id,
		ParentId:        parentId,
		Name:            name,
		Source:          source,
		SourceWallclock: uint64(wallclock.Unix()),
		Metadata:        metadata,
	}
}

var _ msgpack.Marshaler = (*Event)(nil)

func (i *Event) MarshalMsgpack() ([]byte, error) {
	return json.Marshal(i)
}

type ChangeEvent struct {
	ObjectId   string
	Time       time.Time
	ObjectType string
	Action     string
	Namespace  string
	Payload    map[string]interface{}
}

func ToFloatLogFromProto(value *proto.MetricsValue) *logging.FloatLog {
	return &logging.FloatLog{
		Value:     value.GetFVal(),
		Step:      uint64(value.Step),
		WallClock: uint64(value.WallclockTime),
	}
}

type MLFramework uint16

const (
	Unknown MLFramework = iota
	Pytorch
	Keras
)

type SerializableTags []string

func (s *SerializableTags) Scan(val interface{}) error {
	switch v := val.(type) {
	case []byte:
		json.Unmarshal(v, &s)
		return nil
	case string:
		json.Unmarshal([]byte(v), &s)
		return nil
	default:
		return fmt.Errorf("unsupported type: %v", v)
	}
}

func (s SerializableTags) Value() (driver.Value, error) {
	return json.Marshal(s)
}

type SerializableMetrics map[string]float32

func (s *SerializableMetrics) Scan(val interface{}) error {
	switch v := val.(type) {
	case []byte:
		json.Unmarshal(v, &s)
		return nil
	case string:
		json.Unmarshal([]byte(v), &s)
		return nil
	default:
		return fmt.Errorf("unsupported type: %v", v)
	}
}

func (s SerializableMetrics) Value() (driver.Value, error) {
	return json.Marshal(s)
}

type BackendInfo struct {
	Name string
}

func (b BackendInfo) String() string {
	return b.Name
}

type Experiment struct {
	Id         string
	Name       string
	Owner      string
	Namespace  string
	ExternalId string
	Framework  MLFramework
	Exists     bool
	CreatedAt  int64
	UpdatedAt  int64
}

func NewExperiment(
	name, owner, namespace, externId string,
	fwk MLFramework,
) *Experiment {
	currentTime := time.Now().Unix()
	experiment := &Experiment{
		Name:       name,
		Owner:      owner,
		ExternalId: externId,
		Namespace:  namespace,
		Framework:  fwk,
		CreatedAt:  currentTime,
		UpdatedAt:  currentTime,
	}
	experiment.Id = experiment.Hash()
	return experiment
}

func (e *Experiment) Hash() string {
	h := sha1.New()
	utils.HashString(h, e.Name)
	utils.HashString(h, e.Namespace)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (e Experiment) Value() (driver.Value, error) {
	return json.Marshal(e)
}

func (a *Experiment) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &a)
}

type Checkpoint struct {
	Id           string
	ExperimentId string
	Epoch        uint64
	Files        artifacts.FileSet
	Metrics      SerializableMetrics
	CreatedAt    int64
	UpdtedAt     int64
}

func NewCheckpoint(
	experimentId string,
	epoch uint64,
	metrics map[string]float32,
) *Checkpoint {
	currentTime := time.Now().Unix()
	chk := &Checkpoint{
		ExperimentId: experimentId,
		Epoch:        epoch,
		Metrics:      metrics,
		CreatedAt:    currentTime,
		UpdtedAt:     currentTime,
	}
	chk.CreateId()
	return chk
}

func (c *Checkpoint) SetFiles(files artifacts.FileSet) {
	c.Files = files
}

func GetCheckpointID(experiment string, epoch uint64) string {
	h := sha1.New()
	utils.HashString(h, experiment)
	utils.HashUint64(h, epoch)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (c *Checkpoint) CreateId() {
	h := sha1.New()
	utils.HashString(h, c.ExperimentId)
	utils.HashUint64(h, c.Epoch)
	c.Id = fmt.Sprintf("%x", h.Sum(nil))
}

type Model struct {
	Id          string
	Name        string
	Owner       string
	Namespace   string
	Task        string
	Description string
	Files       artifacts.FileSet
	CreatedAt   int64
	UpdatedAt   int64
}

func NewModel(name, owner, namespace, task, description string) *Model {
	currentTime := time.Now().Unix()
	model := &Model{
		Name:        name,
		Owner:       owner,
		Namespace:   namespace,
		Task:        task,
		Description: description,
		CreatedAt:   currentTime,
		UpdatedAt:   currentTime,
	}
	model.CreateId()
	return model
}

func (m *Model) CreateId() {
	h := sha1.New()
	utils.HashString(h, m.Name)
	utils.HashString(h, m.Namespace)
	m.Id = fmt.Sprintf("%x", h.Sum(nil))
}

func (m *Model) SetFiles(files artifacts.FileSet) {
	m.Files = files
}

type ModelVersion struct {
	Id          string
	Name        string
	ModelId     string
	Version     string
	Description string
	Framework   MLFramework
	Files       artifacts.FileSet
	UniqueTags  SerializableTags
	CreatedAt   int64
	UpdatedAt   int64
}

func NewModelVersion(name, model, version, description string,
	framework MLFramework,
	files []*artifacts.FileMetadata,
	uniqueTags []string,
) *ModelVersion {
	currentTime := time.Now().Unix()
	mv := &ModelVersion{
		Name:        name,
		ModelId:     model,
		Version:     version,
		Description: description,
		Framework:   framework,
		Files:       files,
		UniqueTags:  uniqueTags,
		CreatedAt:   currentTime,
		UpdatedAt:   currentTime,
	}
	mv.CreateId()
	return mv
}

func (m *ModelVersion) CreateId() {
	h := sha1.New()
	utils.HashString(h, m.ModelId)
	utils.HashString(h, m.Version)
	utils.HashString(h, m.Name)
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
	CreateExperiment(ctx context.Context, experiment *Experiment, metadata SerializableMetadata) (*CreateExperimentResult, error)

	GetExperiment(ctx context.Context, id string) (*Experiment, error)

	CreateCheckpoint(ctx context.Context, checkpoint *Checkpoint, metadata SerializableMetadata) (*CreateCheckpointResult, error)

	ListExperiments(ctx context.Context, namespace string) ([]*Experiment, error)

	ListCheckpoints(ctx context.Context, experimentId string) ([]*Checkpoint, error)

	GetCheckpoint(ctx context.Context, checkpointId string) (*Checkpoint, error)

	CreateModel(ctx context.Context, model *Model, metadata SerializableMetadata) (*CreateModelResult, error)

	GetModel(ctx context.Context, id string) (*Model, error)

	CreateModelVersion(ctx context.Context, modelVersion *ModelVersion, metadata SerializableMetadata) (*CreateModelVersionResult, error)

	GetModelVersion(ctx context.Context, id string) (*ModelVersion, error)

	ListModels(ctx context.Context, namespace string) ([]*Model, error)

	ListModelVersions(ctx context.Context, modelId string) ([]*ModelVersion, error)

	Ping() error

	CreateSchema(schema string) error

	Backend() *BackendInfo

	WriteFiles(context.Context, artifacts.FileSet) error

	GetFiles(ctx context.Context, parentId string) (artifacts.FileSet, error)

	GetFile(ctx context.Context, id string) (*artifacts.FileMetadata, error)

	UpdateMetadata(ctx context.Context, parentId string, metadata map[string]*structpb.Value) error

	ListMetadata(ctx context.Context, parentId string) (map[string]*structpb.Value, error)

	ListChanges(ctx context.Context, namespace string, since time.Time) ([]*ChangeEvent, error)

	LogEvent(ctx context.Context, parentId string, event *Event) error

	ListEvents(ctx context.Context, parentId string) ([]*Event, error)

	CreateAction(ctx context.Context, action *Action) error

	ListActions(ctx context.Context, parentId string) ([]*Action, error)

	GetAction(ctx context.Context, id string) (*ActionState, error)

	GetActionEvals(ctx context.Context) ([]*ActionEval, error)

	Close() error
}

func NewMetadataStorage(
	svrConfig *config.ServerConfig,
	logger *zap.Logger,
) (MetadataStorage, error) {
	switch svrConfig.MetadataBackend {
	case config.METADATA_BACKEND_MYSQL:
		mysqlConfig := svrConfig.MySQLConfig
		if mysqlConfig == nil {
			return nil, fmt.Errorf("mysql config is not set up")
		}
		return NewMySqlStorage(&storageconfig.MySqlStorageConfig{
			Host:     mysqlConfig.Host,
			Port:     mysqlConfig.Port,
			Password: mysqlConfig.Password,
			UserName: mysqlConfig.User,
			DbName:   mysqlConfig.DbName,
		}, logger)
	case config.METADATA_BACKEND_POSTGRES:
		postgresConfig := svrConfig.PostgresConfig
		if postgresConfig == nil {
			return nil, fmt.Errorf("postgres config is not set up")
		}
		return NewPostgresStorage(
			&storageconfig.PostgresConfig{
				Host:     postgresConfig.Host,
				Port:     postgresConfig.Port,
				Password: postgresConfig.Password,
				UserName: postgresConfig.User,
				DbName:   postgresConfig.DbName,
			}, logger)
	case config.METADATA_BACKEND_SQLITE3:
		sqliteConfig := &storageconfig.Sqlite3Config{
			File: svrConfig.SqliteConfig.Path,
		}
		return NewSqlite3Storage(sqliteConfig, logger)

	}
	return nil, fmt.Errorf("unknown metadata backend: %v", svrConfig.MetadataBackend)
}
