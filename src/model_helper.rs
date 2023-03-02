use std::collections::{hash_map::DefaultHasher, HashMap};

use prost_types::Timestamp;
use std::hash::Hasher;
use thiserror::Error;
use time::{Duration, OffsetDateTime, PrimitiveDateTime};

use super::modelbox;

#[derive(Error, Debug)]
pub enum InvalidRequestError {
    #[error("Invalid time")]
    InvalidTime { source: time::error::ComponentRange },

    #[error("deserialization error")]
    DeserializationError(#[from] serde_json::Error),

    #[error("missing field: {field}")]
    MissingField { field: String },
}

fn now() -> PrimitiveDateTime {
    let n = OffsetDateTime::now_utc();
    PrimitiveDateTime::new(n.date(), n.time())
}

fn to_primtive_time(ts: Timestamp) -> Result<PrimitiveDateTime, InvalidRequestError> {
    let off_dt = OffsetDateTime::from_unix_timestamp(ts.seconds)
        .map_err(|source| InvalidRequestError::InvalidTime { source })?
        + Duration::nanoseconds(ts.nanos as i64);
    Ok(PrimitiveDateTime::new(off_dt.date(), off_dt.time()))
}

impl modelbox::CreateExperimentRequest {
    pub fn into_model(self) -> entity::experiments::Model {
        entity::experiments::Model {
            id: self.generate_id(),
            name: self.name,
            external_id: self.external_id,
            owner: self.owner,
            namespace: self.namespace,
            ml_framework: self.framework,
            created_at: now(),
            updated_at: now(),
        }
    }

    fn generate_id(&self) -> String {
        let mut hasher = DefaultHasher::new();
        hasher.write(self.name.as_bytes());
        hasher.write(self.owner.as_bytes());
        hasher.write(self.namespace.as_bytes());
        hasher.finish().to_string()
    }
}

impl modelbox::CreateModelRequest {
    pub fn into_model(self) -> entity::models::Model {
        entity::models::Model {
            id: self.generate_id(),
            name: self.name,
            owner: self.owner,
            namespace: self.namespace,
            task: self.task,
            description: self.description,
            created_at: now(),
            updated_at: now(),
        }
    }

    fn generate_id(&self) -> String {
        let mut hasher = DefaultHasher::new();
        hasher.write(self.name.as_bytes());
        hasher.write(self.namespace.as_bytes());
        hasher.finish().to_string()
    }
}

impl modelbox::CreateModelVersionRequest {
    pub fn into_model_version(self) -> Result<entity::model_versions::Model, serde_json::Error> {
        let json = serde_json::to_value(&self.unique_tags)?;
        Ok(entity::model_versions::Model {
            id: self.generate_id(),
            name: self.name,
            model_id: self.model,
            experiment_id: "".into(),
            version: self.version,
            namespace: self.namespace,
            description: self.description,
            ml_framework: self.framework,
            unique_tags: json,
            created_at: now(),
            updated_at: now(),
        })
    }

    fn generate_id(&self) -> String {
        let mut hasher = DefaultHasher::new();
        hasher.write(self.model.as_bytes());
        hasher.write(self.version.as_bytes());
        hasher.finish().to_string()
    }
}

impl modelbox::UpdateMetadataRequest {
    pub fn into_metadata_model(self) -> Result<Vec<entity::metadata::Model>, serde_json::Error> {
        let mut models: Vec<entity::metadata::Model> = Vec::new();
        match &self.metadata {
            Some(meta) => {
                for (k, v) in &meta.metadata {
                    let json = serde_json::to_value(v)?;
                    models.push(entity::metadata::Model {
                        id: self.generate_id(k),
                        parent_id: self.parent_id.clone(),
                        name: k.clone(),
                        meta: json,
                        created_at: now(),
                        updated_at: now(),
                    })
                }
            }
            None => {}
        }
        Ok(models)
    }

