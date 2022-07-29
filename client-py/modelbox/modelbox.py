from datetime import datetime
from re import I, S
from time import time
from typing import Dict, List
from typing_extensions import Self
from enum import Enum
from dataclasses import dataclass
from hashlib import md5

import grpc
from grpc_status import rpc_status
from numpy import float32, uint, uint64
from regex import W
from . import service_pb2
from . import service_pb2_grpc

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
class UploadArtifactResponse:
    id: str


@dataclass
class DownloadArtifactResponse:
    id: str
    path: str
    checksum: str


@dataclass
class Model:
    id: str
    name: str
    owner: str
    namespace: str
    task: str
    description: str
    metadata: str


@dataclass
class ModelVersion:
    id: str
    model_id: str
    name: str
    version: str
    description: str
    blobs: List
    metadata: Dict
    unique_tags: List
    framework: MLFramework


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
            metadata=metadata,
        )
        response = self._client.CreateModel(req)
        return Model(response.id, name, owner, namespace, task, description, metadata)

    def create_model_version(
        self,
        model_id: str,
        name: str,
        version: str,
        description: str,
        blobs: List[service_pb2.BlobMetadata],
        metadata: Dict,
        framework: MLFramework,
        unique_tags: List[str],
    ) -> ModelVersion:
        req = service_pb2.CreateModelVersionRequest(
            model=model_id,
            name=name,
            version=version,
            description=description,
            blobs=blobs,
            metadata=metadata,
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
            blobs=blobs,
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
        self, experiment: str, epoch: uint, path: str, metrics: Dict,
    ) -> CreateCheckpointResponse:
        req = service_pb2.CreateCheckpointRequest(
            experiment_id=experiment,
            epoch=epoch,
            blobs=[
                service_pb2.BlobMetadata(
                    checksum="", path=path, blob_type=service_pb2.CHECKPOINT,
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

    def _artifact_iterator(self, parent, path):
        checksum = ""
        with open(path, "rb") as f:
            checksum = md5(f.read()).hexdigest()

        blob_meta = service_pb2.BlobMetadata(
            parent_id=parent,
            checksum=checksum,
            blob_type=service_pb2.CHECKPOINT,
            path=path,
        )
        yield service_pb2.UploadBlobRequest(metadata=blob_meta)
        with open(path, "rb") as f:
            while True:
                data = f.read(CHUNK_SZ)
                if not data:
                    break
                yield service_pb2.UploadBlobRequest(chunks=data)

    def upload_artifact(self, parent: str, path: str) -> UploadArtifactResponse:
        itr = self._artifact_iterator(parent, path)
        resp = self._client.UploadBlob(itr)
        return UploadArtifactResponse(id=resp.blob_id)

    def download_artifact(self, id: str, path: str) -> DownloadArtifactResponse:
        req = service_pb2.DownloadBlobRequest(blob_id=id)
        resp_itr = self._client.DownloadBlob(req)
        ret = DownloadArtifactResponse
        with open(path, 'wb') as f:
            for resp in resp_itr:
                if resp.HasField("chunks"):
                    f.write(resp.chunks)
                if resp.HasField("metadata"):
                    ret.id = resp.metadata.id
                    ret.checksum = resp.metadata.checksum
                    ret.path = path
        return ret

    def close(self):
        if self._channel is not None:
            self._channel.close()

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_value, traceback):
        return self.close()
