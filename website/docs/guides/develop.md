# Developing ModelBox


## Install Python SDK

The Python SDK can be installed in the local environment using `pip`. Run the following command from the root directory.

```
make install-sdk-py
```

We automatically install the python lib in the jupyter container.

## Run Tests

The tests require the dependencies of the server to be started first. 

1. Start the dependencies for tests
```
docker compose --profile unittests up
```

2. Run the server tests
```
make test-server
```

The SDK tests only the interfaces to the server and other logic internal to the library, so a full-blown server is not required. We create an ephemeral gRPC server before running the tests. The following commands will run the SDK tests.

```
cd sdk-py
python tests/test_modelbox_api.py
```

## Push latest snapshot containers of ModelBox to DockerHub

1. Build the server and the docker containers.
```
goreleaser release --snapshot --rm-dist
```

2. Push the images to dockerhub. *Replace the versions*
```
docker push modelboxdotio/modelbox:0.0.1-next-arm64v8
docker push modelboxdotio/modelbox:0.0.1-next-amd64
```

3. Create a manifest to link the containers of various arch under one tag
```
docker manifest create modelboxdotio/modelbox:latest modelboxdotio/modelbox:0.0.1-next-arm64v8 modelboxdotio/modelbox:0.0.1-next-amd64
```