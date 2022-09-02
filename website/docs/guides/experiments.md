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
result = client.create_experiment(
            name="yolo-v4", owner="foo@bar.com", namespace="cv", external_id="ext1", framework=MLFramework.PYTORCH
        )
experiment_id = result.experiment_id
```

#### Log additional metadata
We can log arbitrary metadata related to the experiment as python dictionaries.

```
result = client.update_metadata(parent_id=experiment_id, key="hyperparams", value={"fc_layers": 3, "lr": 0.0002})
// result has the updated_at timestamp
```

#### Log Metrics
Arbitrary metrics can be logged at any point of the experiment lifecycle. The step represents the step in the experiment such as epoch or update step, etc. The wallclock time is the human interpretable time at which the metrics is created, and the value is the metric value. The following types of values are supported - float, strings and bytes. Tensors can be serialized to bytes or strings.

```
metric_value = MetricValue(step=1, wallclock_time=12325, value=97.6)
result = client.log_metrics(parent_id=experiment_id, key="val_accu", value=metric_value)
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

