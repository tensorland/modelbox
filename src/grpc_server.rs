use std::collections::HashMap;
use std::net::SocketAddr;
use std::sync::Arc;
use tokio::io::AsyncWriteExt;

use object_store::{path::Path, ObjectStore};
use tokio_stream::wrappers::ReceiverStream;
use tonic::transport::Server;
use tonic::{Request, Response, Status};

use crate::modelbox::MetricsValue;

use super::modelbox::model_store_server::{ModelStore, ModelStoreServer};
use super::modelbox::upload_file_request;
use super::modelbox::{
    Artifact, CreateExperimentRequest, CreateExperimentResponse, CreateModelRequest,
    CreateModelResponse, CreateModelVersionRequest, CreateModelVersionResponse,
    DownloadFileRequest, DownloadFileResponse, Event, Experiment, FileMetadata,
    GetExperimentRequest, GetExperimentResponse, GetMetricsRequest, GetMetricsResponse,
    ListArtifactsRequest, ListArtifactsResponse, ListEventsRequest, ListEventsResponse,
    ListExperimentsRequest, ListExperimentsResponse, ListMetadataRequest, ListMetadataResponse,
    ListModelVersionsRequest, ListModelVersionsResponse, ListModelsRequest, ListModelsResponse,
    LogEventRequest, LogEventResponse, LogMetricsRequest, LogMetricsResponse, Metadata, Metrics,
    Model, ModelVersion, TrackArtifactsRequest, TrackArtifactsResponse, UpdateMetadataRequest,
    UpdateMetadataResponse, UploadFileRequest, UploadFileResponse, WatchNamespaceRequest,
    WatchNamespaceResponse,
};

use super::model_helper;

const SERVICE_DESCRIPTOR_SET: &[u8] = tonic::include_file_descriptor_set!("service_descriptor");

#[derive(Debug)]
pub struct ModelBoxService {
    repository: Arc<super::repository::Repository>,
    object_store: Arc<dyn object_store::ObjectStore>,
}

#[tonic::async_trait]
impl ModelStore for ModelBoxService {
    async fn create_model(
        &self,
        request: Request<CreateModelRequest>,
    ) -> Result<Response<CreateModelResponse>, Status> {
        let model = request.into_inner().into_model();
        self.repository
            .create_model(model.clone())
            .await
            .map_err(|e| Status::internal(e.to_string()))
            .map(|result| {
                Response::new(CreateModelResponse {
                    id: model.id.clone(),
                    exists: result.exists,
                    created_at: model_helper::from_timestamp(model.created_at),
                    updated_at: model_helper::from_timestamp(model.updated_at),
                })
            })
    }

    async fn list_models(
        &self,
        request: Request<ListModelsRequest>,
    ) -> Result<Response<ListModelsResponse>, Status> {
        let namespace = request.into_inner().namespace;
        self.repository
            .models_by_namespace(namespace)
            .await
            .map_or_else(
                |e| Err(Status::internal(e.to_string())),
                |models| {
                    let response = Response::new(ListModelsResponse {
                        models: models.into_iter().map(Model::from_model).collect(),
                    });
                    Ok(response)
                },
            )
    }
    /// Creates a new model version for a model
    async fn create_model_version(
        &self,
        request: Request<CreateModelVersionRequest>,
    ) -> Result<Response<CreateModelVersionResponse>, Status> {
        let model_version = request.into_inner().into_model_version();
        if model_version.is_err() {
            return Err(Status::invalid_argument(
                model_version.err().unwrap().to_string(),
            ));
        }
        let model_version = model_version.unwrap();
        let result = self
            .repository
            .create_model_version(model_version.clone())
            .await
            .map_or_else(
                |e| Err(Status::internal(e.to_string())),
                |result| {
                    let response = Response::new(CreateModelVersionResponse {
                        model_version: model_version.id.clone(),
                        exists: result.exists,
                        created_at: model_helper::from_timestamp(model_version.created_at),
                        updated_at: model_helper::from_timestamp(model_version.updated_at),
                    });
                    Ok(response)
                },
            );
        result
    }

