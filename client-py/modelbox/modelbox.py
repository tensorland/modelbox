from datetime import datetime
from re import I, S
from time import time
from typing import Dict, List
from typing_extensions import Self
from enum import Enum
from dataclasses import dataclass

import grpc
from grpc_status import rpc_status
from numpy import float32, uint, uint64
from regex import W
from . import service_pb2
from . import service_pb2_grpc

DEFAULT_NAMESPACE = "default"


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


class Model:
    def __init__(
        self,
        id: str,
        name: str,
        owner: str,
        namespace: str,
        task: str,
        description: str,
        metadata: Dict,
    ) -> Self:
        self._id = id
        self._name = name
        self._owner = owner
        self._namespace = namespace
        self._task = task
        self._description = description
        self._metadata = metadata

    @property
    def id(self):
        return self._id

    @property
    def name(self):
        return self._name

    @property
    def owner(self):
        return self._owner

    @property
    def namespace(self):
        return self._namespace

    @property
    def task(self):
        return self._task

    @property
    def description(self):
        return self._description

    @property
    def metadata(self):
        return self._metadata


class ModelVersion:
    def __init__(
        self,
        id: str,
        model_id: str,
        name: str,
        version: str,
        description: str,
        blobs: List,
        metadata: Dict,
        unique_tags: List,
        framework: MLFramework,
    ) -> Self:
        self._id = id
        self._model_id = model_id
        self._name = name
        self._version = version
        self._description = description
        self._blobs = blobs
        self._metadata = metadata
        self._framework = framework
        self._unique_tags = unique_tags

    @property
    def id(self):
        return self._id

    @property
    def model_id(self):
        return self._model_id

    @property
    def name(self):
        return self._name

    @property
    def version(self):
        return self._version

    @property
    def description(self):
        return self._description

    @property
    def blobs(self):
        return self._blobs

    @property
    def metadata(self):
        return self._metadata

    @property
    def framework(self):
        return self._framework

    @property
    def unique_tags(self):
        return self._unique_tags


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

    def upload_artifact(self, parent: str, path: str) -> UploadArtifactResponse:
        pass

    def close(self):
        if self._channel is not None:
            self._channel.close()

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_value, traceback):
        return self.close()
