---
sidebar_position: 5
---


# Creating Model and Model Versions

## Terminologies 

* Model - Model objects are used to hold common metadata across all versions of models trained to solve a particular Machine Learning task. For example, Language ID models which are trained must be evaluated against a particular test dataset that an organization cares about. This allows for building some common knowledge about the characteristics of the problem domain when evaluating new models. 

* ModelVersion -  ModelVersions are trained models coming out of experiments. They usually have metrics related to accuracy and performance attached to them which helps in understanding the expected behavior when applications consume them. ModelVersions can be stored in external storage systems and their location can be tracked, or they can be uploaded in ModelBox directly and stored in the configured storage backend.

## Model - Python API 

### Create Model
```
model = client.new_model(name, owner, namespace, task, description, artifacts, metadata)
```

* name - Name of the model
* owner - Owner of the model
* namespace - Namespace to which a model is attached.
* task - Task that versions of this model. Ex - English ASR, Language ID, etc.
* description - A brief description of the model.
* artifacts - List of artifacts to track. This doesn't upload the artifacts. Use the upload_artifacts API to upload artifacts related to the model.
* metadata - Arbitrary key/value metadata related to the model


## ModelVersion - Python API

### Create Model Version

```
model_version = model.new_model_version(version, name, description, artifacts, metadata, unique_tags, framework)
```
* version - Version of the model 
* name -  Name of the model version. If this is omitted the name of the model is used.
* description - Description of the model version.
* artifacts - Artifacts tracked with the model version.
* metadata -  Additional key/value associated with the version.
* unique_tags - Tags to identify the model version. These are unique across all the versions of a given model. They are useful in denoting something unique about a model such as the version deployed in production. 
* framework - Framework used to build the model.

## APIs to log metrics, metadata, track and upload Artifacts

The following APIs are common across models and model versions. Once a model or model version object is created the following APIs are available on those objects.

### Log Metrics 
```
log_metrics(metrics, step, wallclock)
```
* metrics - A dictionary of metrics keys and values. The values could be either a float, string or bytes. 
* step - The step in the training lifecycle when the metric was emitted. For example the epoch number of a training loop.
* wallclock - the wallclock time when the metric was logged.  In the absence of the step value, metrics are ordered by wallclock.

### Log Metadata
```
update_metadata(key, value)
```

* key - key to identifying the metadata
* value - Any arbitrary python value that can be JSON encoded.

### Upload Artifacts 
```
upload_artifact(files)
```
* files - A list of file paths to be uploaded.

### Track Artifacts stored elsewhere
```
track_artifacts(files)
```
* files - A list of file paths to be tracked.

### Listing Artifacts
Once artifacts are tracked, the list of artifacts can be fetched. This will download the Artifact metadata
```
artifacts() -> List[Artifact]
```

Artifact has the following attributes  -
```
class Artifact:
    parent: str
    path: str
    mime_type: ArtifactMime = ArtifactMime.Unknown
    checksum: str = ""
    id: str = ""
```

## gRPC API

The following gRPC APIs allow creating model and model versions

```
  // Create a new Model under a namespace. If no namespace is specified, models
  // are created under a default namespace.
  rpc CreateModel(CreateModelRequest) returns (CreateModelResponse);

  // List Models uploaded for a namespace 
  rpc ListModels(ListModelsRequest) returns (ListModelsResponse);

  // Creates a new model version for a model
  rpc CreateModelVersion(CreateModelVersionRequest)
      returns (CreateModelVersionResponse);

  // Lists model versions for a model.
  rpc ListModelVersions(ListModelVersionsRequest)
      returns (ListModelVersionsResponse);


// UploadFile streams a files to ModelBox and stores the binaries to the condfigured storage
rpc UploadFile(stream UploadFileRequest) returns (UploadFileResponse);

message UploadFileRequest {
  oneof stream_frame {
    FileMetadata metadata = 1;
    bytes chunks = 2;
  }
}

message FileMetadata {
  string id = 1;

  // The ID of the checkpoint, experiment, model to which this file belongs to
  string parent_id = 2;

  // MIMEType of the file
  FileType file_type = 3;

  // checksum of the file 
  string checksum = 4;

  // path of the file
  string path = 5;

  google.protobuf.Timestamp created_at = 20;
  google.protobuf.Timestamp updated_at = 21;
}

enum FileType {
  UNDEFINED = 0;
  MODEL = 1;
  CHECKPOINT = 2;
  TEXT = 3;
  IMAGE = 4;
  AUDIO = 5;
  VIDEO = 6;
}
```