    fn generate_id(&self, key: &str) -> String {
        let mut hasher = DefaultHasher::new();
        hasher.write(key.as_bytes());
        hasher.write(self.parent_id.as_bytes());
        hasher.finish().to_string()
    }
}

pub fn from_timestamp(ts: PrimitiveDateTime) -> Option<Timestamp> {
    Some(Timestamp {
        seconds: ts.second() as i64,
        nanos: ts.nanosecond() as i32,
    })
}

impl modelbox::Experiment {
    pub fn from_model(model: entity::experiments::Model) -> Self {
        Self {
            id: model.id,
            name: model.name,
            external_id: model.external_id,
            owner: model.owner,
            namespace: model.namespace,
            framework: model.ml_framework,
            created_at: from_timestamp(model.created_at),
            updated_at: from_timestamp(model.updated_at),
        }
    }
}

impl modelbox::Model {
    pub fn from_model(model: entity::models::Model) -> Self {
        Self {
            id: model.id,
            name: model.name,
            owner: model.owner,
            namespace: model.namespace,
            description: model.description,
            task: model.task,
            created_at: from_timestamp(model.created_at),
            updated_at: from_timestamp(model.updated_at),
        }
    }
}

impl modelbox::ModelVersion {
    pub fn from_model(model: entity::model_versions::Model) -> Result<Self, serde_json::Error> {
        let tags = serde_json::from_value(model.unique_tags)?;
        Ok(Self {
            id: model.id,
            name: model.name,
            model_id: model.model_id,
            version: model.version,
            description: model.description,
            framework: model.ml_framework,
            unique_tags: tags,
            created_at: from_timestamp(model.created_at),
            updated_at: from_timestamp(model.updated_at),
        })
    }
}

impl modelbox::Metadata {
    pub fn from_model(model: Vec<entity::metadata::Model>) -> Result<Self, serde_json::Error> {
        let mut meta: HashMap<String, String> = HashMap::new();
        for m in model {
            let value = serde_json::from_value(m.meta)?;
            meta.insert(m.name, value);
        }
        Ok(Self { metadata: meta })
    }
}

impl modelbox::FileMetadata {
    pub fn from_models(model: Vec<entity::files::Model>) -> Result<Vec<Self>, serde_json::Error> {
        let mut meta: Vec<Self> = Vec::new();
        for m in model {
            let value: HashMap<String, String> = serde_json::from_value(m.metadata)?;
            let checksum = value.get("checksum").map_or("", String::as_str);
            meta.push(Self {
                id: m.id,
                parent_id: m.parent_id,
                checksum: checksum.to_string(),
                src_path: m.src_path,
                upload_path: m.upload_path.unwrap_or("".into()),
                file_type: modelbox::FileType::to_file_meta(&m.file_type) as i32,
                created_at: from_timestamp(m.created_at),
                updated_at: from_timestamp(m.updated_at),
            });
        }
        Ok(meta)
    }

    pub fn into_file_metadata_model(
        &self,
        artifact_name: String,
    ) -> Result<entity::files::Model, serde_json::Error> {
        let mut meta: HashMap<String, String> = HashMap::new();
        meta.insert("checksum".to_string(), self.checksum.clone());
        let json = serde_json::to_value(&meta)?;
        let f_type = self.r#file_type().as_string();
        Ok(entity::files::Model {
            id: self.generate_id(),
            parent_id: self.parent_id.clone(),
            src_path: self.src_path.clone(),
            upload_path: Some(self.upload_path.clone()),
            file_type: f_type,
            artifact_name: artifact_name.clone(),
            artifact_id: self.generate_artifact_id(artifact_name.clone()),
            metadata: json,
            created_at: now(),
            updated_at: now(),
        })
    }

    fn generate_id(&self) -> String {
        let mut hasher = DefaultHasher::new();
        hasher.write(self.parent_id.as_bytes());
        hasher.write(self.src_path.as_bytes());
        hasher.write(self.checksum.as_bytes());
        hasher.write(self.r#file_type().as_string().as_bytes());
        if let Some(ts) = self.created_at.as_ref() {
            hasher.write(ts.seconds.to_string().as_bytes());
            hasher.write(ts.nanos.to_string().as_bytes());
        }
        if let Some(ts) = self.updated_at.as_ref() {
            hasher.write(ts.seconds.to_string().as_bytes());
            hasher.write(ts.nanos.to_string().as_bytes());
        }
        hasher.finish().to_string()
    }

    fn generate_artifact_id(&self, artifact_name: String) -> String {
        let mut hasher = DefaultHasher::new();
        hasher.write(self.parent_id.as_bytes());
        hasher.write(artifact_name.as_bytes());
        hasher.finish().to_string()
    }
}

impl modelbox::UploadFileMetadata {
    pub fn file_model(&self) -> Result<entity::files::Model, InvalidRequestError> {
        if let Some(res) = self
            .metadata
            .as_ref()
            .map(|m| m.into_file_metadata_model(self.artifact_name.clone()))
        {
            res.map_err(InvalidRequestError::DeserializationError)
        } else {
            Err(InvalidRequestError::MissingField {
                field: "metadata".into(),
            })
        }
    }
}

impl modelbox::FileType {
    pub fn to_file_meta(s: &str) -> Self {
        match s {
            "checkpoint" => modelbox::FileType::Checkpoint,
            "model" => modelbox::FileType::Model,
            "text" => modelbox::FileType::Text,
            "image" => modelbox::FileType::Image,
            "video" => modelbox::FileType::Video,
            "audio" => modelbox::FileType::Audio,
            _ => modelbox::FileType::Undefined,
        }
    }

