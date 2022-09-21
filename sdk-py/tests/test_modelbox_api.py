from argparse import Namespace
from importlib.metadata import metadata
from time import time
import unittest
import grpc
import sys
import os
import pathlib

from random import randrange
from faker import Faker
from concurrent import futures

from google.protobuf.struct_pb2 import Value
from google.protobuf import json_format, timestamp_pb2

from modelbox.modelbox import (
    ModelBox,
    MLFramework,
    Artifact,
    ArtifactMime,
    MetricValue,
    Event,
    EventSource,
)
from modelbox import service_pb2_grpc
from modelbox import service_pb2

SERVER_ADDR = "localhost:8085"

TEST_PROJECT_NAME = "langtech"
TEST_OWNER = "owner@gmail.com"
TEST_EXTERN_ID = "xyz"
TEST_NAMESPACE = "ai/langtech/translation"

TEST_MODEL_NAME = "yolo4"


class MockModelStoreServicer(service_pb2_grpc.ModelStoreServicer):
    def __init__(self):
        self._fake = Faker()

    def CreateModel(self, request, context):
        model = service_pb2.CreateModelResponse(id=self._fake.uuid4())
        return model

    def CreateExperiment(self, request, context):
        experiment_resp = service_pb2.CreateExperimentResponse(
            experiment_id=self._fake.uuid4(),
            experiment_exists=True,
        )
        return experiment_resp

    def CreateCheckpoint(self, request, context):
        checkpoint = service_pb2.CreateCheckpointResponse(
            checkpoint_id=self._fake.uuid4()
        )
        return checkpoint

    def CreateModelVersion(self, request, context):
        model_version = service_pb2.CreateModelVersionResponse(
            model_version=self._fake.uuid4()
        )
        return model_version

    def ListCheckpoints(self, request, context):
        c1 = service_pb2.Checkpoint(
            id=self._fake.uuid4(), epoch=23, experiment_id=self._fake.uuid4()
        )
        c2 = service_pb2.Checkpoint(
            id=self._fake.uuid4(), epoch=24, experiment_id=self._fake.uuid4()
        )
        resp = service_pb2.ListCheckpointsResponse(
            checkpoints=[c1, c2],
        )
        return resp

    def ListExperiments(self, request, context):
        e1 = service_pb2.Experiment(
            id=self._fake.uuid4(),
            name="exp1",
            namespace="langtech",
            owner="owner@owner.org",
        )
        e2 = service_pb2.Experiment(
            id=self._fake.uuid4(),
            name="exp2",
            namespace="langtech",
            owner="owner@owner.org",
        )
        resp = service_pb2.ListExperimentsResponse(
            experiments=[e1, e2],
        )
        return resp

    def UploadFile(self, request_iterator, context):
        for req in request_iterator:
            pass
        return service_pb2.UploadFileResponse(file_id=self._fake.uuid4())

    def DownloadFile(self, request, context):
        meta = service_pb2.FileMetadata(
            id=self._fake.uuid4(),
            parent_id=self._fake.uuid4(),
            file_type=service_pb2.CHECKPOINT,
            checksum=self._fake.uuid4(),
            path="foo/bar",
        )
        yield service_pb2.DownloadFileResponse(metadata=meta)
        artifact = str(
            pathlib.Path(__file__).parent.resolve().joinpath("test_artifact.txt")
        )
        with open(artifact, "rb") as f:
            while True:
                data = f.read(1024)
                if not data:
                    break
                yield service_pb2.DownloadFileResponse(chunks=data)

    def UpdateMetadata(self, req, context):
        return service_pb2.UpdateMetadataResponse(updated_at=timestamp_pb2.Timestamp())

    def ListMetadata(self, request, context):
        payload = Value()
        json_format.ParseDict({"key": "value"}, payload)
        return service_pb2.ListMetadataResponse(metadata={"/tmp": payload})

    def TrackArtifacts(self, request, context):
        return service_pb2.TrackArtifactsResponse(num_files_tracked=2)

    def ListModels(self, request, context):
        models = []
        models.append(
            service_pb2.Model(
                id=self._fake.uuid4(),
                name="gpt",
                owner="owner@owner.org",
                namespace="langtech",
                description="long description",
                task="mytask",
                files=[],
            )
        )
        resp = service_pb2.ListModelsResponse(models=models)
        return resp

    def LogMetrics(self, request, context):
        return service_pb2.LogMetricsResponse()

    def GetMetrics(self, request, context):
        values = [service_pb2.MetricsValue(step=1, wallclock_time=500, f_val=0.45)]
        m = service_pb2.Metrics(key="foo", values=values)
        return service_pb2.GetMetricsResponse(metrics=[m])

    def LogEvent(self, request, context):
        return service_pb2.LogEventResponse(
            created_at=timestamp_pb2.Timestamp(seconds=12345)
        )


