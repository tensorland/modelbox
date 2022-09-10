---
sidebar_position: 5
---


# Creating Model and Model Versions

# Uploading Trained Models and Training Artifacts

Models and Training Artifacts can be logged using the SDKs. The server provides a streaming gRPC API to make the upload work well using a fixed memory footprint.

## Python SDK

* `upload_artifact` 
This is the API to upload any artifact to the server. It handles chunking a file, calculating the checksum and subsequently uploading it to the server.
```
upload_artifact(self, parent: str, path: str, artifact_type: ArtifactMime) -> UploadArtifactResponse
```

    - `parent` - id of the parent object which owns the artifact. So if it's a dataset file the parent could be an experiment id or model id. Or if it's a model binary, the parent could be a model version id, which contains more metadata about the model.
    - `path` - path to the artifact file.
    - `artifact_type` - type of the artifact. Possible values are - 

    ```
    class ArtifactMime(Enum):
        Unknown = 0
        ModelVersion = 1
        Checkpoint = 2
        Text = 3
        Image = 4
        Video = 5
        Audio = 6
    ```

## gRPC API

```
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
