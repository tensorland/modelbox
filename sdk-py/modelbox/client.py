import os
from os import times
from typing import Any, Union, Dict, List
import json
from dataclasses import dataclass
from hashlib import md5

import grpc
from . import service_pb2
from . import service_pb2_grpc
from google.protobuf.struct_pb2 import Value
from google.protobuf import json_format
from google.protobuf import timestamp_pb2

# The chunk size at which files are being read.
CHUNK_SZ = 1024


def file_checksum(path) -> str:
    checksum = ""
    with open(path, "rb") as f:
        checksum = md5(f.read()).hexdigest()
    return checksum


@dataclass
class ClientFileUploadResult:
    file_id: str
    artifact_id: str


@dataclass
class ClientTrackArtifactsResult:
    id: str


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
        value = json.dumps(value)
        meta = service_pb2.Metadata(metadata={key: value})
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
        for key, metric in resp.metrics.items():
            m_vals = []
            for v in metric.values:
                m_vals.append(
                    MetricValue(
                        step=v.step, wallclock_time=v.wallclock_time, value=v.f_val
                    )
                )
            metrics[key] = m_vals

        return metrics

    def create_model(
        self,
        name: str,
        owner: str,
        namespace: str,
        task: str,
        description: str,
    ) -> service_pb2.Model:
        req = service_pb2.CreateModelRequest(
            name=name,
            owner=owner,
            namespace=namespace,
            task=task,
            description=description,
        )
        return self._client.CreateModel(req)

    def _file_chunk_iterator(
        self, artifact_name: str, object_id: str, path: str, file_type_proto: int
    ):
        checksum = file_checksum(path)
        file_meta = service_pb2.FileMetadata(
            parent_id=object_id,
            checksum=checksum,
            file_type=file_type_proto,
            src_path=path,
        )
        upload_meta = service_pb2.UploadFileMetadata(
            artifact_name=artifact_name,
            object_id=object_id,
            metadata=file_meta,
        )
        yield service_pb2.UploadFileRequest(metadata=upload_meta)
        with open(path, "rb") as f:
            while True:
                data = f.read(CHUNK_SZ)
                if not data:
                    break
                yield service_pb2.UploadFileRequest(chunks=data)

    def upload_artifact(
        self, artifact_name: str, object_id: str, path: str, file_type_proto: int
    ) -> ClientFileUploadResult:
        itr = self._file_chunk_iterator(artifact_name, object_id, path, file_type_proto)
        resp = self._client.UploadFile(itr)
        return ClientFileUploadResult(
            file_id=resp.file_id, artifact_id=resp.artifact_id
        )

    def download_asset(self, id: str, dst_path: str) -> ClientFileDownloadResult:
        req = service_pb2.DownloadFileRequest(file_id=id)
        resp_itr = self._client.DownloadFile(req)
        ret = ClientFileDownloadResult
        src_path, checksum = None, None
        for resp in resp_itr:
            if resp.HasField("metadata"):
                src_path = resp.metadata.src_path
                checksum = resp.metadata.checksum
        file_name = os.path.join(dst_path, src_path)
        os.makedirs(os.path.dirname(file_name), exist_ok=True)
        with open(file_name, "wb") as f:
            for resp in resp_itr:
                if resp.HasField("chunks"):
                    f.write(resp.chunks)
        return ret

    def track_artifacts(
        self, name: str, object_id: str, files: List[service_pb2.FileMetadata]
    ):
        req = service_pb2.TrackArtifactsRequest(
            name=name, object_id=object_id, files=files
        )
        resp = self._client.TrackArtifacts(req)
        return ClientTrackArtifactsResult(id=resp.id)

    def list_artifacts(self, object_id: str) -> service_pb2.ListArtifactsResponse:
        return self._client.ListArtifacts(
            service_pb2.ListArtifactsRequest(object_id=object_id)
        )

    def log_event(
        self, parent_id: str, event: service_pb2.Event
    ) -> service_pb2.LogEventResponse:
        req = service_pb2.LogEventRequest(parent_id=parent_id, event=event)
        return self._client.LogEvent(req)

    def list_events(self, parent_id: str) -> service_pb2.ListEventsRequest:
        return self._client.ListEvents(
            service_pb2.ListEventsRequest(
                parent_id=parent_id, since=timestamp_pb2.Timestamp(seconds=0)
            )
        )

    def list_metadata(self, id: str) -> Dict:
        req = service_pb2.ListMetadataRequest(parent_id=id)
        resp = self._client.ListMetadata(req)
        if (resp.metadata is None) or (resp.metadata.metadata is None):
            return {}
        return resp.metadata.metadata

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
            framework=framework_proto,
            unique_tags=unique_tags,
        )
        # TODO: Add files
        return self._client.CreateModelVersion(req)

    def list_model_versions(
        self, model_id: str
    ) -> service_pb2.ListModelVersionsResponse:
        return self._client.ListModelVersions(
            service_pb2.ListModelVersionsRequest(model=model_id)
        )

    def get_experiment(self, id: str) -> service_pb2.GetExperimentResponse:
        return self._client.GetExperiment(service_pb2.GetExperimentRequest(id=id))

    def close(self):
        if self._channel is not None:
            self._channel.close()

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_value, traceback):
        return self.close()
