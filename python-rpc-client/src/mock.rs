use tokio_stream::wrappers::ReceiverStream;
use tonic::Status;

use super::modelbox::model_store_server::ModelStore;
use super::modelbox::{
    CreateExperimentRequest, CreateExperimentResponse, CreateModelRequest, CreateModelResponse,
    CreateModelVersionRequest, CreateModelVersionResponse, DownloadFileRequest,
    DownloadFileResponse, GetExperimentRequest, GetExperimentResponse, GetMetricsRequest,
    GetMetricsResponse, ListArtifactsRequest, ListArtifactsResponse, ListEventsRequest,
    ListEventsResponse, ListExperimentsRequest, ListExperimentsResponse, ListMetadataRequest,
    ListMetadataResponse, ListModelVersionsRequest, ListModelVersionsResponse, ListModelsRequest,
    ListModelsResponse, LogEventRequest, LogEventResponse, LogMetricsRequest, LogMetricsResponse,
    TrackArtifactsRequest, TrackArtifactsResponse, UpdateMetadataRequest, UpdateMetadataResponse,
    UploadFileRequest, UploadFileResponse, WatchNamespaceRequest, WatchNamespaceResponse,
};

#[derive(Default)]
pub struct MockModelStoreServer {}

#[tonic::async_trait]
impl ModelStore for MockModelStoreServer {
    async fn create_model(
        &self,
        _request: tonic::Request<CreateModelRequest>,
    ) -> Result<tonic::Response<CreateModelResponse>, tonic::Status> {
        Ok(tonic::Response::new(CreateModelResponse::default()))
    }

    async fn list_models(
        &self,
        _request: tonic::Request<ListModelsRequest>,
    ) -> Result<tonic::Response<ListModelsResponse>, tonic::Status> {
        unimplemented!()
    }
    async fn create_model_version(
        &self,
        _request: tonic::Request<CreateModelVersionRequest>,
    ) -> Result<tonic::Response<CreateModelVersionResponse>, tonic::Status> {
        unimplemented!()
    }
    async fn list_model_versions(
        &self,
        _request: tonic::Request<ListModelVersionsRequest>,
    ) -> Result<tonic::Response<ListModelVersionsResponse>, tonic::Status> {
        unimplemented!()
    }
    async fn create_experiment(
        &self,
        _request: tonic::Request<CreateExperimentRequest>,
    ) -> Result<tonic::Response<CreateExperimentResponse>, tonic::Status> {
        unimplemented!()
    }
    async fn list_experiments(
        &self,
        _request: tonic::Request<ListExperimentsRequest>,
    ) -> Result<tonic::Response<ListExperimentsResponse>, tonic::Status> {
        unimplemented!()
    }
    async fn get_experiment(
        &self,
        _request: tonic::Request<GetExperimentRequest>,
    ) -> Result<tonic::Response<GetExperimentResponse>, tonic::Status> {
        unimplemented!()
    }
    async fn upload_file(
        &self,
        _request: tonic::Request<tonic::Streaming<UploadFileRequest>>,
    ) -> Result<tonic::Response<UploadFileResponse>, tonic::Status> {
        unimplemented!();
    }

    type DownloadFileStream = ReceiverStream<Result<DownloadFileResponse, Status>>;

    async fn download_file(
        &self,
        _request: tonic::Request<DownloadFileRequest>,
    ) -> Result<tonic::Response<Self::DownloadFileStream>, tonic::Status> {
        unimplemented!();
    }

    async fn update_metadata(
        &self,
        _request: tonic::Request<UpdateMetadataRequest>,
    ) -> Result<tonic::Response<UpdateMetadataResponse>, tonic::Status> {
        unimplemented!()
    }

    async fn list_metadata(
        &self,
        _request: tonic::Request<ListMetadataRequest>,
    ) -> Result<tonic::Response<ListMetadataResponse>, tonic::Status> {
        unimplemented!()
    }

    async fn track_artifacts(
        &self,
        _request: tonic::Request<TrackArtifactsRequest>,
    ) -> Result<tonic::Response<TrackArtifactsResponse>, tonic::Status> {
        unimplemented!()
    }

    async fn list_artifacts(
        &self,
        _request: tonic::Request<ListArtifactsRequest>,
    ) -> Result<tonic::Response<ListArtifactsResponse>, tonic::Status> {
        unimplemented!()
    }

    async fn log_metrics(
        &self,
        _request: tonic::Request<LogMetricsRequest>,
    ) -> Result<tonic::Response<LogMetricsResponse>, tonic::Status> {
        unimplemented!()
    }

    async fn get_metrics(
        &self,
        _request: tonic::Request<GetMetricsRequest>,
    ) -> Result<tonic::Response<GetMetricsResponse>, tonic::Status> {
        unimplemented!()
    }

    async fn log_event(
        &self,
        _request: tonic::Request<LogEventRequest>,
    ) -> Result<tonic::Response<LogEventResponse>, tonic::Status> {
        unimplemented!()
    }

    async fn list_events(
        &self,
        _request: tonic::Request<ListEventsRequest>,
    ) -> Result<tonic::Response<ListEventsResponse>, tonic::Status> {
        unimplemented!()
    }

    type WatchNamespaceStream = ReceiverStream<Result<WatchNamespaceResponse, Status>>;
    async fn watch_namespace(
        &self,
        _request: tonic::Request<WatchNamespaceRequest>,
    ) -> Result<tonic::Response<Self::WatchNamespaceStream>, tonic::Status> {
        unimplemented!()
    }
}