# We are really testing whether the client actually works against the current version
# of the grpc server definition. Tests related to logic in server based on what the
# client is passing should be in server.
class TestModelBoxApi(unittest.TestCase):
    def setUp(self) -> None:
        self.mbox = ModelBox(SERVER_ADDR)
        return super().setUp()

    def tearDown(self) -> None:
        return super().tearDown()

    def test_create_experiment(self):
        experiment = self.mbox.new_experiment(
            "yolo", TEST_OWNER, TEST_NAMESPACE, TEST_EXTERN_ID, MLFramework.PYTORCH
        )
        self.assertNotEqual(experiment.id, "")

    def test_create_checkpoint(self):
        experiment = self.mbox.new_experiment(
            TEST_MODEL_NAME,
            TEST_OWNER,
            TEST_NAMESPACE,
            TEST_EXTERN_ID,
            MLFramework.PYTORCH,
        )
        metrics = {"val_accu": 97.8, "train_accu": 98.8}
        artifacts = [
            Artifact(
                parent="", path="/path/to/checkpoint", mime_type=ArtifactMime.Checkpoint
            )
        ]
        checkpoint = experiment.new_checkpoint(randrange(10000), metrics)
        self.assertNotEqual(checkpoint.id, "")

    def test_create_model_version(self):
        model = self._create_model()
        model_version = model.new_model_version(
            "v1",
            model.name,
            "mv_description",
            [],
            {},
            ["prod"],
            MLFramework.PYTORCH,
        )
        self.assertNotEqual(model_version.id, "")

    def test_list_checkpoints(self):
        experiment = self._create_experiment()
        checkpoints = experiment.list_checkpoints()
        self.assertEqual(2, len(checkpoints))

    def test_list_experiments(self):
        resp = self.mbox.list_experiments("langtech")
        self.assertEqual(2, len(resp.experiments))

    def test_create_model(self):
        model = self.mbox.new_model(
            name="asr_en",
            owner="owner@owner.org",
            namespace="langtech",
            task="asr",
            description="ASR for english",
            metadata={"x": "y"},
        )
        self.assertNotEqual("", model.id)

    def test_upload_artifact(self):
        model = self._create_model()
        file_path = str(
            pathlib.Path(__file__).parent.resolve().joinpath("test_artifact.txt")
        )
        resp = model.upload_artifact(Artifact(model.id, file_path, ArtifactMime.Text))
        self.assertNotEqual("", resp.id)

    def test_download_artifact(self):
        model = self._create_model()
        resp = model.download_artifact("random_id", "/tmp/lol")
        self.assertNotEqual("", resp.id)

    def test_track_artifacts(self):
        model = self._create_model()
        file_path = str(
            pathlib.Path(__file__).parent.resolve().joinpath("test_artifact.txt")
        )
        file = Artifact(
            parent="parent-id",
            checksum="abc",
            path=file_path,
            mime_type=ArtifactMime.Text,
        )
        resp = model.track_artifacts(artifacts=[file])
        self.assertEqual(2, resp.num_artifacts_tracked)

    def test_list_models(self):
        resp = self.mbox.list_models("langtech")
        self.assertEqual(1, len(resp.models))

    def test_log_metrics(self):
        experiment = self.mbox.new_experiment(
            "yolo", TEST_OWNER, TEST_NAMESPACE, TEST_EXTERN_ID, MLFramework.PYTORCH
        )
        resp = experiment.log_metrics(
            key="val_accu", step=1, wallclock=500, value=0.234
        )

    def test_get_metrics(self):
        resp = self._create_model().get_all_metrics()

    def test_metadata(self):
        resp = self._create_model().update_metadata("foo", "bar")
        self.assertNotEqual(resp.updated_at, 0)

    def test_list_metadata(self):
        resp = self._create_model().metadata()
        self.assertEqual(len(resp.metadata.keys()), 1)

    def test_log_event(self):
        event = Event(
            name="checkpoint_start",
            source=EventSource(name="trainer1"),
            wallclock_time=1234,
            metadata={"key1": "value1"},
        )
        resp = self._create_model().log_event(event)
        self.assertEqual(resp.created_at, 12345)

    def _create_model(self):
        return self.mbox.new_model(
            name="asr_en",
            owner="owner@owner.org",
            namespace="langtech",
            task="asr",
            description="ASR for english",
            metadata={"x": "y"},
        )

    def _create_experiment(self):
        return self.mbox.new_experiment(
            TEST_MODEL_NAME,
            TEST_OWNER,
            TEST_NAMESPACE,
            TEST_EXTERN_ID,
            MLFramework.PYTORCH,
        )


if __name__ == "__main__":
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    service_pb2_grpc.add_ModelStoreServicer_to_server(MockModelStoreServicer(), server)
    server.add_insecure_port(SERVER_ADDR)
    server.start()
    unittest.main()
    server.stop()