    async fn list_model_versions(
        &self,
        _request: Request<ListModelVersionsRequest>,
    ) -> Result<Response<ListModelVersionsResponse>, Status> {
        let model_id = _request.into_inner().model;
        let result = self.repository.model_versions_for_model(model_id).await;
        if result.is_err() {
            return Err(Status::internal(result.err().unwrap().to_string()));
        }
        let model_versions = result.unwrap();
        let mut resp: Vec<ModelVersion> = Vec::new();
        for model_version in model_versions.iter() {
            let maybe_mv = ModelVersion::from_model(model_version.clone());
            if maybe_mv.is_err() {
                return Err(Status::internal(maybe_mv.err().unwrap().to_string()));
            }
            resp.push(maybe_mv.unwrap());
        }
        Ok(Response::new(ListModelVersionsResponse {
            model_versions: resp,
        }))
    }

    async fn create_experiment(
        &self,
        request: Request<CreateExperimentRequest>,
    ) -> Result<Response<CreateExperimentResponse>, Status> {
        let experiment = request.into_inner().into_model();
        let result = self
            .repository
            .create_exeperiment(experiment.clone())
            .await
            .map_or_else(
                |e| Err(Status::internal(e.to_string())),
                |result| {
                    let response = Response::new(CreateExperimentResponse {
                        experiment_id: result.id,
                        experiment_exists: result.exists,
                        created_at: model_helper::from_timestamp(experiment.created_at),
                        updated_at: model_helper::from_timestamp(experiment.updated_at),
                    });
                    Ok(response)
                },
            );
        result
    }

    async fn list_experiments(
        &self,
        request: Request<ListExperimentsRequest>,
    ) -> Result<Response<ListExperimentsResponse>, Status> {
        let namespace = request.into_inner().namespace;
        self.repository
            .list_experiments(namespace)
            .await
            .map_or_else(
                |e| Err(Status::internal(e.to_string())),
                |experiments| {
                    let response = Response::new(ListExperimentsResponse {
                        experiments: experiments
                            .into_iter()
                            .map(Experiment::from_model)
                            .collect(),
                    });
                    Ok(response)
                },
            )
    }

    async fn get_experiment(
        &self,
        request: Request<GetExperimentRequest>,
    ) -> Result<Response<GetExperimentResponse>, Status> {
        let id = request.into_inner().id;
        self.repository.get_experiment(&id).await.map_or_else(
            |e| Err(Status::internal(e.to_string())),
            {
                |experiment| match experiment {
                    Some(ex) => Ok(Response::new(GetExperimentResponse {
                        experiment: Some(Experiment::from_model(ex)),
                    })),
                    None => Err(Status::not_found("Experiment not found")),
                }
            },
        )
    }

    /// UploadFile streams a files to ModelBox and stores the binaries to the condfigured storage
    async fn upload_file(
        &self,
        request: Request<tonic::Streaming<UploadFileRequest>>,
    ) -> Result<Response<UploadFileResponse>, Status> {
        let mut stream = request.into_inner();
        let req = stream.message().await?;
        if req.is_none() {
            return Err(Status::invalid_argument("No metadata provided"));
        }
        if req.as_ref().unwrap().stream_frame.is_none() {
            return Err(Status::invalid_argument("No metadata frame provided"));
        }
        let mut file_id: Option<String> = None;
        let mut artifact_id: Option<String> = None;
        let mut path: Option<Path> = None;
        let mut file_model: Option<entity::files::Model> = None;
        let meta_frame = req.unwrap().stream_frame.unwrap();
        if let upload_file_request::StreamFrame::Metadata(metadata) = meta_frame {
            let try_file_model = metadata.file_model();
            if let Ok(file_metadata_model) = try_file_model {
                file_model = Some(file_metadata_model);
                file_id = file_model.as_ref().map(|m| m.id.clone());
                artifact_id = file_model.as_ref().map(|m| m.artifact_id.clone());
            } else {
                return Err(Status::invalid_argument(format!(
                    "Invalid metadata provided {}",
                    try_file_model.err().unwrap()
                )));
            }
            path = Some(
                format!(
                    "modelbox/artifacts/{}/{}",
                    metadata.object_id,
                    file_id.clone().unwrap(),
                )
                .into(),
            );
            file_model
                .as_mut()
                .map(|m| m.upload_path = path.clone().map(|p| p.to_string()));
        }
        if let Err(err) = self
            .repository
            .create_files(vec![file_model.clone().unwrap()])
            .await
        {
            return Err(Status::internal(err.to_string()));
        }
        let handle = self.object_store.put_multipart(&path.unwrap()).await;
        if handle.is_err() {
            return Err(Status::internal(handle.err().unwrap().to_string()));
        }

        let (_id, mut writer) = handle.unwrap();

        while let Some(req) = stream.message().await? {
            if let upload_file_request::StreamFrame::Chunks(data) = req.stream_frame.unwrap() {
                if let Err(e) = writer.write_all(&data).await {
                    return Err(Status::internal(e.to_string()));
                }
            }
        }
        if let Err(e) = writer.flush().await {
            return Err(Status::internal(e.to_string()));
        }
        if let Err(e) = writer.shutdown().await {
            return Err(Status::internal(e.to_string()));
        }

        Ok(Response::new(UploadFileResponse {
            file_id: file_id.unwrap(),
            artifact_id: artifact_id.unwrap(),
        }))
    }
    /// Server streaming response type for the DownloadFile method.
    type DownloadFileStream = ReceiverStream<Result<DownloadFileResponse, Status>>;
    /// DownloadFile downloads a file from configured storage
    async fn download_file(
        &self,
        _request: Request<DownloadFileRequest>,
    ) -> Result<Response<Self::DownloadFileStream>, Status> {
        unimplemented!();
    }

