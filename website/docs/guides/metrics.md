---
sidebar_position: 4 
---

# Logging Experiment and Model Metrics
ModelBox integrates with metrics storage services to store training hardware, experiment and model metrics.

## Python SDK
Metrics can be logged against any object in ModelBox - models, experiments, specific model versions, etc. A `MetricValue` is logged for the object id at a given timestamp.

### API 
* `MetricValue`
```
class MetricValue:
    step: int
    wallclock_time: int
    value: Union[float, str, bytes]
```

The value could be a float to represent a scaler value or bytes or strings to represent serialized tensors.
The `step` is optional and should be a real number if it represents the logical step at a given time of an experiment.
The `wallclock` time is the physical clock time at which the metric was logged.

* SDK API

```
log_metrics(self, parent_id: str, key: str, value: MetricValue)
```

## gRPC API
```
// Log Metrics for an experiment, model or checkpoint
rpc LogMetrics(LogMetricsRequest) returns (LogMetricsResponse);

// Get metrics logged for an experiment, model or checkpoint.
rpc GetMetrics(GetMetricsRequest) returns (GetMetricsResponse);

// Metrics contain the metric values for a given key
message Metrics {
  string key = 1;

  repeated MetricsValue values = 2;
}

// Metric Value at a given point of time.
message MetricsValue {
  uint64 step = 1;

  uint64 wallclock_time = 2;

  oneof value {
     float f_val = 5;

     string s_tensor = 6;

     bytes b_tensor = 7;
  }
}

// Message for logging a metric value at a given time
message LogMetricsRequest {
  string parent_id = 1;

  string key = 2;

  MetricsValue value = 3;
}

message LogMetricsResponse {}

message GetMetricsRequest {
  string parent_id = 1;
}

message GetMetricsResponse {
  repeated Metrics metrics = 1;
}
```
