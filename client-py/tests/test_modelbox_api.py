from argparse import Namespace
import unittest
import grpc
import sys
import os
import pathlib

from random import randrange
from faker import Faker
from concurrent import futures

from modelbox.modelbox import ModelBoxClient, MLFramework
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
            experiment_id=self._fake.uuid4(), experiment_exists=True,
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
        resp = service_pb2.ListCheckpointsResponse(checkpoints=[c1, c2],)
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
        resp = service_pb2.ListExperimentsResponse(experiments=[e1, e2],)
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
            return service_pb2.UpdateMetadataResponse(updated_at=1260)


# We are really testing whether the client actually works against the current version
# of the grpc server definition. Tests related to logic in server based on what the
# client is passing should be in server.
class TestModelBoxApi(unittest.TestCase):
    def setUp(self) -> None:
        self._client = ModelBoxClient(SERVER_ADDR)
        return super().setUp()

    def tearDown(self) -> None:
        self._client.close()
        return super().tearDown()

    def test_create_experiment(self):
        result = self._client.create_experiment(
            "yolo", TEST_OWNER, TEST_NAMESPACE, TEST_EXTERN_ID, MLFramework.PYTORCH
        )
        self.assertNotEqual(result.experiment_id, "")

    def test_create_checkpoint(self):
        result = self._client.create_experiment(
            TEST_MODEL_NAME,
            TEST_OWNER,
            TEST_NAMESPACE,
            TEST_EXTERN_ID,
            MLFramework.PYTORCH,
        )
        metrics = {"val_accu": 97.8, "train_accu": 98.8}
        checkpoint_id = self._client.create_checkpoint(
            result.experiment_id, randrange(10000), "/path/to/checkpoint", metrics
        )
        self.assertNotEqual(checkpoint_id, "")

    def test_create_model_version(self):
        model = self._client.create_model(
            TEST_MODEL_NAME,
            TEST_OWNER,
            TEST_NAMESPACE,
            "object_detection",
            "yolo_des",
            {"meta": "foo"},
        )
        model_version = self._client.create_model_version(
            model.id,
            "yolo4_v1",
            "v1",
            "mv_description",
            [],
            {},
            service_pb2.PYTORCH,
            ["prod"],
        )
        self.assertNotEqual(model_version.id, "")

    def test_list_checkpoints(self):
        resp = self._client.list_checkpoints("exp1")
        self.assertEqual(2, len(resp.checkpoints))

    def test_list_experiments(self):
        resp = self._client.list_experiments("langtech")
        self.assertEqual(2, len(resp.experiments))

    def test_create_model(self):
        resp = resp = self._client.create_model(
            name="asr_en",
            owner="owner@owner.org",
            namespace="langtech",
            task="asr",
            description="ASR for english",
            metadata={"x": "y"},
        )
        self.assertNotEqual("", resp.id)

    def test_create_model_version(self):
        tags = ["test"]
        resp = self._client.create_model_version(
            model_id="abc",
            name="asr_en_july",
            version="1",
            description="ASR for english",
            metadata={"x": "y"},
            unique_tags=tags,
            files=[],
            framework=MLFramework.PYTORCH.to_proto(),
        )
        self.assertNotEqual("", resp.id)

    def test_upload_file(self):
        file_path= str(
            pathlib.Path(__file__).parent.resolve().joinpath("test_artifact.txt")
        )
        resp = self._client.upload_file("abc", file_path)
        self.assertNotEqual("", resp.id)

    def test_download_artifact(self):
        resp = self._client.download_artifact("random_id", "/tmp/lol")
        self.assertNotEqual("", resp.id)


if __name__ == "__main__":
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    service_pb2_grpc.add_ModelStoreServicer_to_server(MockModelStoreServicer(), server)
    server.add_insecure_port(SERVER_ADDR)
    server.start()
    unittest.main()
    server.stop()