    async fn update_metadata(
        &self,
        request: Request<UpdateMetadataRequest>,
    ) -> Result<Response<UpdateMetadataResponse>, Status> {
        let metadata = request.into_inner().into_metadata_model();
        if metadata.is_err() {
            return Err(Status::invalid_argument(
                metadata.err().unwrap().to_string(),
            ));
        }
        let meta = metadata.unwrap();
        self.repository.update_metadata(meta).await.map_or_else(
            |e| Err(Status::internal(e.to_string())),
            |_| {
                let response = Response::new(UpdateMetadataResponse {});
                Ok(response)
            },
        )
    }

    async fn list_metadata(
        &self,
        request: Request<ListMetadataRequest>,
    ) -> Result<Response<ListMetadataResponse>, Status> {
        let parent_id = request.into_inner().parent_id;
        self.repository.get_metadata(parent_id).await.map_or_else(
            |e| Err(Status::internal(e.to_string())),
            |metadata| {
                let meta =
                    Metadata::from_model(metadata).map_err(|e| Status::internal(e.to_string()))?;
                let response = Response::new(ListMetadataResponse {
                    metadata: Some(meta),
                });
                Ok(response)
            },
        )
    }
    /// Tracks a set of artifacts with a experiment/checkpoint/model
    async fn track_artifacts(
        &self,
        request: Request<TrackArtifactsRequest>,
    ) -> Result<Response<TrackArtifactsResponse>, Status> {
        let request = request.into_inner();
        let artifact_name = request.name;
        let file_models: Result<Vec<entity::files::Model>, serde_json::Error> = request
            .files
            .into_iter()
            .map(|f| f.into_file_metadata_model(artifact_name.clone()))
            .collect();
        if file_models.is_err() {
            return Err(Status::invalid_argument(
                file_models.err().unwrap().to_string(),
            ));
        }
        let files = file_models.unwrap();
        self.repository.create_files(files).await.map_or_else(
            |e| Err(Status::internal(e.to_string())),
            |_| {
                let response = Response::new(TrackArtifactsResponse { id: "".to_string() });
                Ok(response)
            },
        )
    }

    async fn list_artifacts(
        &self,
        request: Request<ListArtifactsRequest>,
    ) -> Result<Response<ListArtifactsResponse>, Status> {
        let try_files = self
            .repository
            .get_files(request.into_inner().object_id)
            .await;

        if let Err(e) = try_files {
            return Err(Status::internal(e.to_string()));
        }

        // Artifact ID, Name, Parent ID -> Files
        let mut artifacts_by_name: HashMap<(String, String, String), Vec<entity::files::Model>> =
            HashMap::new();
        try_files.unwrap().into_iter().for_each(|f| {
            let artifact = (
                f.artifact_id.clone(),
                f.artifact_name.clone(),
                f.parent_id.clone(),
            );
            artifacts_by_name.entry(artifact).or_insert(vec![]).push(f);
        });
        let mut resp = ListArtifactsResponse { artifacts: vec![] };
        for (name, files) in artifacts_by_name.into_iter() {
            let try_artifact_assets = FileMetadata::from_models(files);

            if let Ok(file_metadata) = try_artifact_assets {
                let artifact = Artifact {
                    id: name.0,
                    name: name.1,
                    object_id: name.2,
                    files: file_metadata,
                };
                resp.artifacts.push(artifact);
            } else if let Err(e) = try_artifact_assets {
                return Err(Status::internal(format!(
                    "Error while parsing assets: {}",
                    e
                )));
            }
        }
        Ok(Response::new(resp))
    }

