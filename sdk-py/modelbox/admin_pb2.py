# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: admin.proto
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google.protobuf import timestamp_pb2 as google_dot_protobuf_dot_timestamp__pb2
from google.protobuf import struct_pb2 as google_dot_protobuf_dot_struct__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x0b\x61\x64min.proto\x12\x08modelbox\x1a\x1fgoogle/protobuf/timestamp.proto\x1a\x1cgoogle/protobuf/struct.proto\"<\n\x08NodeInfo\x12\x11\n\thost_name\x18\x01 \x01(\t\x12\x0f\n\x07ip_addr\x18\x02 \x01(\t\x12\x0c\n\x04\x61rch\x18\x03 \x01(\t\"K\n\x10HeartbeatRequest\x12\x0f\n\x07node_id\x18\x01 \x01(\t\x12&\n\x02\x61t\x18\x14 \x01(\x0b\x32\x1a.google.protobuf.Timestamp\"\x13\n\x11HeartbeatResponse\"`\n\x15SubscribeEventRequest\x12\x11\n\tnamespace\x18\x01 \x01(\t\x12\x14\n\x0cml_framework\x18\x02 \x01(\t\x12\r\n\x05owner\x18\x03 \x01(\t\x12\x0f\n\x07\x61\x63tions\x18\x04 \x03(\t\"Q\n\x14RegisterAgentRequest\x12%\n\tnode_info\x18\x01 \x01(\x0b\x32\x12.modelbox.NodeInfo\x12\x12\n\nagent_name\x18\x02 \x01(\t\"(\n\x15RegisterAgentResponse\x12\x0f\n\x07node_id\x18\x01 \x01(\t\"F\n!GetRunnableActionInstancesRequest\x12\x13\n\x0b\x61\x63tion_name\x18\x01 \x01(\t\x12\x0c\n\x04\x61rch\x18\x02 \x01(\t\"\xbd\x01\n\x0eRunnableAction\x12\n\n\x02id\x18\x01 \x01(\t\x12\x11\n\taction_id\x18\x02 \x01(\t\x12\x0f\n\x07\x63ommand\x18\x03 \x01(\t\x12\x34\n\x06params\x18\x05 \x03(\x0b\x32$.modelbox.RunnableAction.ParamsEntry\x1a\x45\n\x0bParamsEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12%\n\x05value\x18\x02 \x01(\x0b\x32\x16.google.protobuf.Value:\x02\x38\x01\"Q\n\"GetRunnableActionInstancesResponse\x12+\n\tinstances\x18\x01 \x03(\x0b\x32\x18.modelbox.RunnableAction\"\x11\n\x0fGetWorkResponse\"\x85\x01\n\x19UpdateActionStatusRequest\x12\x1a\n\x12\x61\x63tion_instance_id\x18\x01 \x01(\t\x12\x0e\n\x06status\x18\x02 \x01(\r\x12\x0f\n\x07outcome\x18\x03 \x01(\r\x12\x16\n\x0eoutcome_reason\x18\x04 \x01(\t\x12\x13\n\x0budpate_time\x18\x05 \x01(\x04\"\x1c\n\x1aUpdateActionStatusResponse2\x81\x03\n\rModelBoxAdmin\x12P\n\rRegisterAgent\x12\x1e.modelbox.RegisterAgentRequest\x1a\x1f.modelbox.RegisterAgentResponse\x12\x44\n\tHeartbeat\x12\x1a.modelbox.HeartbeatRequest\x1a\x1b.modelbox.HeartbeatResponse\x12w\n\x1aGetRunnableActionInstances\x12+.modelbox.GetRunnableActionInstancesRequest\x1a,.modelbox.GetRunnableActionInstancesResponse\x12_\n\x12UpdateActionStatus\x12#.modelbox.UpdateActionStatusRequest\x1a$.modelbox.UpdateActionStatusResponseB-Z+github.com/tensorland/modelbox/sdk-go/protob\x06proto3')



