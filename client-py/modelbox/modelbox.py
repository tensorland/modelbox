from typing import Dict, List, Any, Union
from typing_extensions import Self
from enum import Enum
from dataclasses import dataclass
from hashlib import md5

import grpc
from . import service_pb2
from . import service_pb2_grpc
from google.protobuf.struct_pb2 import Struct
from google.protobuf.timestamp_pb2 import Timestamp

DEFAULT_NAMESPACE = "default"

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


@dataclass
class UpdateMetadataResponse:
    updated_at: Timestamp


@dataclass
class ListMetadataResponse:
    metadata: Dict


@dataclass
class Checkpoint:
    id: str
    experiment_id: str
    epoch: str


@dataclass
class ListCheckpointsResponse:
    checkpoints: List[Checkpoint]


@dataclass
class CreateExperimentResult:
    experiment_id: str
    exists: bool


@dataclass
class CreateCheckpointResponse:
    checkpoint_id: str
    exists: bool


@dataclass
class Experiment:
    id: str
    name: str
    owner: str
    namespace: str
    external_id: str
    created_at: int
    updated_at: int


@dataclass
class ListExperimentsResponse:
    experiments: List[Experiment]


@dataclass
class UploadFileResponse:
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
class Model:
    id: str
    name: str
    owner: str
    namespace: str
    task: str
    description: str
    metadata: Dict 
    artifacts: List[Artifact]


@dataclass
class ModelVersion:
    id: str
    model_id: str
    name: str
    version: str
    description: str
    files: List
    metadata: Dict
    unique_tags: List
    framework: MLFramework


@dataclass
class ListModelsResult:
    models: List[Model]


@dataclass
class MetricValue:
    step: int 
    wallclock_time: int 
    value: Union[float, str, bytes]


@dataclass
class Metrics:
    key: str
    values: List[MetricValue]


