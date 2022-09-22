from __future__ import annotations
from typing import Dict, List, Any, Union
from enum import Enum
from dataclasses import InitVar, dataclass, field
from hashlib import md5
import json
from modelbox.client import ModelBoxClient, file_checksum
import os
import time

import grpc
from . import service_pb2
from . import service_pb2_grpc
from google.protobuf.struct_pb2 import Value
from google.protobuf import json_format
from google.protobuf.timestamp_pb2 import Timestamp

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
            return service_pb2.Model
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


@dataclass
class Artifact:
    parent: str
    path: str
    mime_type: ArtifactMime
    checksum: str = ""
    id: str = ""

    def to_proto(self, parent: str) -> service_pb2.FileMetadata:
        return service_pb2.FileMetadata(
            parent_id=parent,
            path=self.path,
            checksum=self.checksum,
            file_type=self.mime_type.to_proto(),
            id=self.id,
        )

    def update_checksum(self):
        if self.checksum == "":
            self.checksum = file_checksum(self.path)


@dataclass
class EventSource:
    name: str


@dataclass
class Event:
    name: str
    source: EventSource
    wallclock_time: int
    metadata: Dict


@dataclass
class LogEventResponse:
    created_at: int


@dataclass
class UpdateMetadataResponse:
    updated_at: Timestamp


@dataclass
class LogMetricsResponse:
    updated_at: Timestamp


@dataclass
class ListMetadataResponse:
    metadata: Dict


@dataclass
class CreateExperimentResult:
    experiment_id: str
    exists: bool


@dataclass
class CreateCheckpointResponse:
    checkpoint_id: str
    exists: bool


@dataclass
class UploadArtifactResponse:
    id: str


@dataclass
class DownloadArtifactResponse:
    id: str
    path: str
    checksum: str


@dataclass
class TrackArtifactsResponse:
    num_artifacts_tracked: int


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
            json_value = Value()
            json_format.Parse(json.dumps(v), json_value)
            meta_dict[k] = json_value
        event = service_pb2.Event(
            name=event.name,
            source=service_pb2.EventSource(name=event.source.name),
            metadata=service_pb2.Metadata(metadata=meta_dict),
        )
        ret = self._client.log_event(self._id, event)
        return LogEventResponse(created_at=ret.created_at.ToSeconds())


class MetricsLoggerMixin:
    def __init__(self, parent_id: str, client: ModelBoxClient) -> None:
        self._id = parent_id
        self._client = client

    def log_metrics(
        self,
        key: str,
        value=Union[float, str, bytes],
        step: int = 0,
        wallclock: int = int(time.time()),
    ) -> LogMetricsResponse:
        response = self._client.log_metrics(
            self._id, key, MetricValue(step, wallclock, value)
        )
        return LogMetricsResponse(updated_at=int(time.time()))

    def get_all_metrics(self) -> Dict[str, List[MetricValue]]:
        return self._client.get_all_metrics(self._id)


class MetadataLoggerMixin:
    def __init__(self, parent_id: str, client: ModelBoxClient) -> None:
        self._id = parent_id
        self._client = client

    def update_metadata(self, key: str, value: Any) -> UpdateMetadataResponse:
        response = self._client.update_metadata(self._id, key, value)
        return UpdateMetadataResponse(response.updated_at)

    def metadata(self) -> ListMetadataResponse:
        resp = self._client.list_metadata(self._id)
        return ListMetadataResponse(metadata=resp)


