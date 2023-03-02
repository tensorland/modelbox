/// Request to watch events in a namespace, such as experiments/models/mocel versions
/// being created or updated.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WatchNamespaceRequest {
    #[prost(string, tag = "1")]
    pub namespace: ::prost::alloc::string::String,
    #[prost(uint64, tag = "2")]
    pub since: u64,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct WatchNamespaceResponse {
    #[prost(enumeration = "ChangeEvent", tag = "1")]
    pub event: i32,
    #[prost(message, optional, tag = "2")]
    pub payload: ::core::option::Option<::prost_types::Value>,
}
/// Metrics contain the metric values for a given key
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Metrics {
    #[prost(string, tag = "1")]
    pub key: ::prost::alloc::string::String,
    #[prost(message, repeated, tag = "2")]
    pub values: ::prost::alloc::vec::Vec<MetricsValue>,
}
/// Metric Value at a given point of time.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct MetricsValue {
    #[prost(uint64, tag = "1")]
    pub step: u64,
    #[prost(uint64, tag = "2")]
    pub wallclock_time: u64,
    #[prost(oneof = "metrics_value::Value", tags = "5, 6, 7")]
    pub value: ::core::option::Option<metrics_value::Value>,
}
/// Nested message and enum types in `MetricsValue`.
pub mod metrics_value {
    #[allow(clippy::derive_partial_eq_without_eq)]
    #[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum Value {
        #[prost(float, tag = "5")]
        FVal(f32),
        #[prost(string, tag = "6")]
        STensor(::prost::alloc::string::String),
        #[prost(bytes, tag = "7")]
        BTensor(::prost::alloc::vec::Vec<u8>),
    }
}
/// Message for logging a metric value at a given time
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct LogMetricsRequest {
    #[prost(string, tag = "1")]
    pub parent_id: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub key: ::prost::alloc::string::String,
    #[prost(message, optional, tag = "3")]
    pub value: ::core::option::Option<MetricsValue>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct LogMetricsResponse {}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetMetricsRequest {
    #[prost(string, tag = "1")]
    pub parent_id: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetMetricsResponse {
    #[prost(message, repeated, tag = "1")]
    pub metrics: ::prost::alloc::vec::Vec<Metrics>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TrackArtifactsRequest {
    #[prost(message, repeated, tag = "1")]
    pub files: ::prost::alloc::vec::Vec<FileMetadata>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TrackArtifactsResponse {}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ListArtifactsRequest {
    #[prost(string, tag = "1")]
    pub parent_id: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ListArtifactsResponse {
    #[prost(message, repeated, tag = "1")]
    pub files: ::prost::alloc::vec::Vec<FileMetadata>,
}
///
/// FileMetadata contains information about the file associated with a model version
/// such as model binaries, other meta data files related to the model.
/// This could either be sent as part of the model version creation request to track files
/// already managed by another storage service, or as the first message while uploading a file
/// to be managed by ModelBox.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct FileMetadata {
    #[prost(string, tag = "1")]
    pub id: ::prost::alloc::string::String,
    /// The ID of the experiment, model to which this file belongs to
    #[prost(string, tag = "2")]
    pub parent_id: ::prost::alloc::string::String,
    /// MIMEType of the file
    #[prost(enumeration = "FileType", tag = "3")]
    pub file_type: i32,
    /// checksum of the file
    #[prost(string, tag = "4")]
    pub checksum: ::prost::alloc::string::String,
    /// path of the file
    #[prost(string, tag = "5")]
    pub path: ::prost::alloc::string::String,
    #[prost(message, optional, tag = "20")]
    pub created_at: ::core::option::Option<::prost_types::Timestamp>,
    #[prost(message, optional, tag = "21")]
    pub updated_at: ::core::option::Option<::prost_types::Timestamp>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DownloadFileRequest {
    #[prost(string, tag = "1")]
    pub file_id: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DownloadFileResponse {
    #[prost(oneof = "download_file_response::StreamFrame", tags = "1, 2")]
    pub stream_frame: ::core::option::Option<download_file_response::StreamFrame>,
}
/// Nested message and enum types in `DownloadFileResponse`.
pub mod download_file_response {
    #[allow(clippy::derive_partial_eq_without_eq)]
    #[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum StreamFrame {
        #[prost(message, tag = "1")]
        Metadata(super::FileMetadata),
        #[prost(bytes, tag = "2")]
        Chunks(::prost::alloc::vec::Vec<u8>),
    }
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct UploadFileRequest {
    #[prost(oneof = "upload_file_request::StreamFrame", tags = "1, 2")]
    pub stream_frame: ::core::option::Option<upload_file_request::StreamFrame>,
}
/// Nested message and enum types in `UploadFileRequest`.
pub mod upload_file_request {
    #[allow(clippy::derive_partial_eq_without_eq)]
    #[derive(Clone, PartialEq, ::prost::Oneof)]
    pub enum StreamFrame {
        #[prost(message, tag = "1")]
        Metadata(super::FileMetadata),
        #[prost(bytes, tag = "2")]
        Chunks(::prost::alloc::vec::Vec<u8>),
    }
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct UploadFileResponse {
    #[prost(string, tag = "1")]
    pub file_id: ::prost::alloc::string::String,
}
///
/// Model contains metadata about a model which solves a particular use case.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Model {
    #[prost(string, tag = "1")]
    pub id: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub name: ::prost::alloc::string::String,
    #[prost(string, tag = "3")]
    pub owner: ::prost::alloc::string::String,
    #[prost(string, tag = "4")]
    pub namespace: ::prost::alloc::string::String,
    #[prost(string, tag = "5")]
    pub description: ::prost::alloc::string::String,
    #[prost(string, tag = "6")]
    pub task: ::prost::alloc::string::String,
    #[prost(message, optional, tag = "20")]
    pub created_at: ::core::option::Option<::prost_types::Timestamp>,
    #[prost(message, optional, tag = "21")]
    pub updated_at: ::core::option::Option<::prost_types::Timestamp>,
}
/// *
/// Create a new Model. If the id points to an existing model a new model version
/// is created.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CreateModelRequest {
    #[prost(string, tag = "2")]
    pub name: ::prost::alloc::string::String,
    #[prost(string, tag = "3")]
    pub owner: ::prost::alloc::string::String,
    #[prost(string, tag = "4")]
    pub namespace: ::prost::alloc::string::String,
    #[prost(string, tag = "5")]
    pub task: ::prost::alloc::string::String,
    #[prost(string, tag = "6")]
    pub description: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CreateModelResponse {
    #[prost(string, tag = "1")]
    pub id: ::prost::alloc::string::String,
    #[prost(message, optional, tag = "20")]
    pub created_at: ::core::option::Option<::prost_types::Timestamp>,
    #[prost(message, optional, tag = "21")]
    pub updated_at: ::core::option::Option<::prost_types::Timestamp>,
}
/// *
/// ModelVersion contains a trained model binary, metrics related to the mode
/// such as accuracy on various datasets, performance on a hardware, etc. Model
/// Versions are always linked to a model.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ModelVersion {
    #[prost(string, tag = "1")]
    pub id: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub model_id: ::prost::alloc::string::String,
    #[prost(string, tag = "3")]
    pub name: ::prost::alloc::string::String,
    #[prost(string, tag = "4")]
    pub version: ::prost::alloc::string::String,
    #[prost(string, tag = "5")]
    pub description: ::prost::alloc::string::String,
    #[prost(enumeration = "MlFramework", tag = "8")]
    pub framework: i32,
    #[prost(string, repeated, tag = "9")]
    pub unique_tags: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
    #[prost(message, optional, tag = "20")]
    pub created_at: ::core::option::Option<::prost_types::Timestamp>,
    #[prost(message, optional, tag = "21")]
    pub updated_at: ::core::option::Option<::prost_types::Timestamp>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CreateModelVersionRequest {
    #[prost(string, tag = "1")]
    pub model: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub name: ::prost::alloc::string::String,
    #[prost(string, tag = "3")]
    pub version: ::prost::alloc::string::String,
    #[prost(string, tag = "4")]
    pub description: ::prost::alloc::string::String,
    #[prost(string, tag = "5")]
    pub namespace: ::prost::alloc::string::String,
    #[prost(enumeration = "MlFramework", tag = "8")]
    pub framework: i32,
    #[prost(string, repeated, tag = "9")]
    pub unique_tags: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CreateModelVersionResponse {
    #[prost(string, tag = "1")]
    pub model_version: ::prost::alloc::string::String,
    #[prost(message, optional, tag = "20")]
    pub created_at: ::core::option::Option<::prost_types::Timestamp>,
    #[prost(message, optional, tag = "21")]
    pub updated_at: ::core::option::Option<::prost_types::Timestamp>,
}
/// *
/// Experiments are the sources of Model checkpoints. They track various details
/// related to the training runs which created the models such as hyper
/// parameters, etc.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Experiment {
    #[prost(string, tag = "1")]
    pub id: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub name: ::prost::alloc::string::String,
    #[prost(string, tag = "3")]
    pub namespace: ::prost::alloc::string::String,
    #[prost(string, tag = "4")]
    pub owner: ::prost::alloc::string::String,
    #[prost(enumeration = "MlFramework", tag = "5")]
    pub framework: i32,
    #[prost(string, tag = "7")]
    pub external_id: ::prost::alloc::string::String,
    #[prost(message, optional, tag = "20")]
    pub created_at: ::core::option::Option<::prost_types::Timestamp>,
    #[prost(message, optional, tag = "21")]
    pub updated_at: ::core::option::Option<::prost_types::Timestamp>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CreateExperimentRequest {
    #[prost(string, tag = "1")]
    pub name: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub owner: ::prost::alloc::string::String,
    #[prost(string, tag = "3")]
    pub namespace: ::prost::alloc::string::String,
    #[prost(enumeration = "MlFramework", tag = "4")]
    pub framework: i32,
    #[prost(string, tag = "5")]
    pub task: ::prost::alloc::string::String,
    #[prost(string, tag = "7")]
    pub external_id: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct CreateExperimentResponse {
    #[prost(string, tag = "1")]
    pub experiment_id: ::prost::alloc::string::String,
    #[prost(bool, tag = "2")]
    pub experiment_exists: bool,
    #[prost(message, optional, tag = "20")]
    pub created_at: ::core::option::Option<::prost_types::Timestamp>,
    #[prost(message, optional, tag = "21")]
    pub updated_at: ::core::option::Option<::prost_types::Timestamp>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ListExperimentsRequest {
    #[prost(string, tag = "1")]
    pub namespace: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ListExperimentsResponse {
    #[prost(message, repeated, tag = "1")]
    pub experiments: ::prost::alloc::vec::Vec<Experiment>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ListModelVersionsRequest {
    #[prost(string, tag = "1")]
    pub model: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ListModelVersionsResponse {
    #[prost(message, repeated, tag = "1")]
    pub model_versions: ::prost::alloc::vec::Vec<ModelVersion>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ListModelsRequest {
    #[prost(string, tag = "1")]
    pub namespace: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ListModelsResponse {
    #[prost(message, repeated, tag = "1")]
    pub models: ::prost::alloc::vec::Vec<Model>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Metadata {
    #[prost(map = "string, string", tag = "1")]
    pub metadata:
        ::std::collections::HashMap<::prost::alloc::string::String, ::prost::alloc::string::String>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct UpdateMetadataRequest {
    #[prost(string, tag = "1")]
    pub parent_id: ::prost::alloc::string::String,
    #[prost(message, optional, tag = "2")]
    pub metadata: ::core::option::Option<Metadata>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct UpdateMetadataResponse {}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ListMetadataRequest {
    #[prost(string, tag = "1")]
    pub parent_id: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ListMetadataResponse {
    #[prost(message, optional, tag = "1")]
    pub metadata: ::core::option::Option<Metadata>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct EventSource {
    #[prost(string, tag = "1")]
    pub name: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Event {
    #[prost(string, tag = "2")]
    pub name: ::prost::alloc::string::String,
    #[prost(message, optional, tag = "3")]
    pub source: ::core::option::Option<EventSource>,
    #[prost(message, optional, tag = "4")]
    pub wallclock_time: ::core::option::Option<::prost_types::Timestamp>,
    #[prost(message, optional, tag = "5")]
    pub metadata: ::core::option::Option<Metadata>,
}
/// *
/// Contains information about an event being logged about
/// an experiment or a model or a checkpoint by any system interacting
/// or using the object.
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct LogEventRequest {
    #[prost(string, tag = "1")]
    pub parent_id: ::prost::alloc::string::String,
    #[prost(message, optional, tag = "2")]
    pub event: ::core::option::Option<Event>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct LogEventResponse {
    #[prost(message, optional, tag = "1")]
    pub created_at: ::core::option::Option<::prost_types::Timestamp>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ListEventsRequest {
    #[prost(string, tag = "1")]
    pub parent_id: ::prost::alloc::string::String,
    #[prost(message, optional, tag = "2")]
    pub since: ::core::option::Option<::prost_types::Timestamp>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ListEventsResponse {
    #[prost(message, repeated, tag = "1")]
    pub events: ::prost::alloc::vec::Vec<Event>,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetExperimentRequest {
    #[prost(string, tag = "1")]
    pub id: ::prost::alloc::string::String,
}
#[allow(clippy::derive_partial_eq_without_eq)]
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GetExperimentResponse {
    #[prost(message, optional, tag = "1")]
    pub experiment: ::core::option::Option<Experiment>,
}
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum ChangeEvent {
    Undefined = 0,
    ObjectCreated = 1,
    ObjectUpdated = 2,
}
impl ChangeEvent {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            ChangeEvent::Undefined => "CHANGE_EVENT_UNDEFINED",
            ChangeEvent::ObjectCreated => "OBJECT_CREATED",
            ChangeEvent::ObjectUpdated => "OBJECT_UPDATED",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "CHANGE_EVENT_UNDEFINED" => Some(Self::Undefined),
            "OBJECT_CREATED" => Some(Self::ObjectCreated),
            "OBJECT_UPDATED" => Some(Self::ObjectUpdated),
            _ => None,
        }
    }
}
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum FileType {
    Undefined = 0,
    Model = 1,
    Checkpoint = 2,
    Text = 3,
    Image = 4,
    Audio = 5,
    Video = 6,
}
impl FileType {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            FileType::Undefined => "UNDEFINED",
            FileType::Model => "MODEL",
            FileType::Checkpoint => "CHECKPOINT",
            FileType::Text => "TEXT",
            FileType::Image => "IMAGE",
            FileType::Audio => "AUDIO",
            FileType::Video => "VIDEO",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "UNDEFINED" => Some(Self::Undefined),
            "MODEL" => Some(Self::Model),
            "CHECKPOINT" => Some(Self::Checkpoint),
            "TEXT" => Some(Self::Text),
            "IMAGE" => Some(Self::Image),
            "AUDIO" => Some(Self::Audio),
            "VIDEO" => Some(Self::Video),
            _ => None,
        }
    }
}
///
/// Deep Learning frameworks known to ModelBox
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum MlFramework {
    Unknown = 0,
    Pytorch = 1,
    Keras = 2,
}
impl MlFramework {
    /// String value of the enum field names used in the ProtoBuf definition.
    ///
    /// The values are not transformed in any way and thus are considered stable
    /// (if the ProtoBuf definition does not change) and safe for programmatic use.
    pub fn as_str_name(&self) -> &'static str {
        match self {
            MlFramework::Unknown => "UNKNOWN",
            MlFramework::Pytorch => "PYTORCH",
            MlFramework::Keras => "KERAS",
        }
    }
    /// Creates an enum from field names used in the ProtoBuf definition.
    pub fn from_str_name(value: &str) -> ::core::option::Option<Self> {
        match value {
            "UNKNOWN" => Some(Self::Unknown),
            "PYTORCH" => Some(Self::Pytorch),
            "KERAS" => Some(Self::Keras),
            _ => None,
        }
    }
}
/// Generated client implementations.
pub mod model_store_client {
    #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
    use tonic::codegen::http::Uri;
    use tonic::codegen::*;
    /// *
    /// ModelStore is the service exposed to upload trained models and training
    /// checkpoints, and manage metadata around them.
    #[derive(Debug, Clone)]
    pub struct ModelStoreClient<T> {
        inner: tonic::client::Grpc<T>,
    }
    impl ModelStoreClient<tonic::transport::Channel> {
        /// Attempt to create a new client by connecting to a given endpoint.
        pub async fn connect<D>(dst: D) -> Result<Self, tonic::transport::Error>
        where
            D: std::convert::TryInto<tonic::transport::Endpoint>,
            D::Error: Into<StdError>,
        {
            let conn = tonic::transport::Endpoint::new(dst)?.connect().await?;
            Ok(Self::new(conn))
        }
    }
    impl<T> ModelStoreClient<T>
    where
        T: tonic::client::GrpcService<tonic::body::BoxBody>,
        T::Error: Into<StdError>,
        T::ResponseBody: Body<Data = Bytes> + Send + 'static,
        <T::ResponseBody as Body>::Error: Into<StdError> + Send,
    {
        pub fn new(inner: T) -> Self {
            let inner = tonic::client::Grpc::new(inner);
            Self { inner }
        }
        pub fn with_origin(inner: T, origin: Uri) -> Self {
            let inner = tonic::client::Grpc::with_origin(inner, origin);
            Self { inner }
        }
        pub fn with_interceptor<F>(
            inner: T,
            interceptor: F,
        ) -> ModelStoreClient<InterceptedService<T, F>>
        where
            F: tonic::service::Interceptor,
            T::ResponseBody: Default,
            T: tonic::codegen::Service<
                http::Request<tonic::body::BoxBody>,
                Response = http::Response<
                    <T as tonic::client::GrpcService<tonic::body::BoxBody>>::ResponseBody,
                >,
            >,
            <T as tonic::codegen::Service<http::Request<tonic::body::BoxBody>>>::Error:
                Into<StdError> + Send + Sync,
        {
            ModelStoreClient::new(InterceptedService::new(inner, interceptor))
        }
        /// Compress requests with the given encoding.
        ///
        /// This requires the server to support it otherwise it might respond with an
        /// error.
        #[must_use]
        pub fn send_compressed(mut self, encoding: CompressionEncoding) -> Self {
            self.inner = self.inner.send_compressed(encoding);
            self
        }
        /// Enable decompressing responses.
        #[must_use]
        pub fn accept_compressed(mut self, encoding: CompressionEncoding) -> Self {
            self.inner = self.inner.accept_compressed(encoding);
            self
        }
        /// Create a new Model under a namespace. If no namespace is specified, models
        /// are created under a default namespace.
        pub async fn create_model(
            &mut self,
            request: impl tonic::IntoRequest<super::CreateModelRequest>,
        ) -> Result<tonic::Response<super::CreateModelResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/modelbox.ModelStore/CreateModel");
            self.inner.unary(request.into_request(), path, codec).await
        }
        /// List Models uploaded for a namespace
        pub async fn list_models(
            &mut self,
            request: impl tonic::IntoRequest<super::ListModelsRequest>,
        ) -> Result<tonic::Response<super::ListModelsResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/modelbox.ModelStore/ListModels");
            self.inner.unary(request.into_request(), path, codec).await
        }
        /// Creates a new model version for a model
        pub async fn create_model_version(
            &mut self,
            request: impl tonic::IntoRequest<super::CreateModelVersionRequest>,
        ) -> Result<tonic::Response<super::CreateModelVersionResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path =
                http::uri::PathAndQuery::from_static("/modelbox.ModelStore/CreateModelVersion");
            self.inner.unary(request.into_request(), path, codec).await
        }
        /// Lists model versions for a model.
        pub async fn list_model_versions(
            &mut self,
            request: impl tonic::IntoRequest<super::ListModelVersionsRequest>,
        ) -> Result<tonic::Response<super::ListModelVersionsResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path =
                http::uri::PathAndQuery::from_static("/modelbox.ModelStore/ListModelVersions");
            self.inner.unary(request.into_request(), path, codec).await
        }
        /// Creates a new experiment
        pub async fn create_experiment(
            &mut self,
            request: impl tonic::IntoRequest<super::CreateExperimentRequest>,
        ) -> Result<tonic::Response<super::CreateExperimentResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path =
                http::uri::PathAndQuery::from_static("/modelbox.ModelStore/CreateExperiment");
            self.inner.unary(request.into_request(), path, codec).await
        }
        /// List Experiments
        pub async fn list_experiments(
            &mut self,
            request: impl tonic::IntoRequest<super::ListExperimentsRequest>,
        ) -> Result<tonic::Response<super::ListExperimentsResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/modelbox.ModelStore/ListExperiments");
            self.inner.unary(request.into_request(), path, codec).await
        }
        /// Get Experiments
        pub async fn get_experiment(
            &mut self,
            request: impl tonic::IntoRequest<super::GetExperimentRequest>,
        ) -> Result<tonic::Response<super::GetExperimentResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/modelbox.ModelStore/GetExperiment");
            self.inner.unary(request.into_request(), path, codec).await
        }
        /// UploadFile streams a files to ModelBox and stores the binaries to the condfigured storage
        pub async fn upload_file(
            &mut self,
            request: impl tonic::IntoStreamingRequest<Message = super::UploadFileRequest>,
        ) -> Result<tonic::Response<super::UploadFileResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/modelbox.ModelStore/UploadFile");
            self.inner
                .client_streaming(request.into_streaming_request(), path, codec)
                .await
        }
        /// DownloadFile downloads a file from configured storage
        pub async fn download_file(
            &mut self,
            request: impl tonic::IntoRequest<super::DownloadFileRequest>,
        ) -> Result<
            tonic::Response<tonic::codec::Streaming<super::DownloadFileResponse>>,
            tonic::Status,
        > {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/modelbox.ModelStore/DownloadFile");
            self.inner
                .server_streaming(request.into_request(), path, codec)
                .await
        }
        /// Persists a set of metadata related to objects
        pub async fn update_metadata(
            &mut self,
            request: impl tonic::IntoRequest<super::UpdateMetadataRequest>,
        ) -> Result<tonic::Response<super::UpdateMetadataResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/modelbox.ModelStore/UpdateMetadata");
            self.inner.unary(request.into_request(), path, codec).await
        }
        /// Lists metadata associated with an object
        pub async fn list_metadata(
            &mut self,
            request: impl tonic::IntoRequest<super::ListMetadataRequest>,
        ) -> Result<tonic::Response<super::ListMetadataResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/modelbox.ModelStore/ListMetadata");
            self.inner.unary(request.into_request(), path, codec).await
        }
        /// Tracks a set of artifacts with a experiment/checkpoint/model
        pub async fn track_artifacts(
            &mut self,
            request: impl tonic::IntoRequest<super::TrackArtifactsRequest>,
        ) -> Result<tonic::Response<super::TrackArtifactsResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/modelbox.ModelStore/TrackArtifacts");
            self.inner.unary(request.into_request(), path, codec).await
        }
        /// List artifacts for an expriment/model/model version
        pub async fn list_artifacts(
            &mut self,
            request: impl tonic::IntoRequest<super::ListArtifactsRequest>,
        ) -> Result<tonic::Response<super::ListArtifactsResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/modelbox.ModelStore/ListArtifacts");
            self.inner.unary(request.into_request(), path, codec).await
        }
        /// Log Metrics for an experiment, model or checkpoint
        pub async fn log_metrics(
            &mut self,
            request: impl tonic::IntoRequest<super::LogMetricsRequest>,
        ) -> Result<tonic::Response<super::LogMetricsResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/modelbox.ModelStore/LogMetrics");
            self.inner.unary(request.into_request(), path, codec).await
        }
        /// Get metrics logged for an experiment, model or checkpoint.
        pub async fn get_metrics(
            &mut self,
            request: impl tonic::IntoRequest<super::GetMetricsRequest>,
        ) -> Result<tonic::Response<super::GetMetricsResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/modelbox.ModelStore/GetMetrics");
            self.inner.unary(request.into_request(), path, codec).await
        }
        /// Log an event from any system interacting with metadata of a experiment, models or
        /// using a trained model or checkpoint.
        pub async fn log_event(
            &mut self,
            request: impl tonic::IntoRequest<super::LogEventRequest>,
        ) -> Result<tonic::Response<super::LogEventResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/modelbox.ModelStore/LogEvent");
            self.inner.unary(request.into_request(), path, codec).await
        }
        /// List events logged for an experiment/model, etc.
        pub async fn list_events(
            &mut self,
            request: impl tonic::IntoRequest<super::ListEventsRequest>,
        ) -> Result<tonic::Response<super::ListEventsResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/modelbox.ModelStore/ListEvents");
            self.inner.unary(request.into_request(), path, codec).await
        }
        /// Streams change events in any of objects such as experiments, models, etc, for a given namespace
        /// Response is a json representation of the new state of the obejct
        pub async fn watch_namespace(
            &mut self,
            request: impl tonic::IntoRequest<super::WatchNamespaceRequest>,
        ) -> Result<
            tonic::Response<tonic::codec::Streaming<super::WatchNamespaceResponse>>,
            tonic::Status,
        > {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/modelbox.ModelStore/WatchNamespace");
            self.inner
                .server_streaming(request.into_request(), path, codec)
                .await
        }
    }
}
/// Generated server implementations.
pub mod model_store_server {
    #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
    use tonic::codegen::*;
    /// Generated trait containing gRPC methods that should be implemented for use with ModelStoreServer.
    #[async_trait]
    pub trait ModelStore: Send + Sync + 'static {
        /// Create a new Model under a namespace. If no namespace is specified, models
        /// are created under a default namespace.
        async fn create_model(
            &self,
            request: tonic::Request<super::CreateModelRequest>,
        ) -> Result<tonic::Response<super::CreateModelResponse>, tonic::Status>;
        /// List Models uploaded for a namespace
        async fn list_models(
            &self,
            request: tonic::Request<super::ListModelsRequest>,
        ) -> Result<tonic::Response<super::ListModelsResponse>, tonic::Status>;
        /// Creates a new model version for a model
        async fn create_model_version(
            &self,
            request: tonic::Request<super::CreateModelVersionRequest>,
        ) -> Result<tonic::Response<super::CreateModelVersionResponse>, tonic::Status>;
        /// Lists model versions for a model.
        async fn list_model_versions(
            &self,
            request: tonic::Request<super::ListModelVersionsRequest>,
        ) -> Result<tonic::Response<super::ListModelVersionsResponse>, tonic::Status>;
        /// Creates a new experiment
        async fn create_experiment(
            &self,
            request: tonic::Request<super::CreateExperimentRequest>,
        ) -> Result<tonic::Response<super::CreateExperimentResponse>, tonic::Status>;
        /// List Experiments
        async fn list_experiments(
            &self,
            request: tonic::Request<super::ListExperimentsRequest>,
        ) -> Result<tonic::Response<super::ListExperimentsResponse>, tonic::Status>;
        /// Get Experiments
        async fn get_experiment(
            &self,
            request: tonic::Request<super::GetExperimentRequest>,
        ) -> Result<tonic::Response<super::GetExperimentResponse>, tonic::Status>;
        /// UploadFile streams a files to ModelBox and stores the binaries to the condfigured storage
        async fn upload_file(
            &self,
            request: tonic::Request<tonic::Streaming<super::UploadFileRequest>>,
        ) -> Result<tonic::Response<super::UploadFileResponse>, tonic::Status>;
        /// Server streaming response type for the DownloadFile method.
        type DownloadFileStream: futures_core::Stream<Item = Result<super::DownloadFileResponse, tonic::Status>>
            + Send
            + 'static;
        /// DownloadFile downloads a file from configured storage
        async fn download_file(
            &self,
            request: tonic::Request<super::DownloadFileRequest>,
        ) -> Result<tonic::Response<Self::DownloadFileStream>, tonic::Status>;
        /// Persists a set of metadata related to objects
        async fn update_metadata(
            &self,
            request: tonic::Request<super::UpdateMetadataRequest>,
        ) -> Result<tonic::Response<super::UpdateMetadataResponse>, tonic::Status>;
        /// Lists metadata associated with an object
        async fn list_metadata(
            &self,
            request: tonic::Request<super::ListMetadataRequest>,
        ) -> Result<tonic::Response<super::ListMetadataResponse>, tonic::Status>;
        /// Tracks a set of artifacts with a experiment/checkpoint/model
        async fn track_artifacts(
            &self,
            request: tonic::Request<super::TrackArtifactsRequest>,
        ) -> Result<tonic::Response<super::TrackArtifactsResponse>, tonic::Status>;
        /// List artifacts for an expriment/model/model version
        async fn list_artifacts(
            &self,
            request: tonic::Request<super::ListArtifactsRequest>,
        ) -> Result<tonic::Response<super::ListArtifactsResponse>, tonic::Status>;
        /// Log Metrics for an experiment, model or checkpoint
        async fn log_metrics(
            &self,
            request: tonic::Request<super::LogMetricsRequest>,
        ) -> Result<tonic::Response<super::LogMetricsResponse>, tonic::Status>;
        /// Get metrics logged for an experiment, model or checkpoint.
        async fn get_metrics(
            &self,
            request: tonic::Request<super::GetMetricsRequest>,
        ) -> Result<tonic::Response<super::GetMetricsResponse>, tonic::Status>;
        /// Log an event from any system interacting with metadata of a experiment, models or
        /// using a trained model or checkpoint.
        async fn log_event(
            &self,
            request: tonic::Request<super::LogEventRequest>,
        ) -> Result<tonic::Response<super::LogEventResponse>, tonic::Status>;
        /// List events logged for an experiment/model, etc.
        async fn list_events(
            &self,
            request: tonic::Request<super::ListEventsRequest>,
        ) -> Result<tonic::Response<super::ListEventsResponse>, tonic::Status>;
        /// Server streaming response type for the WatchNamespace method.
        type WatchNamespaceStream: futures_core::Stream<Item = Result<super::WatchNamespaceResponse, tonic::Status>>
            + Send
            + 'static;
        /// Streams change events in any of objects such as experiments, models, etc, for a given namespace
        /// Response is a json representation of the new state of the obejct
        async fn watch_namespace(
            &self,
            request: tonic::Request<super::WatchNamespaceRequest>,
        ) -> Result<tonic::Response<Self::WatchNamespaceStream>, tonic::Status>;
    }
    /// *
    /// ModelStore is the service exposed to upload trained models and training
    /// checkpoints, and manage metadata around them.
    #[derive(Debug)]
    pub struct ModelStoreServer<T: ModelStore> {
        inner: _Inner<T>,
        accept_compression_encodings: EnabledCompressionEncodings,
        send_compression_encodings: EnabledCompressionEncodings,
    }
    struct _Inner<T>(Arc<T>);
    impl<T: ModelStore> ModelStoreServer<T> {
        pub fn new(inner: T) -> Self {
            Self::from_arc(Arc::new(inner))
        }
        pub fn from_arc(inner: Arc<T>) -> Self {
            let inner = _Inner(inner);
            Self {
                inner,
                accept_compression_encodings: Default::default(),
                send_compression_encodings: Default::default(),
            }
        }
        pub fn with_interceptor<F>(inner: T, interceptor: F) -> InterceptedService<Self, F>
        where
            F: tonic::service::Interceptor,
        {
            InterceptedService::new(Self::new(inner), interceptor)
        }
        /// Enable decompressing requests with the given encoding.
        #[must_use]
        pub fn accept_compressed(mut self, encoding: CompressionEncoding) -> Self {
            self.accept_compression_encodings.enable(encoding);
            self
        }
        /// Compress responses with the given encoding, if the client supports it.
        #[must_use]
        pub fn send_compressed(mut self, encoding: CompressionEncoding) -> Self {
            self.send_compression_encodings.enable(encoding);
            self
        }
    }
    impl<T, B> tonic::codegen::Service<http::Request<B>> for ModelStoreServer<T>
    where
        T: ModelStore,
        B: Body + Send + 'static,
        B::Error: Into<StdError> + Send + 'static,
    {
        type Response = http::Response<tonic::body::BoxBody>;
        type Error = std::convert::Infallible;
        type Future = BoxFuture<Self::Response, Self::Error>;
        fn poll_ready(&mut self, _cx: &mut Context<'_>) -> Poll<Result<(), Self::Error>> {
            Poll::Ready(Ok(()))
        }
        fn call(&mut self, req: http::Request<B>) -> Self::Future {
            let inner = self.inner.clone();
            match req.uri().path() {
                "/modelbox.ModelStore/CreateModel" => {
                    #[allow(non_camel_case_types)]
                    struct CreateModelSvc<T: ModelStore>(pub Arc<T>);
                    impl<T: ModelStore> tonic::server::UnaryService<super::CreateModelRequest> for CreateModelSvc<T> {
                        type Response = super::CreateModelResponse;
                        type Future = BoxFuture<tonic::Response<Self::Response>, tonic::Status>;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::CreateModelRequest>,
                        ) -> Self::Future {
                            let inner = self.0.clone();
                            let fut = async move { (*inner).create_model(request).await };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = CreateModelSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec).apply_compression_config(
                            accept_compression_encodings,
                            send_compression_encodings,
                        );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/modelbox.ModelStore/ListModels" => {
                    #[allow(non_camel_case_types)]
                    struct ListModelsSvc<T: ModelStore>(pub Arc<T>);
                    impl<T: ModelStore> tonic::server::UnaryService<super::ListModelsRequest> for ListModelsSvc<T> {
                        type Response = super::ListModelsResponse;
                        type Future = BoxFuture<tonic::Response<Self::Response>, tonic::Status>;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::ListModelsRequest>,
                        ) -> Self::Future {
                            let inner = self.0.clone();
                            let fut = async move { (*inner).list_models(request).await };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = ListModelsSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec).apply_compression_config(
                            accept_compression_encodings,
                            send_compression_encodings,
                        );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/modelbox.ModelStore/CreateModelVersion" => {
                    #[allow(non_camel_case_types)]
                    struct CreateModelVersionSvc<T: ModelStore>(pub Arc<T>);
                    impl<T: ModelStore>
                        tonic::server::UnaryService<super::CreateModelVersionRequest>
                        for CreateModelVersionSvc<T>
                    {
                        type Response = super::CreateModelVersionResponse;
                        type Future = BoxFuture<tonic::Response<Self::Response>, tonic::Status>;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::CreateModelVersionRequest>,
                        ) -> Self::Future {
                            let inner = self.0.clone();
                            let fut = async move { (*inner).create_model_version(request).await };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = CreateModelVersionSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec).apply_compression_config(
                            accept_compression_encodings,
                            send_compression_encodings,
                        );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/modelbox.ModelStore/ListModelVersions" => {
                    #[allow(non_camel_case_types)]
                    struct ListModelVersionsSvc<T: ModelStore>(pub Arc<T>);
                    impl<T: ModelStore> tonic::server::UnaryService<super::ListModelVersionsRequest>
                        for ListModelVersionsSvc<T>
                    {
                        type Response = super::ListModelVersionsResponse;
                        type Future = BoxFuture<tonic::Response<Self::Response>, tonic::Status>;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::ListModelVersionsRequest>,
                        ) -> Self::Future {
                            let inner = self.0.clone();
                            let fut = async move { (*inner).list_model_versions(request).await };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = ListModelVersionsSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec).apply_compression_config(
                            accept_compression_encodings,
                            send_compression_encodings,
                        );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/modelbox.ModelStore/CreateExperiment" => {
                    #[allow(non_camel_case_types)]
                    struct CreateExperimentSvc<T: ModelStore>(pub Arc<T>);
                    impl<T: ModelStore> tonic::server::UnaryService<super::CreateExperimentRequest>
                        for CreateExperimentSvc<T>
                    {
                        type Response = super::CreateExperimentResponse;
                        type Future = BoxFuture<tonic::Response<Self::Response>, tonic::Status>;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::CreateExperimentRequest>,
                        ) -> Self::Future {
                            let inner = self.0.clone();
                            let fut = async move { (*inner).create_experiment(request).await };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = CreateExperimentSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec).apply_compression_config(
                            accept_compression_encodings,
                            send_compression_encodings,
                        );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/modelbox.ModelStore/ListExperiments" => {
                    #[allow(non_camel_case_types)]
                    struct ListExperimentsSvc<T: ModelStore>(pub Arc<T>);
                    impl<T: ModelStore> tonic::server::UnaryService<super::ListExperimentsRequest>
                        for ListExperimentsSvc<T>
                    {
                        type Response = super::ListExperimentsResponse;
                        type Future = BoxFuture<tonic::Response<Self::Response>, tonic::Status>;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::ListExperimentsRequest>,
                        ) -> Self::Future {
                            let inner = self.0.clone();
                            let fut = async move { (*inner).list_experiments(request).await };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = ListExperimentsSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec).apply_compression_config(
                            accept_compression_encodings,
                            send_compression_encodings,
                        );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/modelbox.ModelStore/GetExperiment" => {
                    #[allow(non_camel_case_types)]
                    struct GetExperimentSvc<T: ModelStore>(pub Arc<T>);
                    impl<T: ModelStore> tonic::server::UnaryService<super::GetExperimentRequest>
                        for GetExperimentSvc<T>
                    {
                        type Response = super::GetExperimentResponse;
                        type Future = BoxFuture<tonic::Response<Self::Response>, tonic::Status>;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetExperimentRequest>,
                        ) -> Self::Future {
                            let inner = self.0.clone();
                            let fut = async move { (*inner).get_experiment(request).await };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = GetExperimentSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec).apply_compression_config(
                            accept_compression_encodings,
                            send_compression_encodings,
                        );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/modelbox.ModelStore/UploadFile" => {
                    #[allow(non_camel_case_types)]
                    struct UploadFileSvc<T: ModelStore>(pub Arc<T>);
                    impl<T: ModelStore>
                        tonic::server::ClientStreamingService<super::UploadFileRequest>
                        for UploadFileSvc<T>
                    {
                        type Response = super::UploadFileResponse;
                        type Future = BoxFuture<tonic::Response<Self::Response>, tonic::Status>;
                        fn call(
                            &mut self,
                            request: tonic::Request<tonic::Streaming<super::UploadFileRequest>>,
                        ) -> Self::Future {
                            let inner = self.0.clone();
                            let fut = async move { (*inner).upload_file(request).await };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = UploadFileSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec).apply_compression_config(
                            accept_compression_encodings,
                            send_compression_encodings,
                        );
                        let res = grpc.client_streaming(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/modelbox.ModelStore/DownloadFile" => {
                    #[allow(non_camel_case_types)]
                    struct DownloadFileSvc<T: ModelStore>(pub Arc<T>);
                    impl<T: ModelStore>
                        tonic::server::ServerStreamingService<super::DownloadFileRequest>
                        for DownloadFileSvc<T>
                    {
                        type Response = super::DownloadFileResponse;
                        type ResponseStream = T::DownloadFileStream;
                        type Future =
                            BoxFuture<tonic::Response<Self::ResponseStream>, tonic::Status>;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::DownloadFileRequest>,
                        ) -> Self::Future {
                            let inner = self.0.clone();
                            let fut = async move { (*inner).download_file(request).await };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = DownloadFileSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec).apply_compression_config(
                            accept_compression_encodings,
                            send_compression_encodings,
                        );
                        let res = grpc.server_streaming(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/modelbox.ModelStore/UpdateMetadata" => {
                    #[allow(non_camel_case_types)]
                    struct UpdateMetadataSvc<T: ModelStore>(pub Arc<T>);
                    impl<T: ModelStore> tonic::server::UnaryService<super::UpdateMetadataRequest>
                        for UpdateMetadataSvc<T>
                    {
                        type Response = super::UpdateMetadataResponse;
                        type Future = BoxFuture<tonic::Response<Self::Response>, tonic::Status>;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::UpdateMetadataRequest>,
                        ) -> Self::Future {
                            let inner = self.0.clone();
                            let fut = async move { (*inner).update_metadata(request).await };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = UpdateMetadataSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec).apply_compression_config(
                            accept_compression_encodings,
                            send_compression_encodings,
                        );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/modelbox.ModelStore/ListMetadata" => {
                    #[allow(non_camel_case_types)]
                    struct ListMetadataSvc<T: ModelStore>(pub Arc<T>);
                    impl<T: ModelStore> tonic::server::UnaryService<super::ListMetadataRequest> for ListMetadataSvc<T> {
                        type Response = super::ListMetadataResponse;
                        type Future = BoxFuture<tonic::Response<Self::Response>, tonic::Status>;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::ListMetadataRequest>,
                        ) -> Self::Future {
                            let inner = self.0.clone();
                            let fut = async move { (*inner).list_metadata(request).await };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = ListMetadataSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec).apply_compression_config(
                            accept_compression_encodings,
                            send_compression_encodings,
                        );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/modelbox.ModelStore/TrackArtifacts" => {
                    #[allow(non_camel_case_types)]
                    struct TrackArtifactsSvc<T: ModelStore>(pub Arc<T>);
                    impl<T: ModelStore> tonic::server::UnaryService<super::TrackArtifactsRequest>
                        for TrackArtifactsSvc<T>
                    {
                        type Response = super::TrackArtifactsResponse;
                        type Future = BoxFuture<tonic::Response<Self::Response>, tonic::Status>;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::TrackArtifactsRequest>,
                        ) -> Self::Future {
                            let inner = self.0.clone();
                            let fut = async move { (*inner).track_artifacts(request).await };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = TrackArtifactsSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec).apply_compression_config(
                            accept_compression_encodings,
                            send_compression_encodings,
                        );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/modelbox.ModelStore/ListArtifacts" => {
                    #[allow(non_camel_case_types)]
                    struct ListArtifactsSvc<T: ModelStore>(pub Arc<T>);
                    impl<T: ModelStore> tonic::server::UnaryService<super::ListArtifactsRequest>
                        for ListArtifactsSvc<T>
                    {
                        type Response = super::ListArtifactsResponse;
                        type Future = BoxFuture<tonic::Response<Self::Response>, tonic::Status>;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::ListArtifactsRequest>,
                        ) -> Self::Future {
                            let inner = self.0.clone();
                            let fut = async move { (*inner).list_artifacts(request).await };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = ListArtifactsSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec).apply_compression_config(
                            accept_compression_encodings,
                            send_compression_encodings,
                        );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/modelbox.ModelStore/LogMetrics" => {
                    #[allow(non_camel_case_types)]
                    struct LogMetricsSvc<T: ModelStore>(pub Arc<T>);
                    impl<T: ModelStore> tonic::server::UnaryService<super::LogMetricsRequest> for LogMetricsSvc<T> {
                        type Response = super::LogMetricsResponse;
                        type Future = BoxFuture<tonic::Response<Self::Response>, tonic::Status>;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::LogMetricsRequest>,
                        ) -> Self::Future {
                            let inner = self.0.clone();
                            let fut = async move { (*inner).log_metrics(request).await };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = LogMetricsSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec).apply_compression_config(
                            accept_compression_encodings,
                            send_compression_encodings,
                        );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/modelbox.ModelStore/GetMetrics" => {
                    #[allow(non_camel_case_types)]
                    struct GetMetricsSvc<T: ModelStore>(pub Arc<T>);
                    impl<T: ModelStore> tonic::server::UnaryService<super::GetMetricsRequest> for GetMetricsSvc<T> {
                        type Response = super::GetMetricsResponse;
                        type Future = BoxFuture<tonic::Response<Self::Response>, tonic::Status>;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::GetMetricsRequest>,
                        ) -> Self::Future {
                            let inner = self.0.clone();
                            let fut = async move { (*inner).get_metrics(request).await };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = GetMetricsSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec).apply_compression_config(
                            accept_compression_encodings,
                            send_compression_encodings,
                        );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/modelbox.ModelStore/LogEvent" => {
                    #[allow(non_camel_case_types)]
                    struct LogEventSvc<T: ModelStore>(pub Arc<T>);
                    impl<T: ModelStore> tonic::server::UnaryService<super::LogEventRequest> for LogEventSvc<T> {
                        type Response = super::LogEventResponse;
                        type Future = BoxFuture<tonic::Response<Self::Response>, tonic::Status>;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::LogEventRequest>,
                        ) -> Self::Future {
                            let inner = self.0.clone();
                            let fut = async move { (*inner).log_event(request).await };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = LogEventSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec).apply_compression_config(
                            accept_compression_encodings,
                            send_compression_encodings,
                        );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/modelbox.ModelStore/ListEvents" => {
                    #[allow(non_camel_case_types)]
                    struct ListEventsSvc<T: ModelStore>(pub Arc<T>);
                    impl<T: ModelStore> tonic::server::UnaryService<super::ListEventsRequest> for ListEventsSvc<T> {
                        type Response = super::ListEventsResponse;
                        type Future = BoxFuture<tonic::Response<Self::Response>, tonic::Status>;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::ListEventsRequest>,
                        ) -> Self::Future {
                            let inner = self.0.clone();
                            let fut = async move { (*inner).list_events(request).await };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = ListEventsSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec).apply_compression_config(
                            accept_compression_encodings,
                            send_compression_encodings,
                        );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/modelbox.ModelStore/WatchNamespace" => {
                    #[allow(non_camel_case_types)]
                    struct WatchNamespaceSvc<T: ModelStore>(pub Arc<T>);
                    impl<T: ModelStore>
                        tonic::server::ServerStreamingService<super::WatchNamespaceRequest>
                        for WatchNamespaceSvc<T>
                    {
                        type Response = super::WatchNamespaceResponse;
                        type ResponseStream = T::WatchNamespaceStream;
                        type Future =
                            BoxFuture<tonic::Response<Self::ResponseStream>, tonic::Status>;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::WatchNamespaceRequest>,
                        ) -> Self::Future {
                            let inner = self.0.clone();
                            let fut = async move { (*inner).watch_namespace(request).await };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let inner = inner.0;
                        let method = WatchNamespaceSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec).apply_compression_config(
                            accept_compression_encodings,
                            send_compression_encodings,
                        );
                        let res = grpc.server_streaming(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                _ => Box::pin(async move {
                    Ok(http::Response::builder()
                        .status(200)
                        .header("grpc-status", "12")
                        .header("content-type", "application/grpc")
                        .body(empty_body())
                        .unwrap())
                }),
            }
        }
    }
    impl<T: ModelStore> Clone for ModelStoreServer<T> {
        fn clone(&self) -> Self {
            let inner = self.inner.clone();
            Self {
                inner,
                accept_compression_encodings: self.accept_compression_encodings,
                send_compression_encodings: self.send_compression_encodings,
            }
        }
    }
    impl<T: ModelStore> Clone for _Inner<T> {
        fn clone(&self) -> Self {
            Self(self.0.clone())
        }
    }
    impl<T: std::fmt::Debug> std::fmt::Debug for _Inner<T> {
        fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
            write!(f, "{:?}", self.0)
        }
    }
    impl<T: ModelStore> tonic::server::NamedService for ModelStoreServer<T> {
        const NAME: &'static str = "modelbox.ModelStore";
    }
}
