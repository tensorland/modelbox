use pyo3::prelude::*;
use std::collections::HashMap;

use super::modelbox;

#[pyclass]
#[derive(Default, Clone)]
pub struct CreateExperiment {
    #[pyo3(get, set)]
    pub name: String,

    #[pyo3(get, set)]
    pub namespace: String,

    #[pyo3(get, set)]
    pub owner: String,

    #[pyo3(get, set)]
    pub ml_framework: i32,

    #[pyo3(get, set)]
    pub external_id: String,

    #[pyo3(get, set)]
    pub task: String,
}

#[pymethods]
impl CreateExperiment {
    #[new]
    pub fn new(
        name: String,
        namespace: String,
        owner: String,
        framework: i32,
        external_id: String,
        task: String,
    ) -> Self {
        Self {
            name: name,
            namespace: namespace,
            owner: owner,
            ml_framework: framework,
            external_id: external_id,
            task: task,
        }
    }
}

#[pyclass]
#[derive(Default, Clone)]
pub struct CreateModel {
    #[pyo3(get, set)]
    pub name: String,

    #[pyo3(get, set)]
    pub namespace: String,

    #[pyo3(get, set)]
    pub owner: String,

    #[pyo3(get, set)]
    pub description: String,

    #[pyo3(get, set)]
    pub task: String,
}

#[pymethods]
impl CreateModel {
    #[new]
    pub fn new(
        name: String,
        namespace: String,
        owner: String,
        description: String,
        task: String,
    ) -> Self {
        Self {
            name: name,
            namespace: namespace,
            owner: owner,
            description: description,
            task: task,
        }
    }
}

#[pyclass]
#[derive(Default, Clone)]
pub struct CreateModelResult {
    #[pyo3(get)]
    pub id: String,
}

#[pyclass]
#[derive(Default, Clone)]
pub struct CreateModelVersion {
    #[pyo3(get, set)]
    pub name: String,

    #[pyo3(get, set)]
    pub model_id: String,

    #[pyo3(get, set)]
    pub namespace: String,

    #[pyo3(get, set)]
    pub owner: String,

    #[pyo3(get, set)]
    pub description: String,

    #[pyo3(get, set)]
    pub ml_framework: MLFramework,

    #[pyo3(get, set)]
    pub version: String,

    #[pyo3(get, set)]
    pub tags: Vec<String>,
}

#[pymethods]
impl CreateModelVersion {
    #[new]
    pub fn new(
        name: String,
        model_id: String,
        namespace: String,
        owner: String,
        description: String,
        framework: MLFramework,
        version: String,
        tags: Vec<String>,
    ) -> Self {
        Self {
            name: name,
            model_id: model_id,
            namespace: namespace,
            owner: owner,
            description: description,
            ml_framework: framework,
            version: version,
            tags: tags,
        }
    }
}

#[pyclass]
#[derive(Default, Clone)]
pub struct CreateModelVersionResult {
    #[pyo3(get)]
    pub id: String,
}

#[pyclass]
#[derive(Default, Clone)]
pub struct LogEvent {
    #[pyo3(get, set)]
    pub object_id: String,

    #[pyo3(get, set)]
    pub name: String,

    #[pyo3(get, set)]
    pub source: String,

    #[pyo3(get, set)]
    pub timestamp: i64,

    #[pyo3(get, set)]
    pub metadata: HashMap<String, String>,
}

#[pymethods]
impl LogEvent {
    #[new]
    pub fn new(
        object_id: String,
        name: String,
        source: String,
        timestamp: i64,
        metadata: HashMap<String, String>,
    ) -> Self {
        Self {
            object_id: object_id,
            name: name,
            source: source,
            timestamp: timestamp,
            metadata: metadata,
        }
    }
}

#[pyclass]
pub struct LogEventResult {}

#[pyclass]
pub struct CreateExpeirmentResult {
    #[pyo3(get)]
    pub id: String,
}

#[pyclass]
#[derive(Clone, Copy)]
pub enum MLFramework {
    Pytorch = 0,
    Tensorflow = 1,
    MXNet = 2,
    XGBoost = 3,
    LightGBM = 4,
    Sklearn = 5,
    H2O = 6,
    SparkML = 7,
    CatBoost = 8,
    Keras = 9,
    Other = 10,
}

impl MLFramework {
    pub fn to_proto(&self) -> modelbox::MlFramework {
        modelbox::MlFramework::from_i32(*self as i32).unwrap_or(modelbox::MlFramework::Unknown)
    }
}

impl Default for MLFramework {
    fn default() -> Self {
        MLFramework::Other
    }
}

#[pyclass]
pub enum ArtifactMime {
    Unknown = 0,
    ModelVersion = 1,
    Checkpoint = 2,
    Text = 3,
    Image = 4,
    Video = 5,
    Audio = 6,
}

pub(crate) fn register(_py: Python<'_>, m: &PyModule) -> PyResult<()> {
    m.add_class::<CreateExpeirmentResult>()?;
    m.add_class::<CreateExperiment>()?;
    m.add_class::<CreateModel>()?;
    m.add_class::<CreateModelResult>()?;
    m.add_class::<CreateModelVersion>()?;
    m.add_class::<CreateModelVersionResult>()?;
    m.add_class::<MLFramework>()?;
    m.add_class::<ArtifactMime>()?;
    m.add_class::<LogEvent>()?;
    m.add_class::<LogEventResult>()?;
    Ok(())
}
