from __future__ import annotations
from genericpath import isdir
from typing import Dict, List, Any, Union
from enum import Enum
from dataclasses import InitVar, dataclass, field
import json
from modelbox.client import ModelBoxClient, file_checksum
import os
import time

from . import service_pb2
from google.protobuf.struct_pb2 import Value
from google.protobuf import json_format
from google.protobuf.timestamp_pb2 import Timestamp
from google.protobuf import json_format

DEFAULT_NAMESPACE = "default"

DEFAULT_API_ADDR = "localhost:8085"

# The chunk size at which files are being read.
CHUNK_SZ = 1024


class MLFramework(Enum):
    UNKNOWN = 1
    PYTORCH = 2
    TENSORFLOW = 3

    def to_proto(self) -> service_pb2.MLFramework:
        if self == self.PYTORCH:
            return service_pb2.PYTORCH
        if self == self.TENSORFLOW:
            return service_pb2.KERAS
        return service_pb2.UNKNOWN

    def from_proto(p):
        if p == service_pb2.PYTORCH:
            return MLFramework.PYTORCH
        if p == service_pb2.KERAS:
            return MLFramework.TENSORFLOW
        return MLFramework.UNKNOWN


class ArtifactMime(Enum):
    Unknown = 0
    ModelVersion = 1
    Checkpoint = 2
    Text = 3
    Image = 4
    Video = 5
    Audio = 6

    def to_proto(self) -> service_pb2.FileType:
        if self == self.ModelVersion:
            return service_pb2.MODEL
        if self == self.Checkpoint:
            return service_pb2.CHECKPOINT
        if self == self.Text:
            return service_pb2.TEXT
        if self == self.Image:
            return service_pb2.IMAGE
        if self == self.Video:
            return service_pb2.VIDEO
        if self == self.Audio:
            return service_pb2.AUDIO

        return service_pb2.UNDEFINED

    @staticmethod
    def from_path(path: str) -> ArtifactMime:
        _, ext = os.path.splitext(path)
        if ext in ["jpg", "png", "jpeg", "gif", "bmp"]:
            return ArtifactMime.Image
        if ext in ["mp4", "ogg", "ogv", "mov", "m4v", "mkv", "webm"]:
            return ArtifactMime.Audio
        if ext in ["pt", "pth"]:
            return ArtifactMime.ModelVersion
        if ext in ["txt"]:
            return ArtifactMime.Text
        return ArtifactMime.Unknown


class LocalFile:
    def __init__(
        self,
        path: str,
        checksum: str = "",
        mime_type: ArtifactMime = ArtifactMime.Unknown,
    ):
        self.path = path
        self.checksum = checksum
        self.mime_type = mime_type

    @classmethod
    def from_path(cls, p: str, checksum: str = "") -> LocalFile:
        p = os.path.abspath(p)
        _checksum = file_checksum(p) if checksum == "" else checksum
        mime_type = ArtifactMime.from_path(p)
        return cls(p, _checksum, mime_type)

    def to_proto(self, parent_id: str) -> service_pb2.FileMetadata:
        return service_pb2.FileMetadata(
            parent_id=parent_id,
            src_path=self.path,
            checksum=self.checksum,
            file_type=self.mime_type.to_proto(),
        )


@dataclass
class ArtifactAsset:
    parent: str
    src_path: str
    upload_path: str = ""
    mime_type: ArtifactMime = ArtifactMime.Unknown
    checksum: str = ""
    id: str = ""

    @staticmethod
    def from_proto(file: service_pb2.FileMetadata) -> ArtifactAsset:
        return ArtifactAsset(
            parent=file.parent_id,
            src_path=file.src_path,
            upload_path=file.upload_path,
            mime_type=ArtifactMime(file.file_type),
            checksum=file.checksum,
            id=file.id,
        )


@dataclass
class EventSource:
    name: str


@dataclass
class Event:
    name: str
    source: EventSource
    wallclock_time: int
    metadata: Dict

    def from_proto(ev: service_pb2.Event) -> Event:
        metadata = {}
        for k, v in ev.metadata.metadata.items():
            metadata[k] = json_format.MessageToDict(v)

        return Event(
            name=ev.name,
            source=EventSource(name=ev.source.name),
            wallclock_time=ev.wallclock_time,
            metadata=metadata,
        )


@dataclass
class LogEventResponse:
    created_at: int


@dataclass
class UpdateMetadataResponse:
    ...


@dataclass
class LogMetricsResponse:
    updated_at: Timestamp


@dataclass
class ListMetadataResponse:
    metadata: Dict


@dataclass
class MetricValue:
    step: int
    wallclock_time: int
    value: Union[float, str, bytes]


@dataclass
class Metrics:
    key: str
    values: List[MetricValue]


