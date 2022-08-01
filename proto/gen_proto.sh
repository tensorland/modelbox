#!/bin/bash

protoc --go_out=client --go_opt=paths=source_relative --go-grpc_out=client --go-grpc_opt=paths=source_relative proto/service.proto

python -m grpc_tools.protoc -Iproto/ --python_out=client-py/modelbox/ --grpc_python_out=client-py/modelbox/ proto/service.proto