_NODEINFO = DESCRIPTOR.message_types_by_name['NodeInfo']
_HEARTBEATREQUEST = DESCRIPTOR.message_types_by_name['HeartbeatRequest']
_HEARTBEATRESPONSE = DESCRIPTOR.message_types_by_name['HeartbeatResponse']
_SUBSCRIBEEVENTREQUEST = DESCRIPTOR.message_types_by_name['SubscribeEventRequest']
_REGISTERAGENTREQUEST = DESCRIPTOR.message_types_by_name['RegisterAgentRequest']
_REGISTERAGENTRESPONSE = DESCRIPTOR.message_types_by_name['RegisterAgentResponse']
_GETRUNNABLEACTIONINSTANCESREQUEST = DESCRIPTOR.message_types_by_name['GetRunnableActionInstancesRequest']
_RUNNABLEACTION = DESCRIPTOR.message_types_by_name['RunnableAction']
_RUNNABLEACTION_PARAMSENTRY = _RUNNABLEACTION.nested_types_by_name['ParamsEntry']
_GETRUNNABLEACTIONINSTANCESRESPONSE = DESCRIPTOR.message_types_by_name['GetRunnableActionInstancesResponse']
_GETWORKRESPONSE = DESCRIPTOR.message_types_by_name['GetWorkResponse']
_UPDATEACTIONSTATUSREQUEST = DESCRIPTOR.message_types_by_name['UpdateActionStatusRequest']
_UPDATEACTIONSTATUSRESPONSE = DESCRIPTOR.message_types_by_name['UpdateActionStatusResponse']
NodeInfo = _reflection.GeneratedProtocolMessageType('NodeInfo', (_message.Message,), {
  'DESCRIPTOR' : _NODEINFO,
  '__module__' : 'admin_pb2'
  # @@protoc_insertion_point(class_scope:modelbox.NodeInfo)
  })
_sym_db.RegisterMessage(NodeInfo)

HeartbeatRequest = _reflection.GeneratedProtocolMessageType('HeartbeatRequest', (_message.Message,), {
  'DESCRIPTOR' : _HEARTBEATREQUEST,
  '__module__' : 'admin_pb2'
  # @@protoc_insertion_point(class_scope:modelbox.HeartbeatRequest)
  })
_sym_db.RegisterMessage(HeartbeatRequest)

HeartbeatResponse = _reflection.GeneratedProtocolMessageType('HeartbeatResponse', (_message.Message,), {
  'DESCRIPTOR' : _HEARTBEATRESPONSE,
  '__module__' : 'admin_pb2'
  # @@protoc_insertion_point(class_scope:modelbox.HeartbeatResponse)
  })
_sym_db.RegisterMessage(HeartbeatResponse)

SubscribeEventRequest = _reflection.GeneratedProtocolMessageType('SubscribeEventRequest', (_message.Message,), {
  'DESCRIPTOR' : _SUBSCRIBEEVENTREQUEST,
  '__module__' : 'admin_pb2'
  # @@protoc_insertion_point(class_scope:modelbox.SubscribeEventRequest)
  })
_sym_db.RegisterMessage(SubscribeEventRequest)

RegisterAgentRequest = _reflection.GeneratedProtocolMessageType('RegisterAgentRequest', (_message.Message,), {
  'DESCRIPTOR' : _REGISTERAGENTREQUEST,
  '__module__' : 'admin_pb2'
  # @@protoc_insertion_point(class_scope:modelbox.RegisterAgentRequest)
  })
_sym_db.RegisterMessage(RegisterAgentRequest)

RegisterAgentResponse = _reflection.GeneratedProtocolMessageType('RegisterAgentResponse', (_message.Message,), {
  'DESCRIPTOR' : _REGISTERAGENTRESPONSE,
  '__module__' : 'admin_pb2'
  # @@protoc_insertion_point(class_scope:modelbox.RegisterAgentResponse)
  })
_sym_db.RegisterMessage(RegisterAgentResponse)

GetRunnableActionInstancesRequest = _reflection.GeneratedProtocolMessageType('GetRunnableActionInstancesRequest', (_message.Message,), {
  'DESCRIPTOR' : _GETRUNNABLEACTIONINSTANCESREQUEST,
  '__module__' : 'admin_pb2'
  # @@protoc_insertion_point(class_scope:modelbox.GetRunnableActionInstancesRequest)
  })
_sym_db.RegisterMessage(GetRunnableActionInstancesRequest)

RunnableAction = _reflection.GeneratedProtocolMessageType('RunnableAction', (_message.Message,), {

  'ParamsEntry' : _reflection.GeneratedProtocolMessageType('ParamsEntry', (_message.Message,), {
    'DESCRIPTOR' : _RUNNABLEACTION_PARAMSENTRY,
    '__module__' : 'admin_pb2'
    # @@protoc_insertion_point(class_scope:modelbox.RunnableAction.ParamsEntry)
    })
  ,
  'DESCRIPTOR' : _RUNNABLEACTION,
  '__module__' : 'admin_pb2'
  # @@protoc_insertion_point(class_scope:modelbox.RunnableAction)
  })