class EventLoggerMixin:
    def __init__(self, parent_id: str, client: ModelBoxClient) -> None:
        self._id = parent_id
        self._client = client

    def log_event(self, event: Event) -> LogEventResponse:
        meta_dict = {}
        for k, v in event.metadata.items():
            meta_dict[k] = json.dumps(v)
        wallclock_time = Timestamp()
        wallclock_time.GetCurrentTime()
        event = service_pb2.Event(
            name=event.name,
            source=service_pb2.EventSource(name=event.source.name),
            wallclock_time=wallclock_time,
            metadata=service_pb2.Metadata(metadata=meta_dict),
        )
        ret = self._client.log_event(self._id, event)
        return LogEventResponse(created_at=ret.created_at.ToSeconds())

    def events(self) -> List[Event]:
        resp = self._client.list_events(self._id)
        ret = []
        for event in resp.events:
            ret.append(Event.from_proto(event))
        return ret


class MetricsLoggerMixin:
    def __init__(self, parent_id: str, client: ModelBoxClient) -> None:
        self._id = parent_id
        self._client = client

    def log_metrics(
        self,
        metrics: Dict[str, Union[float, str, bytes]],
        step: int = 0,
        wallclock: int = int(time.time()),
    ) -> LogMetricsResponse:
        for k, v in metrics.items():
            response = self._client.log_metrics(
                self._id, k, MetricValue(step, wallclock, v)
            )
        return LogMetricsResponse(updated_at=int(time.time()))

    def all_metrics(self) -> Dict[str, List[MetricValue]]:
        return self._client.get_all_metrics(self._id)


class MetadataLoggerMixin:
    def __init__(self, parent_id: str, client: ModelBoxClient) -> None:
        self._id = parent_id
        self._client = client

    def update_metadata(self, key: str, value: Any) -> UpdateMetadataResponse:
        response = self._client.update_metadata(self._id, key, value)
        return UpdateMetadataResponse()

    def metadata(self) -> ListMetadataResponse:
        resp = self._client.list_metadata(self._id)
        metadata = {}
        for k, v in resp.items():
            metadata[k] = json.loads(v)
        return ListMetadataResponse(metadata=metadata)


class ArtifactLoggerMixin:
    def __init__(self, parent_id: str, client: ModelBoxClient) -> None:
        self._parent_id = parent_id
        self._client = client

    @property
    def artifacts(self) -> List[Artifact]:
        resp = self._client.list_artifacts(self._id)
        artifact_list = []
        for artifact in resp.artifacts:
            artifact_list.append(Artifact.from_proto(artifact, self._client))
        return artifact_list

    def track_file(self, artifact_name: str, f: LocalFile) -> Artifact:
        proto = f.to_proto(self._parent_id)
        self._client.track_artifacts(artifact_name, self._parent_id, [proto])

    def upload_file(self, artifact_name: str, f: LocalFile):
        self._client.upload_artifact(
            artifact_name,
            self._parent_id,
            f.path,
            f.mime_type.to_proto(),
        )


@dataclass
class Artifact(MetricsLoggerMixin, EventLoggerMixin):
    name: str
    id: str
    parent_id: str
    assets: List[ArtifactAsset] = field(default_factory=list)
    _client: InitVar[ModelBoxClient] = None

    def __post_init__(self, _client: ModelBoxClient):
        MetricsLoggerMixin.__init__(self, self.id, _client)
        EventLoggerMixin.__init__(self, self.id, _client)
        self._client = _client

    @staticmethod
    def from_proto(artifact: service_pb2.Artifact, client: ModelBoxClient) -> Artifact:
        assets = []
        for file in artifact.files:
            assets.append(ArtifactAsset.from_proto(file))
        return Artifact(
            name=artifact.name,
            id=artifact.id,
            parent_id=artifact.object_id,
            assets=assets,
            _client=client,
        )

    def download(self, dst_path: str):
        for asset in self.assets:
            self._client.download_asset(asset.id, dst_path)


@dataclass
class ModelVersion(
    ArtifactLoggerMixin, MetricsLoggerMixin, MetadataLoggerMixin, EventLoggerMixin
):
    id: str
    model_id: str
    name: str
    version: str
    description: str
    files: List
    unique_tags: List = field(default_factory=list)
    framework: MLFramework = MLFramework.PYTORCH
    _client: InitVar[ModelBoxClient] = None

    def __post_init__(self, _client: ModelBoxClient):
        ArtifactLoggerMixin.__init__(self, self.id, _client)
        MetricsLoggerMixin.__init__(self, self.id, _client)
        MetadataLoggerMixin.__init__(self, self.id, _client)
        EventLoggerMixin.__init__(self, self.id, _client)
        self._client = _client


