use serde::{Deserialize, Serialize};
use std::fs;
use std::path::PathBuf;

use thiserror::Error;

#[derive(Error, Debug)]
pub enum ConfigParsingError {
    #[error("unable to read config file")]
    IoError(#[from] std::io::Error),

    #[error("unable to de-serialize yaml")]
    DeserializationError {
        #[from]
        source: serde_yaml::Error,
    },
}

#[derive(Debug, PartialEq, Serialize, Deserialize)]
pub enum ObjectStoreProvider {
    S3,
    Gcs,
    FileSystem,
}

#[derive(Debug, PartialEq, Serialize, Deserialize)]
pub struct ObjectStoreConfig {
    pub bucket: String,
    pub provider: ObjectStoreProvider,
}

#[derive(Debug, PartialEq, Serialize, Deserialize)]
pub struct ServerConfig {
    pub grpc_listen_addr: String,
    pub database_host: String,
    pub database_name: String,
    pub database_username: String,
    pub database_password: String,
    pub object_store: ObjectStoreConfig,
}

impl ServerConfig {
    pub fn database_url(&self) -> String {
        format!(
            "postgres://{}/{}?user={}&password={}",
            &self.database_host,
            &self.database_name,
            &self.database_username,
            &self.database_password
        )
    }
}

impl Default for ServerConfig {
    fn default() -> Self {
        Self {
            grpc_listen_addr: "127.0.0.1:8085".into(),
            database_host: "localhost:5432".into(),
            database_name: "tensorland".into(),
            database_username: "postgres".into(),
            database_password: "foo".into(),
            object_store: ObjectStoreConfig {
                bucket: "/tmp/modelbox/".into(),
                provider: ObjectStoreProvider::FileSystem,
            },
        }
    }
}

impl ServerConfig {
    pub fn from_path(path: PathBuf) -> Result<Self, ConfigParsingError> {
        let yaml = fs::read_to_string(path.as_path())?;
        ServerConfig::from_str(yaml)
    }

    fn from_str(content: String) -> Result<ServerConfig, ConfigParsingError> {
        let config: Result<ServerConfig, ConfigParsingError> = serde_yaml::from_str(&content)
            .map_err(|e| ConfigParsingError::DeserializationError { source: e });
        config
    }

    pub fn generate_config(path: PathBuf) -> Result<(), ConfigParsingError> {
        let config = ServerConfig::default();
        let str = serde_yaml::to_string(&config)
            .map_err(|e| ConfigParsingError::DeserializationError { source: e })?;
        std::fs::write(path.as_path(), str)?;
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use indoc::indoc;

    use super::{ConfigParsingError, ServerConfig};

    #[test]
    fn invalid_config_path() {
        let config = ServerConfig::from_path("/invalid/path".into());
        assert!(matches!(config, Err(ConfigParsingError::IoError(_))));
    }

    #[test]
    fn valid_yaml() {
        let valid_config = indoc! {r#"
            ---
            grpc_listen_addr: "127.0.0.1:9089"
            database_host: "localhost:5234"
            database_name: "tensorland"
            database_username: "postgres"
            database_password: "foo"
            object_store:
                bucket: "/tmp/modelbox/"
                provider: FileSystem
        "#};
        let config = ServerConfig::from_str(valid_config.into()).unwrap();
        assert_eq!(
            config.database_url(),
            "postgres://localhost:5234/tensorland?user=postgres&password=foo"
        );
        assert_eq!(config.grpc_listen_addr, "127.0.0.1:9089");
        assert!(matches!(
            config.object_store.provider,
            super::ObjectStoreProvider::FileSystem
        ));
        assert_eq!(config.object_store.bucket, "/tmp/modelbox/")
    }

    #[test]
    fn invalid_yaml() {}
}
