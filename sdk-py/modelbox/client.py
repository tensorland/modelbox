from typing import Any, Union, Dict, List
import json
from dataclasses import dataclass
from hashlib import md5

import grpc
from . import service_pb2
from . import service_pb2_grpc
from google.protobuf.struct_pb2 import Value
from google.protobuf import json_format
from google.protobuf.timestamp_pb2 import Timestamp

# The chunk size at which files are being read.
CHUNK_SZ = 1024


def file_checksum(path) -> str:
    checksum = ""
    with open(path, "rb") as f:
        checksum = md5(f.read()).hexdigest()
    return checksum


@dataclass
class ClientFileUploadResult:
    id: str


@dataclass
class ClientTrackArtifactsResult:
    num_artifacts_tracked: int


@dataclass
class ClientFileDownloadResult:
    id: str
    path: str
    checksum: str


@dataclass
class MetricValue:
    step: int
    wallclock_time: int
    value: Union[float, str, bytes]


class ModelBoxClient:
    def __init__(self, addr):
        self._addr = addr
        self._channel = grpc.insecure_channel(addr)
        self._client = service_pb2_grpc.ModelStoreStub(self._channel)

    def create_experiment(
        self,
        name: str,
        owner: str,
        namespace: str,
        external_id: str,
        framework_proto: int,
    ) -> service_pb2.CreateExperimentResponse:
        req = service_pb2.CreateExperimentRequest(
            name=name,
            owner=owner,
            namespace=namespace,
            external_id=external_id,
            framework=framework_proto,
        )
        return self._client.CreateExperiment(req)

    def update_metadata(
        self, parent_id: str, key: str, value: Any
    ) -> service_pb2.UpdateMetadataResponse:
        json_value = Value()
        json_format.Parse(json.dumps(value), json_value)
        meta = service_pb2.Metadata(metadata={key: json_value})
        req = service_pb2.UpdateMetadataRequest(parent_id=parent_id, metadata=meta)
        return self._client.UpdateMetadata(req)

    def log_metrics(
        self, parent_id: str, key: str, value: MetricValue
    ) -> service_pb2.LogMetricsResponse:
        req = service_pb2.LogMetricsRequest(
            parent_id=parent_id,
            key=key,
            value=service_pb2.MetricsValue(
                step=value.step, wallclock_time=value.wallclock_time, f_val=value.value
            ),
        )
        return self._client.LogMetrics(req)

    def get_all_metrics(self, parent_id: str) -> Dict[str, List[MetricValue]]:
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
            metrics[metric.key] = m_vals

        return metrics

    def create_model(
        self,
        name: str,
        owner: str,
        namespace: str,
        task: str,
        description: str,
        metadata: Dict,
    ) -> service_pb2.Model:
        req = service_pb2.CreateModelRequest(
            name=name,
            owner=owner,
            namespace=namespace,
            task=task,
            description=description,
        )
        return self._client.CreateModel(req)

    def _file_chunk_iterator(self, parent: str, path: str, file_type_proto: int):
        checksum = file_checksum(path)
        file_meta = service_pb2.FileMetadata(
            parent_id=parent,
            checksum=checksum,
            file_type=file_type_proto,
            path=path,
        )
        yield service_pb2.UploadFileRequest(metadata=file_meta)
        with open(path, "rb") as f:
            while True:
                data = f.read(CHUNK_SZ)
                if not data:
                    break
                yield service_pb2.UploadFileRequest(chunks=data)

    def upload_artifact(
        self, parent: str, path: str, file_type_proto: int
    ) -> ClientFileUploadResult:
        itr = self._file_chunk_iterator(parent, path, file_type_proto)
        resp = self._client.UploadFile(itr)
        return ClientFileUploadResult(id=resp.file_id)

    def download_artifact(self, id: str, path: str) -> ClientFileDownloadResult:
        req = service_pb2.DownloadFileRequest(file_id=id)
        resp_itr = self._client.DownloadFile(req)
        ret = ClientFileDownloadResult
        with open(path, "wb") as f:
            for resp in resp_itr:
                if resp.HasField("chunks"):
                    f.write(resp.chunks)
                if resp.HasField("metadata"):
                    ret.id = resp.metadata.id
                    ret.checksum = resp.metadata.checksum
                    ret.path = path
        return ret

    def track_artifacts(self, files: List[service_pb2.FileMetadata]):
        req = service_pb2.TrackArtifactsRequest(files=files)
        resp = self._client.TrackArtifacts(req)
        return ClientTrackArtifactsResult(num_artifacts_tracked=resp.num_files_tracked)

    def log_event(
        self, parent_id: str, event: service_pb2.Event
    ) -> service_pb2.LogEventResponse:
        req = service_pb2.LogEventRequest(parent_id=parent_id, event=event)
        return self._client.LogEvent(req)

    def list_metadata(self, id: str) -> Dict:
        req = service_pb2.ListMetadataRequest(parent_id=id)
        resp = self._client.ListMetadata(req)
        meta = {}
        for k, v in resp.metadata.items():
            meta[k] = v
        return meta

    def list_models(self, namespace: str) -> service_pb2.ListModelsResponse:
        req = service_pb2.ListModelsRequest(namespace=namespace)
        return self._client.ListModels(req)

    def list_experiments(self, namespace: str) -> service_pb2.ListExperimentsResponse:
        req = service_pb2.ListExperimentsRequest(namespace=namespace)
        return self._client.ListExperiments(req)

    def create_model_version(
        self,
        model_id: str,
        version: str,
        name: str,
        description: str,
        files: List[service_pb2.FileMetadata],
        framework_proto: int,
        unique_tags: List[str],
    ) -> service_pb2.CreateModelVersionResponse:
        req = service_pb2.CreateModelVersionRequest(
            model=model_id,
            name=name,
            version=version,
            description=description,
            files=files,
            framework=framework_proto,
            unique_tags=unique_tags,
        )
        return self._client.CreateModelVersion(req)

    def create_checkpoint(
        self,
        experiment_id: str,
        epoch: int,
        files: List[service_pb2.FileMetadata],
        metrics=Dict[str, float],
    ) -> service_pb2.CreateCheckpointResponse:
        req = service_pb2.CreateCheckpointRequest(
            experiment_id=experiment_id,
            epoch=epoch,
            files=files,
            metrics=metrics,
        )
        return self._client.CreateCheckpoint(req)

    def list_checkpoints(self, experiment_id: str) -> service_pb2.ListCheckpointsResponse:
        req = service_pb2.ListCheckpointsRequest(experiment_id=experiment_id)
        return self._client.ListCheckpoints(req)

    def close(self):
        if self._channel is not None:
            self._channel.close()

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_value, traceback):
        return self.close()