_sym_db.RegisterMessage(RunnableAction)
_sym_db.RegisterMessage(RunnableAction.ParamsEntry)

GetRunnableActionInstancesResponse = _reflection.GeneratedProtocolMessageType('GetRunnableActionInstancesResponse', (_message.Message,), {
  'DESCRIPTOR' : _GETRUNNABLEACTIONINSTANCESRESPONSE,
  '__module__' : 'admin_pb2'
  # @@protoc_insertion_point(class_scope:modelbox.GetRunnableActionInstancesResponse)
  })
_sym_db.RegisterMessage(GetRunnableActionInstancesResponse)

GetWorkResponse = _reflection.GeneratedProtocolMessageType('GetWorkResponse', (_message.Message,), {
  'DESCRIPTOR' : _GETWORKRESPONSE,
  '__module__' : 'admin_pb2'
  # @@protoc_insertion_point(class_scope:modelbox.GetWorkResponse)
  })
_sym_db.RegisterMessage(GetWorkResponse)

UpdateActionStatusRequest = _reflection.GeneratedProtocolMessageType('UpdateActionStatusRequest', (_message.Message,), {
  'DESCRIPTOR' : _UPDATEACTIONSTATUSREQUEST,
  '__module__' : 'admin_pb2'
  # @@protoc_insertion_point(class_scope:modelbox.UpdateActionStatusRequest)
  })
_sym_db.RegisterMessage(UpdateActionStatusRequest)

UpdateActionStatusResponse = _reflection.GeneratedProtocolMessageType('UpdateActionStatusResponse', (_message.Message,), {
  'DESCRIPTOR' : _UPDATEACTIONSTATUSRESPONSE,
  '__module__' : 'admin_pb2'
  # @@protoc_insertion_point(class_scope:modelbox.UpdateActionStatusResponse)
  })
_sym_db.RegisterMessage(UpdateActionStatusResponse)

_MODELBOXADMIN = DESCRIPTOR.services_by_name['ModelBoxAdmin']
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'Z+github.com/tensorland/modelbox/sdk-go/proto'
  _RUNNABLEACTION_PARAMSENTRY._options = None
  _RUNNABLEACTION_PARAMSENTRY._serialized_options = b'8\001'
  _NODEINFO._serialized_start=88
  _NODEINFO._serialized_end=148
  _HEARTBEATREQUEST._serialized_start=150
  _HEARTBEATREQUEST._serialized_end=225
  _HEARTBEATRESPONSE._serialized_start=227
  _HEARTBEATRESPONSE._serialized_end=246
  _SUBSCRIBEEVENTREQUEST._serialized_start=248
  _SUBSCRIBEEVENTREQUEST._serialized_end=344
  _REGISTERAGENTREQUEST._serialized_start=346
  _REGISTERAGENTREQUEST._serialized_end=427
  _REGISTERAGENTRESPONSE._serialized_start=429
  _REGISTERAGENTRESPONSE._serialized_end=469
  _GETRUNNABLEACTIONINSTANCESREQUEST._serialized_start=471
  _GETRUNNABLEACTIONINSTANCESREQUEST._serialized_end=541
  _RUNNABLEACTION._serialized_start=544
  _RUNNABLEACTION._serialized_end=733
  _RUNNABLEACTION_PARAMSENTRY._serialized_start=664
  _RUNNABLEACTION_PARAMSENTRY._serialized_end=733
  _GETRUNNABLEACTIONINSTANCESRESPONSE._serialized_start=735
  _GETRUNNABLEACTIONINSTANCESRESPONSE._serialized_end=816
  _GETWORKRESPONSE._serialized_start=818
  _GETWORKRESPONSE._serialized_end=835
  _UPDATEACTIONSTATUSREQUEST._serialized_start=838
  _UPDATEACTIONSTATUSREQUEST._serialized_end=971
  _UPDATEACTIONSTATUSRESPONSE._serialized_start=973
  _UPDATEACTIONSTATUSRESPONSE._serialized_end=1001
  _MODELBOXADMIN._serialized_start=1004
  _MODELBOXADMIN._serialized_end=1389
# @@protoc_insertion_point(module_scope)
