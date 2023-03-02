use super::server_config::ServerConfig;
use object_store::aws::AmazonS3Builder;
use object_store::gcp::GoogleCloudStorageBuilder;
use object_store::local::LocalFileSystem;
use std::sync::Arc;
use tokio::signal;

pub struct Agent {
    grpc_agent: crate::grpc_server::GrpcServer,
    repository: Arc<super::repository::Repository>,
}

impl Agent {
    pub async fn new(config: ServerConfig) -> Self {
        tracing::info!("creating agent");

        tracing::info!(
            "creating object store client: {:?}",
            config.object_store.provider
        );
        let object_store = Agent::get_object_store(&config)
            .unwrap_or_else(|e| panic!("unable to create object store client {}", e));
        let grpc_agent =
            super::grpc_server::GrpcServer::new(config.grpc_listen_addr.clone(), object_store)
                .unwrap_or_else(|e| panic!("unable to create grpc server {}", e));
        let respository = super::repository::Repository::new(&config.database_url())
            .await
            .unwrap_or_else(|e| panic!("unable to create db {}", e));
        tracing::info!("finished creating agent");
        Agent {
            grpc_agent,
            repository: Arc::new(respository),
        }
    }

    pub fn get_object_store(
        server_config: &ServerConfig,
    ) -> Result<Arc<dyn object_store::ObjectStore>, Box<dyn std::error::Error>> {
        match server_config.object_store.provider {
            super::server_config::ObjectStoreProvider::S3 => {
                let s3 = AmazonS3Builder::from_env()
                    .with_bucket_name(&server_config.object_store.bucket)
                    .build()?;
                Ok(Arc::new(s3))
            }
            super::server_config::ObjectStoreProvider::Gcs => {
                let gcs = GoogleCloudStorageBuilder::from_env()
                    .with_bucket_name(&server_config.object_store.bucket)
                    .build()?;
                Ok(Arc::new(gcs))
            }
            super::server_config::ObjectStoreProvider::FileSystem => {
                let fs = LocalFileSystem::new_with_prefix(&server_config.object_store.bucket)?;
                Ok(Arc::new(fs))
            }
        }
    }

    pub async fn start(&self) -> Result<(), Box<dyn std::error::Error>> {
        tracing::info!("starting grpc server");
        let agent = &self.grpc_agent;
        agent.start(self.repository.clone()).await?;
        Ok(())
    }

    pub async fn wait_for_signal(&self) {
        match signal::ctrl_c().await {
            Ok(()) => {
                tracing::info!("received sigterm, shutting down cleanly");
            }
            Err(err) => {
                tracing::error!("unable to listen for shutdown signal: {}", err);
            }
        }
    }
}
