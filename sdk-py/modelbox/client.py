from typing import Any, Union
import json
from dataclasses import dataclass

import grpc
from . import service_pb2
from . import service_pb2_grpc
from google.protobuf.struct_pb2 import Value
from google.protobuf import json_format
from google.protobuf.timestamp_pb2 import Timestamp

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
        self, name: str, owner: str, namespace: str, external_id: str, framework: int
    ) -> service_pb2.CreateExperimentResponse:
        req = service_pb2.CreateExperimentRequest(
            name=name,
            owner=owner,
            namespace=namespace,
            external_id=external_id,
            framework=framework.to_proto(),
        )
        return self._client.CreateExperiment(req)

    def update_metadata(self, parent_id: str, key:str, value: Any) -> service_pb2.UpdateMetadataResponse:
        json_value = Value()
        json_format.Parse(json.dumps(value), json_value)
        meta = service_pb2.Metadata(metadata={key: json_value})
        req = service_pb2.UpdateMetadataRequest(parent_id=parent_id, metadata=meta)
        return self._client.UpdateMetadata(req)

    def log_metrics(self, parent_id: str, key:str, value: MetricValue) -> service_pb2.LogMetricsResponse:
        req = service_pb2.LogMetricsRequest(
            parent_id=parent_id,
            key=key,
            value=service_pb2.MetricsValue(
                step=value.step, wallclock_time=value.wallclock_time, f_val=value.value
            ),
        )
        return self._client.LogMetrics(req)

    def close(self):
        if self._channel is not None:
            self._channel.close()

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_value, traceback):
        return self.close()
