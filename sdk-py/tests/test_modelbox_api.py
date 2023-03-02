import unittest
import grpc
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
    Event,
    EventSource,
    LocalFile,
)
from modelbox import service_pb2_grpc
from modelbox import service_pb2
import json

SERVER_ADDR = "localhost:8085"

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

    def GetExperiment(self, request, context):
        experiment_resp = service_pb2.Experiment(
            id=self._fake.uuid4(), name="yolo-test", owner=TEST_OWNER
        )
        return service_pb2.GetExperimentResponse(experiment=experiment_resp)

    def CreateModelVersion(self, request, context):
        model_version = service_pb2.CreateModelVersionResponse(
            model_version=self._fake.uuid4()
        )
        return model_version

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
        return service_pb2.UploadFileResponse(
            file_id=self._fake.uuid4(), artifact_id=self._fake.uuid4()
        )

    def DownloadFile(self, request, context):
        meta = service_pb2.FileMetadata(
            id=self._fake.uuid4(),
            parent_id=self._fake.uuid4(),
            file_type=service_pb2.CHECKPOINT,
            checksum=self._fake.uuid4(),
            src_path="foo/bar",
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
        return service_pb2.UpdateMetadataResponse()

    def ListMetadata(self, request, context):
        payload = Value()
        value = json.dumps({"key": "value"})
        metadata = service_pb2.Metadata(metadata={"/tmp": value})
        return service_pb2.ListMetadataResponse(metadata=metadata)

    def TrackArtifacts(self, request, context):
        return service_pb2.TrackArtifactsResponse(id=self._fake.uuid4())

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
            )
        )
        resp = service_pb2.ListModelsResponse(models=models)
        return resp

    def LogMetrics(self, request, context):
        return service_pb2.LogMetricsResponse()

    def GetMetrics(self, request, context):
        values = [service_pb2.MetricsValue(step=1, wallclock_time=500, f_val=0.45)]
        m = service_pb2.Metrics(key="foo", values=values)
        return service_pb2.GetMetricsResponse(metrics={"foo": m})

    def LogEvent(self, request, context):
        return service_pb2.LogEventResponse(
            created_at=timestamp_pb2.Timestamp(seconds=12345)
        )

    def ListEvents(self, request, context):
        events = service_pb2.ListEventsResponse(
            events=[
                service_pb2.Event(
                    name="checkpoint_write_start",
                    source=service_pb2.EventSource(name="host1"),
                ),
                service_pb2.Event(
                    name="checkpoint_write_end",
                    source=service_pb2.EventSource(name="host1"),
                ),
            ]
        )
        return events

    def ListArtifacts(self, request, context):
        files = [
            service_pb2.FileMetadata(
                id=self._fake.uuid4(),
                parent_id=self._fake.uuid4(),
                file_type=service_pb2.CHECKPOINT,
                checksum=self._fake.uuid4(),
                src_path="foo/bar",
            ),
            service_pb2.FileMetadata(
                id=self._fake.uuid4(),
                parent_id=self._fake.uuid4(),
                file_type=service_pb2.CHECKPOINT,
                checksum=self._fake.uuid4(),
                src_path="foo/lol",
            ),
        ]
        artifacts = [
            service_pb2.Artifact(
                id=self._fake.uuid4(),
                name="foo_astifact",
                object_id=self._fake.uuid4(),
                files=files,
            )
        ]
        return service_pb2.ListArtifactsResponse(artifacts=artifacts)


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

    def test_get_experiment(self):
        experiment = self.mbox.experiment("foo")
        self.assertNotEqual(experiment.id, "")

    def test_create_model_version(self):
        model = self._create_model()
        model_version = model.new_model_version(
            "v1",
            model.name,
            "mv_description",
            [],
            ["prod"],
            MLFramework.PYTORCH,
        )
        self.assertNotEqual(model_version.id, "")

    def test_list_experiments(self):
        resp = self.mbox.experiments("langtech")
        self.assertEqual(2, len(resp.experiments))

    def test_create_model(self):
        model = self.mbox.new_model(
            name="asr_en",
            owner="owner@owner.org",
            namespace="langtech",
            task="asr",
            description="ASR for english",
        )
        self.assertNotEqual("", model.id)

    def test_upload_file(self):
        model = self._create_model()
        file_path = str(
            pathlib.Path(__file__).parent.resolve().joinpath("test_artifact.txt")
        )
        try:
            model.upload_file(
                "checkpoint1", LocalFile.from_path('tests/test_artifact.txt')
            )
        except Exception as e:
            self.fail(e)

    def test_download_artifact(self):
        model = self._create_model()
        assets = model.artifacts
        try:
            assets[0].download("/tmp/lol")
        except Exception as e:
            self.fail(e)

    def test_track_file(self):
        model = self._create_model()
        file_path = str(
            pathlib.Path(__file__).parent.resolve().joinpath("test_artifact.txt")
        )
        try:
            resp = model.track_file(artifact_name="artifactX", f=LocalFile.from_path(file_path))
        except Exception as e:
            self.fail(e)

    def test_list_artifacts(self):
        model = self._create_model()
        artifact_list = model.artifacts
        self.assertEqual(1, len(artifact_list))
        self.assertEqual(2, len(artifact_list[0].assets))

    def test_list_models(self):
        resp = self.mbox.models("langtech")
        self.assertEqual(1, len(resp.models))

    def test_log_metrics(self):
        experiment = self.mbox.new_experiment(
            "yolo", TEST_OWNER, TEST_NAMESPACE, TEST_EXTERN_ID, MLFramework.PYTORCH
        )
        resp = experiment.log_metrics(
            metrics={"val_accu": 0.234}, step=1, wallclock=500
        )

    def test_get_metrics(self):
        resp = self._create_model().all_metrics()

    def test_metadata(self):
        try:
            resp = self._create_model().update_metadata("foo", "bar")
        except Exception as e:
            self.fail(e)

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

    def test_list_events(self):
        experiment = self._create_experiment()
        events = experiment.events()
        self.assertEqual(2, len(events))


    def test_file(self):
        try:
            f = LocalFile.from_path('./tests/test_artifact.txt')
        except Exception as ex:
            self.fail(ex)

    def _create_model(self):
        model = self.mbox.new_model(
            name="asr_en",
            owner="owner@owner.org",
            namespace="langtech",
            task="asr",
            description="ASR for english",
        )
        model.update_metadata("x", "y")
        return model

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
