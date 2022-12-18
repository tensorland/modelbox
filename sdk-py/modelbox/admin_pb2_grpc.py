# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
"""Client and server classes corresponding to protobuf-defined services."""
import grpc

import admin_pb2 as admin__pb2


class ModelBoxAdminStub(object):
    """The RPC interface used by the workers
    """

    def __init__(self, channel):
        """Constructor.

        Args:
            channel: A grpc.Channel.
        """
        self.RegisterAgent = channel.unary_unary(
                '/modelbox.ModelBoxAdmin/RegisterAgent',
                request_serializer=admin__pb2.RegisterAgentRequest.SerializeToString,
                response_deserializer=admin__pb2.RegisterAgentResponse.FromString,
                )
        self.Heartbeat = channel.unary_unary(
                '/modelbox.ModelBoxAdmin/Heartbeat',
                request_serializer=admin__pb2.HeartbeatRequest.SerializeToString,
                response_deserializer=admin__pb2.HeartbeatResponse.FromString,
                )
        self.GetRunnableActionInstances = channel.unary_unary(
                '/modelbox.ModelBoxAdmin/GetRunnableActionInstances',
                request_serializer=admin__pb2.GetRunnableActionInstancesRequest.SerializeToString,
                response_deserializer=admin__pb2.GetRunnableActionInstancesResponse.FromString,
                )
        self.UpdateActionStatus = channel.unary_unary(
                '/modelbox.ModelBoxAdmin/UpdateActionStatus',
                request_serializer=admin__pb2.UpdateActionStatusRequest.SerializeToString,
                response_deserializer=admin__pb2.UpdateActionStatusResponse.FromString,
                )


class ModelBoxAdminServicer(object):
    """The RPC interface used by the workers
    """

    def RegisterAgent(self, request, context):
        """Register an agent capable of running plugins
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def Heartbeat(self, request, context):
        """Workers heartbeat with the server about their presence
        and work progress periodically
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def GetRunnableActionInstances(self, request, context):
        """Download the list of work that can be exectuted by a action runner
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def UpdateActionStatus(self, request, context):
        """Update action status
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')


def add_ModelBoxAdminServicer_to_server(servicer, server):
    rpc_method_handlers = {
            'RegisterAgent': grpc.unary_unary_rpc_method_handler(
                    servicer.RegisterAgent,
                    request_deserializer=admin__pb2.RegisterAgentRequest.FromString,
                    response_serializer=admin__pb2.RegisterAgentResponse.SerializeToString,
            ),
            'Heartbeat': grpc.unary_unary_rpc_method_handler(
                    servicer.Heartbeat,
                    request_deserializer=admin__pb2.HeartbeatRequest.FromString,
                    response_serializer=admin__pb2.HeartbeatResponse.SerializeToString,
            ),
            'GetRunnableActionInstances': grpc.unary_unary_rpc_method_handler(
                    servicer.GetRunnableActionInstances,
                    request_deserializer=admin__pb2.GetRunnableActionInstancesRequest.FromString,
                    response_serializer=admin__pb2.GetRunnableActionInstancesResponse.SerializeToString,
            ),
            'UpdateActionStatus': grpc.unary_unary_rpc_method_handler(
                    servicer.UpdateActionStatus,
                    request_deserializer=admin__pb2.UpdateActionStatusRequest.FromString,
                    response_serializer=admin__pb2.UpdateActionStatusResponse.SerializeToString,
            ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
            'modelbox.ModelBoxAdmin', rpc_method_handlers)
    server.add_generic_rpc_handlers((generic_handler,))


 # This class is part of an EXPERIMENTAL API.
class ModelBoxAdmin(object):
    """The RPC interface used by the workers
    """

    @staticmethod
    def RegisterAgent(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/modelbox.ModelBoxAdmin/RegisterAgent',
            admin__pb2.RegisterAgentRequest.SerializeToString,
            admin__pb2.RegisterAgentResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def Heartbeat(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/modelbox.ModelBoxAdmin/Heartbeat',
            admin__pb2.HeartbeatRequest.SerializeToString,
            admin__pb2.HeartbeatResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def GetRunnableActionInstances(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/modelbox.ModelBoxAdmin/GetRunnableActionInstances',
            admin__pb2.GetRunnableActionInstancesRequest.SerializeToString,
            admin__pb2.GetRunnableActionInstancesResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def UpdateActionStatus(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/modelbox.ModelBoxAdmin/UpdateActionStatus',
            admin__pb2.UpdateActionStatusRequest.SerializeToString,
            admin__pb2.UpdateActionStatusResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)