class ArtifactLoggerMixin:
    def __init__(self, parent_id: str, client: ModelBoxClient) -> None:
        self._id = parent_id
        self._client = client

    def track_artifacts(self, artifacts: List[Artifact]) -> TrackArtifactsResponse:
        proto_artifacts = []
        for artifact in artifacts:
            artifact.update_checksum()
            proto_artifacts.append(artifact.to_proto(self._id))
            proto_artifacts.append(
                service_pb2.FileMetadata(
                    parent_id=artifact.parent,
                    file_type=artifact.mime_type.to_proto(),
                    checksum=artifact.checksum,
                    path=artifact.path,
                )
            )
        resp = self._client.track_artifacts(proto_artifacts)
        return TrackArtifactsResponse(num_artifacts_tracked=resp.num_artifacts_tracked)

    def list_artifacts(self) -> List[Artifact]:
        pass

    def upload_artifact(self, artifact: Artifact) -> UploadArtifactResponse:
        resp = self._client.upload_artifact(
            self._id, artifact.path, artifact.mime_type.to_proto()
        )
        return UploadArtifactResponse(resp.id)

    def download_artifact(self, id: str, path: str) -> DownloadArtifactResponse:
        resp = self._client.download_artifact(id, path)
        return DownloadArtifactResponse(id, path, resp.checksum)


@dataclass
class Checkpoint(
    ArtifactLoggerMixin, MetricsLoggerMixin, MetadataLoggerMixin, EventLoggerMixin
):
    id: str
    experiment_id: str
    epoch: str
    artifacts: List[Artifact]
    metrics: Dict[str, List[MetricValue]]
    _client: InitVar[ModelBoxClient] = None

    def __post_init__(self, _client: ModelBoxClient):
        ArtifactLoggerMixin.__init__(self, self.id, _client)
        MetricsLoggerMixin.__init__(self, self.id, _client)
        MetadataLoggerMixin.__init__(self, self.id, _client)
        EventLoggerMixin.__init__(self, self.id, _client)
        self._client = _client


@dataclass
class ListCheckpointsResponse:
    checkpoints: List[Checkpoint]


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
    _metadata: Dict
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
    _metadata: Dict
    _artifacts: List[Artifact] = field(default_factory=list)
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
        metadata: Dict = {},
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
            _metadata=metadata,
            unique_tags=unique_tags,
            framework=framework,
            _client=self._client,
        )


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

    # TODO Extend the api to track artifacts. Currently it's another line of code.
    def new_checkpoint(self, epoch: int, metrics=Dict[str, float]) -> Checkpoint:
        resp = self._client.create_checkpoint(self.id, epoch, [], metrics)
        return Checkpoint(
            id=resp.checkpoint_id,
            experiment_id=self.id,
            epoch=epoch,
            artifacts=[],
            metrics=metrics,
            _client=self._client,
        )

    def list_checkpoints(self) -> List[Checkpoint]:
        result = []
        resp = self._client.list_checkpoints(self.id)
        for cp_proto in resp.checkpoints:
            result.append(
                Checkpoint(
                    id=cp_proto.id,
                    experiment_id=self.id,
                    epoch=cp_proto.epoch,
                    artifacts=[],
                    metrics={},
                    _client=self._client,
                )
            )
        return result

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

    def get_experiment(self, id: str) -> Experiment:
        pass

    def new_model(
        self,
        name: str,
        owner: str,
        namespace: str,
        task: str,
        description: str,
        artifacts: List[Artifact] = [],
        metadata: Dict = {},
    ) -> Model:
        response = self._client.create_model(
            name, owner, namespace, task, description, metadata
        )
        # TODO Plumb in the artifacts into create_model
        return Model(
            response.id,
            name,
            owner,
            namespace,
            task,
            description,
            metadata,
            artifacts,
            self._client,
        )

    def get_model(self, id: str) -> Model:
        pass

    def list_experiments(self, namespace: str):
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

    def list_models(self, namespace: str) -> ListModelsResult:
        resp = self._client.list_models(namespace)
        result = []
        for m in resp.models:
            artifacts = []
            for f in m.files:
                artifacts.append(
                    Artifact(
                        parent=f.parent_id,
                        path=f.path,
                        checksum=f.checksum,
                        mime_type=ArtifactMime(f.file_type),
                        id=f.id,
                    )
                )
            result.append(
                Model(
                    id=m.id,
                    name=m.name,
                    owner=m.owner,
                    namespace=m.namespace,
                    task=m.task,
                    description=m.description,
                    _metadata={},  # TODO Fix this when we pull meta from top level objects
                    _artifacts=artifacts,
                    _client=self._client,
                )
            )
        return ListModelsResult(models=result)
