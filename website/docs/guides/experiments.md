---
sidebar_position: 3
---

# Logging Experiment Metadata

Log experiment metadata like hyperparameters, discrete operational events from the trainer, model metrics after every epoch, and even hardware metrics where training is being run.


## Tutorials
Several tutorials go over logging metadata -
1. Python SDK Tutorial
2. PyTorch Lightning Tutorial


## Python SDK 

#### Create an Experiment
```
experiment = mbox.new_experiment(
            name="yolo-v4", owner="foo@bar.com", namespace="cv", external_id="ext1", framework=MLFramework.PYTORCH
        )
```

#### Log additional metadata
We can log arbitrary metadata related to the experiment as python dictionaries.

```
experiment.update_metadata(key="hyperparams", value={"fc_layers": 3, "lr": 0.0002})
// result has the updated_at timestamp
```

#### Log Metrics
Arbitrary metrics can be logged at any point of the experiment lifecycle. The step represents the step in the experiment such as epoch or update step, etc. The wallclock time is the human interpretable time at which the metrics is created, and the value is the metric value. The following types of values are supported - float, strings and bytes. Tensors can be serialized to bytes or strings.

```
experiment.log_metrics(metrics={'loss': 2.4, 'accu': 97.6}, step=10, wallclock=12345)
```

#### Log Events
Events can be logged while training models to make debugging and improve the observability of training workflows. Other MLOps and inference systems can also log events against a model to provide information about how and where a model is being consumed or transformed.

The following code logs an event about checkpoint store event from a trainer and records the wallclock time and checkpoint size. The events could be read to troubleshoot performance issues of checkpoint write operations.
```
experiment.log_event(Event(name="checkpoint_started", source=EventSource(name="trainer"), wallclock_time = 12000 , metadata={"chk_size": 2345}))
torch.save()
experiment.log_event(Event(name="checkpoint_finish", source=EventSource(name="trainer"), wallclock_time = 12500 , metadata={"write_speed": 2000}))
```

## gRPC API
The grpc APIs that are used by the SDKs for logging experiment metadata - 

```
  // Creates a new experiment
  rpc CreateExperiment(CreateExperimentRequest)
      returns (CreateExperimentResponse);

  // Persists a set of metadata related to objects
  rpc UpdateMetadata(UpdateMetadataRequest) returns (UpdateMetadataResponse);

  // Log Metrics for an experiment, model or checkpoint
  rpc LogMetrics(LogMetricsRequest) returns (LogMetricsResponse);

message CreateExperimentRequest {
  string name = 1;
  string owner = 2;
  string namespace = 3;
  MLFramework framework = 4;
  string task = 5;
  Metadata metadata = 6;
  string external_id = 7;
}

message CreateExperimentResponse {
  string experiment_id = 1;
  bool experiment_exists = 2;
  google.protobuf.Timestamp created_at = 20;
  google.protobuf.Timestamp updated_at = 21;
}

message UpdateMetadataRequest {
  string parent_id = 1;
  Metadata metadata = 2;
}

message UpdateMetadataResponse {
  int32 num_keys_written = 1;
  google.protobuf.Timestamp updated_at = 5;
}
```

