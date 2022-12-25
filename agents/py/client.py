import grpc
from modelbox import admin_pb2
from modelbox import admin_pb2_grpc
from google.protobuf import timestamp_pb2


class AdminClient:

    def __init__(self, addr):
        self._addr = addr
        self._channel = grpc.insecure_channel(addr)
        self._client = admin_pb2_grpc.ModelBoxAdminStub(self._channel)

    def register_agent(self, node_info: admin_pb2.NodeInfo, name:str) -> admin_pb2.RegisterAgentRequest:
        req = admin_pb2.RegisterAgentRequest(node_info=node_info, agent_name=name)
        return self._client.RegisterAgent(req)

    def heartbeat(self, node_id: str) -> admin_pb2.HeartbeatResponse:
        ts = timestamp_pb2.Timestamp()
        req = admin_pb2.HeartbeatRequest(node_id=node_id, at=ts.GetCurrentTime())
        return self._client.Heartbeat(req)

    def get_runnable_actions(self) -> admin_pb2.GetRunnableActionInstancesResponse:
        pass

    def update_action_status(self) -> admin_pb2.UpdateActionStatusResponse:
        pass

    def close(self):
        if self._channel is not None:
            self._channel.close()

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_value, traceback):
        return self.close()