@dataclass
class Model(
    MetricsLoggerMixin, MetadataLoggerMixin, ArtifactLoggerMixin, EventLoggerMixin
):
    id: str
    name: str
    owner: str
    namespace: str
    task: str
    description: str
    _client: InitVar[ModelBoxClient] = None

    def __post_init__(self, _client: ModelBoxClient):
        MetricsLoggerMixin.__init__(self, self.id, _client)
        MetadataLoggerMixin.__init__(self, self.id, _client)
        ArtifactLoggerMixin.__init__(self, self.id, _client)
        EventLoggerMixin.__init__(self, self.id, _client)
        self._client = _client

    def new_model_version(
        self,
        version: str,
        name: str = "",
        description: str = "",
        artifacts: List[Artifact] = [],
        unique_tags: List[str] = [],
        framework: MLFramework = MLFramework.PYTORCH,
    ):
        files = [artifact.to_proto() for artifact in artifacts]
        response = self._client.create_model_version(
            self.id,
            version,
            name,
            description,
            files,
            framework.to_proto(),
            unique_tags,
        )
        return ModelVersion(
            id=response.model_version,
            model_id=self.id,
            name=name,
            version=version,
            description=description,
            files=artifacts,
            unique_tags=unique_tags,
            framework=framework,
            _client=self._client,
        )

    def model_versions(self) -> List[ModelVersion]:
        resp = self._client.list_model_versions(self._id)
        result = []
        for mv in resp.models:
            result.append(
                ModelVersion(
                    id=mv.id,
                    model_id=self._id,
                    name=mv.name,
                    version=mv.version,
                    description=mv.description,
                    _client=self._client,
                )
            )
        return result


@dataclass
class ListModelsResult:
    models: List[Model]


@dataclass
class Experiment(
    MetricsLoggerMixin, MetadataLoggerMixin, ArtifactLoggerMixin, EventLoggerMixin
):
    id: str
    name: str
    owner: str
    namespace: str
    external_id: str
    created_at: int
    updated_at: int
    framework: MLFramework
    _client: InitVar[ModelBoxClient] = None

    def __post_init__(self, _client: ModelBoxClient):
        MetricsLoggerMixin.__init__(self, self.id, _client)
        MetadataLoggerMixin.__init__(self, self.id, _client)
        ArtifactLoggerMixin.__init__(self, self.id, _client)
        EventLoggerMixin.__init__(self, self.id, _client)
        self._client = _client

    # TODO Plumb this into parent_experiment when we have the graph of experiments
    def parent(self) -> Experiment:
        pass

    # TODO Find the children when we have the graph of experiments
    def children(self) -> List[Experiment]:
        pass


@dataclass
class ListExperimentsResponse:
    experiments: List[Experiment]


class ModelBox:
    def __init__(self, addr: str = "") -> None:
        if addr == "":
            addr = os.get_env("MODELBOX_API_ADDR", DEFAULT_API_ADDR)
        self._client = ModelBoxClient(addr)

    def new_experiment(
        self,
        name: str,
        owner: str,
        namespace: str,
        external_id: str,
        framework: MLFramework,
    ) -> Experiment:
        response = self._client.create_experiment(
            name, owner, namespace, external_id, framework.to_proto()
        )
        return Experiment(
            id=response.experiment_id,
            name=name,
            owner=owner,
            namespace=namespace,
            external_id=external_id,
            framework=framework,
            created_at=response.created_at,
            updated_at=response.updated_at,
            _client=self._client,
        )

    def experiment(self, id: str) -> Experiment:
        resp = self._client.get_experiment(id)
        return Experiment(
            id=id,
            name=resp.experiment.name,
            owner=resp.experiment.owner,
            namespace=resp.experiment.namespace,
            external_id=resp.experiment.external_id,
            created_at=resp.experiment.created_at.ToSeconds(),
            updated_at=resp.experiment.updated_at.ToSeconds(),
            framework=MLFramework.from_proto(resp.experiment.framework),
            _client=self._client,
        )

    def new_model(
        self,
        name: str,
        owner: str,
        namespace: str,
        task: str,
        description: str,
        artifacts: List[Artifact] = [],
    ) -> Model:
        response = self._client.create_model(name, owner, namespace, task, description)
        # TODO Plumb in the artifacts into create_model
        return Model(
            response.id,
            name,
            owner,
            namespace,
            task,
            description,
            self._client,
        )

    def model(self, id: str) -> Model:
        pass

    def experiments(self, namespace: str):
        response = self._client.list_experiments(namespace)
        experiments = []
        for exp in response.experiments:
            e = Experiment(
                id=exp.id,
                name=exp.name,
                owner=exp.owner,
                namespace=exp.namespace,
                external_id=exp.external_id,
                created_at=exp.created_at,
                updated_at=exp.updated_at,
                framework=MLFramework.from_proto(exp.framework),
                _client=self._client,
            )
            experiments.append(e)
        return ListExperimentsResponse(experiments=experiments)

    def models(self, namespace: str) -> ListModelsResult:
        resp = self._client.list_models(namespace)
        result = []
        for m in resp.models:
            result.append(
                Model(
                    id=m.id,
                    name=m.name,
                    owner=m.owner,
                    namespace=m.namespace,
                    task=m.task,
                    description=m.description,
                    _client=self._client,
                )
            )
        return ListModelsResult(models=result)