    pub fn as_string(&self) -> String {
        match &self {
            Self::Checkpoint => "checkpoint".into(),
            Self::Model => "model".into(),
            Self::Text => "text".into(),
            Self::Image => "image".into(),
            Self::Video => "video".into(),
            Self::Audio => "audio".into(),
            Self::Undefined => "undefined".into(),
        }
    }
}

impl modelbox::LogEventRequest {
    pub fn into_log_event_model(self) -> Result<entity::events::Model, InvalidRequestError> {
        let id = self.generate_id();
        let event = self.event.unwrap_or(modelbox::Event::default());
        let event_source = event.source.unwrap_or(modelbox::EventSource::default());
        let metadata = event.metadata.unwrap_or(modelbox::Metadata::default());
        let json = serde_json::to_value(metadata.metadata)?;
        let event_ts = event.wallclock_time.unwrap_or_else(|| {
            let n = now();
            Timestamp {
                seconds: n.second() as i64,
                nanos: n.nanosecond() as i32,
            }
        });
        let w_clock = to_primtive_time(event_ts)?;

        Ok(entity::events::Model {
            id,
            parent_id: self.parent_id,
            name: event.name,
            source: event_source.name,
            metadata: json,
            source_wall_clock: w_clock,
        })
    }

    fn generate_id(&self) -> String {
        let mut hasher = DefaultHasher::new();
        hasher.write(self.parent_id.as_bytes());
        if let Some(e) = self.event.as_ref() {
            hasher.write(e.name.as_bytes());
            if let Some(w_clock) = &e.wallclock_time {
                hasher.write(w_clock.seconds.to_string().as_bytes());
                hasher.write(w_clock.nanos.to_string().as_bytes());
            }
            if let Some(s) = e.source.as_ref() {
                hasher.write(s.name.as_bytes());
            }
        }
        hasher.finish().to_string()
    }
}

impl modelbox::Event {
    pub fn from_model(model: entity::events::Model) -> Self {
        let metadata = modelbox::Metadata {
            metadata: serde_json::from_value(model.metadata).unwrap_or(HashMap::new()),
        };
        let source = modelbox::EventSource { name: model.source };
        Self {
            name: model.name,
            source: Some(source),
            metadata: Some(metadata),
            wallclock_time: from_timestamp(model.source_wall_clock),
        }
    }
}

impl modelbox::LogEventResponse {
    pub fn new() -> Self {
        Self {
            created_at: from_timestamp(now()),
        }
    }
}

impl modelbox::LogMetricsRequest {
    pub fn into_metric_model(self) -> entity::metrics::Model {
        entity::metrics::Model {
            id: 0,
            object_id: self.parent_id,
            name: self.key,
            tensor: self
                .value
                .clone()
                .and_then(|v| {
                    v.value.map(|v| {
                        if let modelbox::metrics_value::Value::STensor(s_tensor) = v {
                            Some(s_tensor)
                        } else {
                            None
                        }
                    })
                })
                .flatten(),
            double_value: self
                .value
                .clone()
                .and_then(|v| {
                    v.value.map(|v| {
                        if let modelbox::metrics_value::Value::FVal(f_val) = v {
                            Some(f_val as f64)
                        } else {
                            None
                        }
                    })
                })
                .flatten(),
            step: self.value.map(|v| v.step as i64),
            wall_clock: None,
            created_at: now(),
        }
    }
}

impl modelbox::MetricsValue {
    pub fn from_metrics(m: entity::metrics::Model) -> Self {
        modelbox::MetricsValue {
            step: m.step.unwrap_or(0) as u64,
            wallclock_time: 0_64,
            value: Some(modelbox::metrics_value::Value::FVal(
                m.double_value.unwrap_or(0.0) as f32,
            )),
        }
    }
}

#[cfg(test)]
mod tests {
    #[test]
    fn test_uniqu_tags() {
        use indoc::indoc;
        let valid_tags_json = indoc! {r#"["a", "b", "c"]"#};
        let tags: Vec<String> = serde_json::from_str(valid_tags_json).unwrap();
        assert_eq!(vec!["a", "b", "c"], tags);
    }
}
