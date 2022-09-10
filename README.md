<p align="center"><img src="https://raw.githubusercontent.com/diptanu/modelbox/main/docs/images/ModelBox1.png" width="300" height="150"></p>

[![CircleCI](https://dl.circleci.com/status-badge/img/gh/tensorland/modelbox/tree/main.svg?style=svg)](https://dl.circleci.com/status-badge/redirect/gh/tensorland/modelbox/tree/main)

# AI Model Operations and Metadata Management Service

ModelBox is an AI model and experiment metadata management service. It can integrate with ML frameworks and other services to log and mine metadata, events and metrics related to Experiments, Models and Checkpoints. 

It integrates with various datastores and blob stores for metadata, metrics and artifact stores. The service is very extensible, interfaces can be extended to support more storage services and metadata can be exported and watched by other systems to help with compliance, access control, auditing and deployment of models.

## Features
#### Experiment Metadata and Metrics Logging
  - Log hyperparameters, accuracy/loss and other model quality-related metrics during training.
  - Log trainer events such as data-loading and checkpoint operations, epoch start and end times which help with debugging performance issues.
#### Model Management
  - Log metadata associated with a model such as binaries, notebooks and model metrics.
  - Manage lineage of models with experiments and datasets used to train the model.
  - Label models with metadata that are useful for operational purposes such as the environments they are deployed in and privacy sensitivity.
  - Load models and deployment artifacts in inference services directly from ModelBox. 
#### Events 
  - Log events about the system/trainer state during training and models from experiment jobs, workflow systems and other AI/Model operations services.
  - Any changes made to experiment and model metadata, new models logged or deployed are logged as change events in the system automatically. Stream these events from other systems for any external workflows which need to be invoked.
#### SDK
  - SDKs in Python, Go, Rust and C++ to integrate with ML frameworks and inference services.
  - SDK is built on top of gRPC so really easy to extend into other languages or add features not available.
  - Use the SDK in training code or from even notebooks.
#### Reliable and Easy to Use Control Plane
  - Reliability and availability are at the center of the design mission for ModelBox. Features and APIs are designed with reliability in mind. For example, the artifact store implements a streaming API for uploading and downloading artifacts to ensure memory usage is controlled while serving really large files.
  - Operational metrics related to the control plane are available as Prometheus metrics.

#### Extensibility
  - The interfaces for metadata, metrics and artifact storage can be extended to support more storage services.

## Planned Features
- Flexible metadata and model retention policies.
- Add RBAC-based access control for models and checkpoints for compliance.
- Automatic model benchmarking for performance(latency and throughput) on inference targets.
- Infrastructure for model transformation such that custom recipes can be applied to train models for optimizations for on-device or accelerator inference targets.

## Tutorials and Demos
If you would like to jump straight in, we have some notebooks which demonstrate the usage of the Python SDK independently and with Pytorch and Pytorch Lightning.
- [Pytorch SDK Tutorial](tutorials/Tutorial_Python_SDK.ipynb) 
- [Pytorch Lightning Integration](tutorials/Pytorch_Lightning_Integration_Tutorial.ipynb) 
- [Pytorch Tutorial](tutorials/Tutorial_Pytorch.ipynb) * Work In Progress * 

## Concepts and Understanding of ModelBox API
![Model Box Concepts!](docs/images/API_Concepts.png "Model Box API Concepts")

### Namespace
A Namespace is a mechanism to organize related models or models published by a team. They are also used for access control and such to the metadata of uploaded models, invoking benchmarks or other model transformation work. Namespaces are automatically created when a new model or experiment specifies the namespace it wants to be associated with.

### Model
A model is an object to track common metadata and to apply policies on models created by experiments to solve a machine learning task. For example, datasets to evaluate all trained models of a task can be tracked using this object. Users can also add rules around retention policies of trained versions, setting up policies for labeling a trained model if it has better metrics on a dataset, and meets all other criteria.

### Model Version
A model version is a trained model, it includes the model binary, related files that a user wants to track such as dataset file handles, any other metadata and model metrics. Model versions are always related to a Model and all the policies created for a Model are applied to Model Versions.

### Experiment and Checkpoints
Experiments are the abstraction to store metadata and events from training jobs that produce models. Checkpoints from experiments can be automatically ingested and can be a means to get fault-tolerant training using the Python SDK. Users of PyTorch Lightning can also use the included lightning logger for automatically logging experiment metadata and training artifacts.
Some examples of metadata logged from an experiment are hyperparameters, structure and shape of the models, training accuracy, loss and other related metrics, hardware metrics of the trainer nodes, checkpoint binaries and training code with dependencies.

### Metrics
Metrics can be logged for experiments and models. Metrics are key, value pairs, the value being a series of float, tensor(serialized as strings), or even bytes that are logged over time. Every metric log can have a step and wallclock attribute associated with them which makes them useful in tracking things like accuracy during training or hardware metrics. 
Model Metrics can be expressed as simple key/value pairs.

### Artifacts
Artifacts such as datasets and trained models or checkpoints can be either uploaded to ModelBox or if they are stored externally they can be tracked as metadata attached to experiments and models objects.

### Discrete Events
Events are generated by external systems running the training experiments, inference engines consuming the Models, or even other ML Operations services consuming the models or metadata to benchmark or deploy a model. Events are useful for debugging or operability of models or training platforms.

For example, if events are logged at the start of an epoch, before and after writing a checkpoint, looking at the timestamps allows an engineer to understand which operation is taking too much time if training slows down.

### Change Events
Change events are automatically generated by ModelBox when any metadata about an experiment or model is updated, created or destroyed. Change Events are also logged automatically when artifacts are uploaded or downloaded. They are useful in production systems to know when and where models are consumed, when new models are created by experiments, etc.

## Architecture
ModelBox has the following components
- Metadata Server
- Blob Server
- CLI
- Client Libraries 

### Metadata Server
Meta Data Server is responsible for tracking metadata around models which are created by the training frameworks or users who are uploading trained models and other training artifacts. The Meta Data server exposes a GRPC endpoint for clients to communicate with the server. Supported databases for the metadata service are MySQL, PostgreSQL and an ephemeral storage engine. Additional datastore support can be very easily added by implementing the metadata storage interface.

### Blob Serving Capabilities
ModelBox provides APIs for clients to upload training artifacts and download models in a streaming fashion. The blob serving capability is stateless and hence they can be scaled based on your serving needs.
Currently, they are part of the ModelStore gRPC service, but the plan is to separate the control plane and blob serving APIs so that ModelBox can be run in a blob server mode. This will allow isolating the artifact streaming capability to a separate cluster for performance.

### CLI
The ModelBox CLI provides an interface to interact with the metadata and artifact storage APIs.

### SDK and Client Libraries
The SDK/client libraries are meant for integration with Deep Learning and ML frameworks to integrate ModelBox with the experiment code which creates the model and other training artifacts. The libraries can also be used with applications or control planes that want to use ModelBox in a larger in-house Machine Learning platform.

### High-Level Architecture
![Model Box High-Level Architecture!](docs/images/ModelBox_HighLevel.png "Model Box High-Level Architecture")


## Configuration
ModelBox Server and CLI are configured by TOML files and the configuration can be generated by the CLI. Please refer to the comments on the config and the documentation below to understand what the attributes of the configuration does.

```
modelbox server init-config
```

### Server Configuration
- `listen_addr`: The interface and port on which the server will be listening for Client RPC Requests.

#### Storage Configuration
- metadata_storage: The name of the storage backend which ModelBox is going to use for storing metadata about models. Possible values -
    - `mysql`
         `Host` Host of the MySQL server.
         `Port` Port of the MySQL server.
         `Password` Password of the server.
    - `integrated`

### CLI Configuration

```
modelbox client init-config
```

- `server_addr`: The address of the Metadata Server

## Deployment

All the components of ModelBox are packaged in a single binary which eases deployment in production. 

## Operation Examples

### Starting the metadata server
Generate the config
```
modelbox server init-config
```

Edit the config and start the server

```
$ modelbox server start --config-path ./path/to/modelbox.toml
```

### CLI Examples
Generate the client config
```
modelbox client init-config
```

Create an experiment. Experiments are usually created programmatically via the ModelBox SDK which integrates with deep learning frameworks.
```
modelbox client experiments create --namespace langtech --owner your@email.com --name wav2vec-lid --framework pytorch
```

List Experiments for a namespace
```
modelbox client experiments list --namespace langtech
```

Create a Model for the experiment. The CLI doesn't support adding metadata and artifacts yet, the Python and Go SDKs are the only options to add metadata
programmatically today.
```
modelbox client models create --name wav2vec --owner diptanuc@gmail.com --task asr --description English ASR --namespace langtech 
```

List Models in a namespace
```
modelbox client list --namespace langtech
```

## Development
Build the ModelBox control plane and CLI locally -
```
go install ./cmd/modelbox/
```
or 
```
go build -o /path/to/binary ./cmd/modelbox/
```
Install the python SDK locally for development -
```
cd client-py
pip install .
```

### Test
Spin up the test dependencies using docker - 
```
docker compose --profile unittests up -d
```

Run the tests for a particular package. For example, the following command runs the storage tests -
```
go test ./server/storage/...
```

Refer to the README inside sdk folders to learn how to run the tests for language-specific SDKs.

## Monitoring
Metrics on the metadata server are exposed by the `/metrics` endpoint and can be collected by a Prometheus collector.
The default port for the endpoint is `:2112` and can be configured in the server config.
