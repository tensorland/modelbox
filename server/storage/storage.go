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

type Agent struct {
	Id       string
	Name     string
	Actions  []string
	HostName string
	IpAddr   string
	Arch     string
}

func NewAgent(name, hostname, ipAddr, arch string, actions []string) *Agent {
	agent := &Agent{
		Name:     name,
		Actions:  actions,
		HostName: hostname,
		IpAddr:   ipAddr,
		Arch:     arch,
	}
	h := sha1.New()
	utils.HashString(h, name)
	utils.HashString(h, hostname)
	utils.HashString(h, ipAddr)
	utils.HashString(h, arch)
	for _, action := range actions {
		utils.HashString(h, action)
	}
	agent.Id = fmt.Sprintf("%x", h.Sum(nil))
	return agent
}

func (a Agent) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *Agent) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &a)
}

type Heartbeat struct {
	AgentId string
	Time    uint64
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

type EventType uint8

const (
	EventTypeUnknown EventType = iota

	EventTypeModelCreated
	EventTypeModelUpdated
	EventTypeModelVersionCreated
	EventTypeModelVersioneUpdated

	EventTypeExperimentCreated
	EventTypeExperimentUpdated

	EventTypeActionCreated

	EventTypeActionInstanceCreated
	EventTypeActionInstancePending
	EventTypeActionInstanceRunning
	EventTypeActionInstanceSuccess
	EventTypeActionInstanceFailure
)

type TriggerType uint8

const (
	TriggerTypeJs TriggerType = iota
)

/**
 * Trigger is a user-defined function that invokes an action
 */
type Trigger struct {
	Payload string
	Type    TriggerType
}

func (t *Trigger) GetAction(event *ChangeEvent) *Action {
	return &Action{}
}

func NewTrigger(payload string, t TriggerType) *Trigger {
	return &Trigger{
		Payload: payload,
		Type:    t,
	}
}

// Action represents work associated with a model or an experiment
type Action struct {
	Id         string
	ParentId   string
	Name       string
	Command    string
	Params     map[string]*structpb.Value
	Trigger    *Trigger
	Arch       string
	CreatedAt  int64
	UpdatedAt  int64
	FinishedAt int64
}

func NewAction(name, arch, parent string, trigger *Trigger, params map[string]*structpb.Value) *Action {
	h := sha1.New()
	utils.HashString(h, name)
	utils.HashString(h, parent)
	utils.HashString(h, arch)
	utils.HashString(h, trigger.Payload)
	utils.HashMeta(h, params)
	id := fmt.Sprintf("%x", h.Sum(nil))
	currentTime := time.Now().Unix()
	return &Action{
		Id:        id,
		ParentId:  parent,
		Name:      name,
		Arch:      arch,
		Params:    params,
		Trigger:   trigger,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	}
}

func (s *Action) Scan(val interface{}) error {
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

func (s Action) Value() (driver.Value, error) {
	return json.Marshal(s)
}

type ActionInstance struct {
	Id            string
	ActionId      string
	Attempt       uint
	Status        ActionStatus
	Outcome       ActionOutcome
	OutcomeReason string
	CreatedAt     int64
	UpdatedAt     int64
	FinishedAt    int64
}

func (a *ActionInstance) Scan(val interface{}) error {
	switch v := val.(type) {
	case []byte:
		json.Unmarshal(v, &a)
		return nil
	case string:
		json.Unmarshal([]byte(v), &a)
		return nil
	default:
		return fmt.Errorf("unsupported type: %v", v)
	}
}

func (a ActionInstance) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func NewActionInstance(actionId string, attempt uint) *ActionInstance {
	createdTime := time.Now().Unix()
	h := sha1.New()
	utils.HashString(h, actionId)
	utils.HashUint64(h, uint64(attempt))
	id := fmt.Sprintf("%x", h.Sum(nil))
	return &ActionInstance{
		Id:        id,
		ActionId:  actionId,
		Attempt:   attempt,
		Status:    StatusPending,
		Outcome:   OutcomeUnknown,
		CreatedAt: createdTime,
	}
}

func (a *ActionInstance) Update(update *ActionInstanceUpdate) bool {
	// This prevents updating the instance when the same update is applied twice
	if a.Status >= update.Status {
		return false
	}
	a.Status = update.Status
	a.Outcome = update.Outcome
	a.OutcomeReason = update.OutcomeReason
	if update.Outcome == OutcomeFailure || update.Outcome == OutcomeSuccess {
		a.FinishedAt = int64(update.Time)
	}
	if update.Outcome == OutcomeFailure {
		return true
	}

	return true
}

type ActionState struct {
	Action    *Action
	Instances []*ActionInstance
}

type ActionInstanceUpdate struct {
	ActionInstanceId string
	Status           ActionStatus
	Outcome          ActionOutcome
	OutcomeReason    string
	Time             uint64
}

func NewActionInstanceUpdate(instanceId string, status ActionStatus, outcome ActionOutcome, reason string, time uint64) *ActionInstanceUpdate {
	return &ActionInstanceUpdate{
		ActionInstanceId: instanceId,
		Status:           status,
		Outcome:          outcome,
		OutcomeReason:    reason,
		Time:             time,
	}
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

type EventObjectType uint8

const (
	EventObjectTypeModel EventObjectType = iota
	EventObjectTypeModelVersion
	EventObjectTypeExperiment
	EventObjectTypeAction
	EventObjectTypeActionInstance
)

type ChangeEvent struct {
	Id             uint64
	ObjectId       string
	EventType      EventType
	ObjectType     EventObjectType
	Namespace      string
	ProcessedAt    uint64
	CreatedAt      uint64
	Experiment     *Experiment
	Model          *Model
	ModelVersion   *ModelVersion
	Action         *Action
	ActionInstance *ActionInstance
}

func NewChangeEvent(objectId string, createdAt uint64, eventType EventType, objType EventObjectType, namespace string, payload interface{}) *ChangeEvent {
	ce := &ChangeEvent{
		ObjectId:  objectId,
		EventType: eventType,
		Namespace: namespace,
		CreatedAt: createdAt,
	}

	switch objType {
	case EventObjectTypeModel:
		m, _ := payload.(*Model)
		ce.Model = m
	case EventObjectTypeExperiment:
		e, _ := payload.(*Experiment)
		ce.Experiment = e
	case EventObjectTypeModelVersion:
		mv, _ := payload.(*ModelVersion)
		ce.ModelVersion = mv
	case EventObjectTypeAction:
		a, _ := payload.(*Action)
		ce.Action = a
	case EventObjectTypeActionInstance:
		ai, _ := payload.(*ActionInstance)
		ce.ActionInstance = ai
	}
	return ce
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

func (m *Model) Scan(val interface{}) error {
	switch v := val.(type) {
	case []byte:
		json.Unmarshal(v, &m)
		return nil
	case string:
		json.Unmarshal([]byte(v), &m)
		return nil
	default:
		return fmt.Errorf("unsupported type: %v", v)
	}
}

func (s Model) Value() (driver.Value, error) {
	return json.Marshal(s)
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

func (m *ModelVersion) Scan(val interface{}) error {
	switch v := val.(type) {
	case []byte:
		json.Unmarshal(v, &m)
		return nil
	case string:
		json.Unmarshal([]byte(v), &m)
		return nil
	default:
		return fmt.Errorf("unsupported type: %v", v)
	}
}

func (s ModelVersion) Value() (driver.Value, error) {
	return json.Marshal(s)
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

	GetUnprocessedChangeEvents(ctx context.Context) ([]*ChangeEvent, error)

	GetChangeEventsForObject(ctx context.Context, id string) ([]*ChangeEvent, error)

	CreateAction(ctx context.Context, action *Action) error

	CreateActionInstance(ctx context.Context, actionInstance *ActionInstance, event *ChangeEvent) error

	ListActions(ctx context.Context, parentId string) ([]*Action, error)

	GetAction(ctx context.Context, id string) (*ActionState, error)

	UpdateActionInstance(ctx context.Context, instance *ActionInstance) error

	RegisterNode(ctx context.Context, agent *Agent) error

	Heartbeat(ctx context.Context, hb *Heartbeat) error

	GetDeadAgents(ctx context.Context) ([]*Agent, error)

	GetActionInstance(ctx context.Context, id string) (*ActionInstance, error)

	GetActionInstances(ctx context.Context, status ActionStatus) ([]*ActionInstance, error)

	GetTriggers(ctx context.Context, parentId string) ([]*Trigger, error)

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
