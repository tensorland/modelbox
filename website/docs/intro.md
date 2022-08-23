---
sidebar_position: 1
---

# Introduction

ModelBox is an AI model and experiment metadata management service. It can integrate with ML frameworks and other services to log and mine metadata, events and metrics related to Experiments, Models and Checkpoints.

It integrates with various datastores and blob stores for metadata, metrics and artifact stores. The service is very extensible, interfaces can be extended to support more storage services and metadata can be exported and watched by other systems to help with compliance, access control, auditing and deployment of models.

## Features
#### Experiment Metadata and Metrics Logging
- Log hyperparameters, accuracy/loss and other quality-related metrics during training.
- Log trainer events such as data-loading and checkpoint operations, epoch start and end times which help with debugging performance issues.

#### Model Management

- Log metadata associated with a model such as binaries, notebooks, model metrics, etc.
- Manage lineage of models with experiments, and datasets used to train the model.
- Label models with metadata that are useful for operational purposes such as the environments they are deployed in, privacy sensitivity, etc.
- Load models and deployment artifacts in inference services directly from ModelBox.

#### Events

- Log events about the system/trainer state during training and models from experiment jobs, workflow systems and other AI/Model operations services.
- Any changes made to experiment and model metadata, new models logged or deployed are logged as change events in the system automatically. Stream these events from other systems for any external workflows which need to be invoked.

#### SDK

- SDKs in Python, Go, Rust and C++ to integrate with ML frameworks and inference services.
- SDK is built on top of gRPC so really easy to extend into other languages or add features not available.
- Use the SDK in training code or from even notebooks.

#### Reliable and Easy to Use Control Plane

- Reliability and availability are at the center of the design mission for ModelBox. Features and APIs are designed with reliability in mind. For example, the artifact store implements a streaming API for upload and download APIs to ensure memory usage is controlled while serving really large files.
- Metrics related to the control plane - API latency, system resource usage, etc, are all available as Prometheus metrics.

#### Extensibility

- Hackable and Interface first design 
- The service can be easily extended to support newer datastores and services for metrics, metadata and artifact storage.
