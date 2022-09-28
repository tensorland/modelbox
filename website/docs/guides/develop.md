# Developing ModelBox

We cover how to contribute to ModelBox using GitPod and some miscellaneous topics related to the development of the service and SDK.

## Gitpod

Gitpod provides an ephemeral development environment which is ideal for -
1. Getting started with ModelBox and evaluate the service if you don't have Docker available locally.
2. Contributing code to ModelBox without setting up the development environment locally.

Click the following button to have a GitPod workspace. When the workspace starts, it will automatically bring up a docker environment in a terminal and the PyTorch notebook can be run inside the workspace.

[![Open in Gitpod](https://gitpod.io/button/open-in-gitpod.svg)](https://gitpod.io/#https://github.com/tensorland/modelbox)


## Install Python SDK

The Python SDK can be installed in the local environment using `pip`. Run the following command from the root directory.

```
make install-sdk-py
```

We automatically install the python lib in the jupyter container.

## Build Python SDK and push to PyPi

### Create a source distribution of modelbox
```
make build-sdk-py
```

### Upload to test pypi
```
make upload-sdk-py-test
```


### Upload to PyPi
```
make upload-sdk-py
```


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

4. Push the manifest
```
docker manifest push --purge modelboxdotio/modelbox:latest
```