    async fn log_metrics(
        &self,
        request: Request<LogMetricsRequest>,
    ) -> Result<Response<LogMetricsResponse>, Status> {
        let metrics = request.into_inner().into_metric_model();
        self.repository
            .log_metrics(vec![metrics])
            .await
            .map_or_else(
                |e| Err(Status::internal(e.to_string())),
                |_| Ok(Response::new(LogMetricsResponse {})),
            )
    }

    async fn get_metrics(
        &self,
        _request: Request<GetMetricsRequest>,
    ) -> Result<Response<GetMetricsResponse>, Status> {
        self.repository
            .metrics(_request.into_inner().parent_id)
            .await
            .map_or_else(
                |e| Err(Status::internal(e.to_string())),
                |metrics| {
                    let mut m: HashMap<String, Metrics> = HashMap::new();
                    metrics.into_iter().for_each(|metric| {
                        let key = metric.name.clone();
                        let metric_velue = MetricsValue::from_metrics(metric);
                        m.entry(key.clone())
                            .or_insert(Metrics {
                                key,
                                values: vec![],
                            })
                            .values
                            .push(metric_velue);
                    });
                    let response = Response::new(GetMetricsResponse { metrics: m });
                    Ok(response)
                },
            )
    }

    async fn log_event(
        &self,
        request: Request<LogEventRequest>,
    ) -> Result<Response<LogEventResponse>, Status> {
        let maybe_event = request.into_inner().into_log_event_model();
        if maybe_event.is_err() {
            return Err(Status::invalid_argument(
                maybe_event.err().unwrap().to_string(),
            ));
        }
        let event = maybe_event.unwrap();
        self.repository
            .create_events(vec![event])
            .await
            .map_or_else(
                |e| Err(Status::internal(e.to_string())),
                |_| {
                    let response = Response::new(LogEventResponse::new());
                    Ok(response)
                },
            )
    }

    async fn list_events(
        &self,
        request: Request<ListEventsRequest>,
    ) -> Result<Response<ListEventsResponse>, Status> {
        let object_id = request.into_inner().parent_id;
        self.repository
            .events_for_object(object_id)
            .await
            .map_or_else(
                |e| Err(Status::internal(e.to_string())),
                |events| {
                    let events = events
                        .into_iter()
                        .map(Event::from_model)
                        .collect::<Vec<Event>>();
                    Ok(Response::new(ListEventsResponse { events }))
                },
            )
    }

    type WatchNamespaceStream = ReceiverStream<Result<WatchNamespaceResponse, Status>>;
    async fn watch_namespace(
        &self,
        _request: Request<WatchNamespaceRequest>,
    ) -> Result<Response<Self::WatchNamespaceStream>, Status> {
        unimplemented!();
    }
}

pub struct GrpcServer {
    addr: SocketAddr,
    object_store: Arc<dyn ObjectStore>,
}

impl GrpcServer {
    pub fn new(
        addr: String,
        obj_store: Arc<dyn ObjectStore>,
    ) -> Result<Self, Box<dyn std::error::Error>> {
        let sock_addr = addr.parse()?;
        Ok(Self {
            addr: sock_addr,
            object_store: obj_store,
        })
    }

    pub async fn start(
        &self,
        repository: Arc<super::repository::Repository>,
    ) -> Result<(), Box<dyn std::error::Error>> {
        tracing::info!("Starting gRPC server at {}", self.addr);
        let srvr = ModelStoreServer::new(ModelBoxService {
            repository,
            object_store: self.object_store.clone(),
        });

        let reflection_server = tonic_reflection::server::Builder::configure()
            .register_encoded_file_descriptor_set(SERVICE_DESCRIPTOR_SET)
            .build()?;

        Server::builder()
            .add_service(srvr)
            .add_service(reflection_server)
            .serve(self.addr)
            .await?;
        Ok(())
    }
}