class ModelBoxClient:
    def __init__(self, addr):
        self._addr = addr
        self._channel = grpc.insecure_channel(addr)
        self._client = service_pb2_grpc.ModelStoreStub(self._channel)

    def create_model(
        self,
        name: str,
        owner: str,
        namespace: str,
        task: str,
        description: str,
        metadata: Dict,
    ) -> Model:
        req = service_pb2.CreateModelRequest(
            name=name,
            owner=owner,
            namespace=namespace,
            task=task,
            description=description,
        )
        response = self._client.CreateModel(req)
        return Model(
            response.id, name, owner, namespace, task, description, metadata, []
        )

    def list_models(self, namespace: str) -> ListModelsResult:
        req = service_pb2.ListModelsRequest(namespace=namespace)
        resp = self._client.ListModels(req)
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
                    metadata={}, #TODO Fix this when we pull meta from top level objects 
                    artifacts=artifacts,
                )
            )
        return ListModelsResult(models=result)

    def create_model_version(
        self,
        model_id: str,
        name: str,
        version: str,
        description: str,
        files: List[service_pb2.FileMetadata],
        metadata: Dict,
        framework: MLFramework,
        unique_tags: List[str],
    ) -> ModelVersion:
        req = service_pb2.CreateModelVersionRequest(
            model=model_id,
            name=name,
            version=version,
            description=description,
            files=files,
            framework=framework,
            unique_tags=unique_tags,
        )
        response = self._client.CreateModelVersion(req)
        return ModelVersion(
            id=response.model_version,
            model_id=model_id,
            name=name,
            version=version,
            description=description,
            files=files,
            metadata=metadata,
            framework=framework,
            unique_tags=unique_tags,
        )

    def create_experiment(
        self,
        name: str,
        owner: str,
        namespace: str,
        external_id: str,
        framework: MLFramework,
    ) -> CreateExperimentResult:
        req = service_pb2.CreateExperimentRequest(
            name=name,
            owner=owner,
            namespace=namespace,
            external_id=external_id,
            framework=framework.to_proto(),
        )
        response = self._client.CreateExperiment(req)
        return CreateExperimentResult(
            response.experiment_id, response.experiment_exists,
        )

    def create_checkpoint(
        self, experiment: str, epoch: int, path: str, metrics: Dict, file_checksum="",
    ) -> CreateCheckpointResponse:
        req = service_pb2.CreateCheckpointRequest(
            experiment_id=experiment,
            epoch=epoch,
            files=[
                service_pb2.FileMetadata(
                    checksum=file_checksum, path=path, file_type=service_pb2.CHECKPOINT,
                )
            ],
            metrics=metrics,
        )
        response = self._client.CreateCheckpoint(req)
        return CreateCheckpointResponse(response.checkpoint_id, False)

    def list_checkpoints(self, experiment_id: str) -> ListCheckpointsResponse:
        req = service_pb2.ListCheckpointsRequest(experiment_id=experiment_id)
        response = self._client.ListCheckpoints(req)
        checkpoints = []
        for c in response.checkpoints:
            chk = Checkpoint(id=c.id, experiment_id=c.experiment_id, epoch=c.epoch)
            checkpoints.append(chk)
        return ListCheckpointsResponse(checkpoints=checkpoints)

    def list_experiments(self, namespace: str) -> ListExperimentsResponse:
        req = service_pb2.ListExperimentsRequest(namespace=namespace)
        response = self._client.ListExperiments(req)
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
            )
            experiments.append(e)
        return ListExperimentsResponse(experiments=experiments)

    def _file_chunk_iterator(self, parent, path):
        checksum = self._file_checksum(path)
        file_meta = service_pb2.FileMetadata(
            parent_id=parent,
            checksum=checksum,
            file_type=service_pb2.CHECKPOINT,
            path=path,
        )
        yield service_pb2.UploadFileRequest(metadata=file_meta)
        with open(path, "rb") as f:
            while True:
                data = f.read(CHUNK_SZ)
                if not data:
                    break
                yield service_pb2.UploadFileRequest(chunks=data)

    def upload_file(self, parent: str, path: str) -> UploadFileResponse:
        itr = self._file_chunk_iterator(parent, path)
        resp = self._client.UploadFile(itr)
        return UploadFileResponse(id=resp.file_id)

    def download_artifact(self, id: str, path: str) -> DownloadArtifactResponse:
        req = service_pb2.DownloadFileRequest(file_id=id)
        resp_itr = self._client.DownloadFile(req)
        ret = DownloadArtifactResponse
        with open(path, "wb") as f:
            for resp in resp_itr:
                if resp.HasField("chunks"):
                    f.write(resp.chunks)
                if resp.HasField("metadata"):
                    ret.id = resp.metadata.id
                    ret.checksum = resp.metadata.checksum
                    ret.path = path
        return ret

    def update_metadata(
        self, parent_id: str, key: str, val: Any
    ) -> UpdateMetadataResponse:
        payload = Struct()
        payload.update({key: val})
        meta = service_pb2.Metadata(parent_id=parent_id, payload=payload)
        req = service_pb2.UpdateMetadataRequest(metadata=[meta])
        resp = self._client.UpdateMetadata(req)
        return UpdateMetadataResponse(updated_at=resp.updated_at)

    def list_metadata(self, id: str) -> ListMetadataResponse:
        req = service_pb2.ListMetadataRequest(parent_id=id)
        resp = self._client.ListMetadata(req)
        meta_resp = ListMetadataResponse(metadata={})
        for k, v in resp.payload.items():
            meta_resp.metadata[k] = v
        return meta_resp

    def track_artifacts(self, artifacts: List[Artifact]) -> TrackArtifactsResponse:
        proto_artifacts = []
        for artifact in artifacts:
            artifact.checksum = self._file_checksum(artifact.path)
            proto_artifacts.append(
                service_pb2.FileMetadata(
                    parent_id=artifact.parent,
                    file_type=artifact.mime_type.to_proto(),
                    checksum=artifact.checksum,
                    path=artifact.path,
                )
            )
        req = service_pb2.TrackArtifactsRequest(files=proto_artifacts)
        resp = self._client.TrackArtifacts(req)
        return TrackArtifactsResponse(num_artifacts_tracked=resp.num_files_tracked)

    def log_metrics(self, parent_id: str, key: str, value: MetricValue):
        req = service_pb2.LogMetricsRequest(
            parent_id=parent_id,
            key=key,
            value=service_pb2.MetricsValue(
                step=value.step, wallclock_time=value.wallclock_time, f_val=value.value
            ),
        )
        return self._client.LogMetrics(req)

    def get_metrics(self, parent_id: str) -> Dict[str, Metrics]:
        req = service_pb2.GetMetricsRequest(parent_id=parent_id)
        resp = self._client.GetMetrics(req)
        metrics = {}
        for metric in resp.metrics:
            m_vals = []
            for v in metric.values:
                m_vals.append(
                    MetricValue(
                        step=v.step, wallclock_time=v.wallclock_time, value=v.f_val
                    )
                )
            m = Metrics(key=metric.key, values=[m_vals])
            metrics[metric.key] = m

        return metrics

    def _file_checksum(self, path) -> str:
        checksum = ""
        with open(path, "rb") as f:
            checksum = md5(f.read()).hexdigest()
        return checksum

    def close(self):
        if self._channel is not None:
            self._channel.close()

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_value, traceback):
        return self.close()
