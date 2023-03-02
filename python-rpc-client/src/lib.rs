use pyo3::prelude::*;
use std::fmt;

mod modelbox;

use modelbox::model_store_client::ModelStoreClient;
use modelbox::{CreateExperimentRequest, GetExperimentRequest};

mod api_structs;

mod mock;

#[pyclass]
#[derive(Default)]
struct ModelBoxRpcClient {
    addr: String,
    rt: Option<tokio::runtime::Runtime>,
    client: Option<ModelStoreClient<tonic::transport::Channel>>,
}

#[pymethods]
impl ModelBoxRpcClient {
    #[new]
    pub fn new(addr: String) -> PyResult<Self> {
        let rt = tokio::runtime::Runtime::new().unwrap();
        let client = rt
            .block_on(modelbox::model_store_client::ModelStoreClient::connect(
                addr.clone(),
            ))
            .map_err(|e| PyErr::new::<pyo3::exceptions::PyException, _>(format!("{}", e)))?;
        Ok(Self {
            addr: addr,
            rt: Some(rt),
            client: Some(client),
        })
    }

    pub fn experiment(&mut self, id: String) -> PyResult<String> {
        let req = GetExperimentRequest { id: id };
        let fut_experiment = self.client.as_mut().unwrap().get_experiment(req);
        let res = self
            .rt
            .as_ref()
            .unwrap()
            .block_on(fut_experiment)
            .map_err(|e| PyErr::new::<pyo3::exceptions::PyException, _>(format!("{}", e)))?;

        match res.into_inner().experiment {
            Some(experiment) => Ok(experiment.name),
            None => Err(PyErr::new::<pyo3::exceptions::PyException, _>(
                "experiment not found",
            )),
        }
    }

    pub fn create_experiment(
        &mut self,
        experiment: api_structs::CreateExperiment,
    ) -> PyResult<api_structs::CreateExpeirmentResult> {
        let req = CreateExperimentRequest {
            name: experiment.name,
            namespace: experiment.namespace,
            owner: experiment.owner,
            framework: experiment.ml_framework,
            external_id: experiment.external_id,
            task: experiment.task,
        };
        let fut_experiment = self.client.as_mut().unwrap().create_experiment(req);
        let res = self
            .rt
            .as_ref()
            .unwrap()
            .block_on(fut_experiment)
            .map_err(|e| PyErr::new::<pyo3::exceptions::PyException, _>(format!("{}", e)))?;

        Ok(api_structs::CreateExpeirmentResult {
            id: res.into_inner().experiment_id,
        })
    }

    pub fn create_model(
        &mut self,
        model: api_structs::CreateModel,
    ) -> PyResult<api_structs::CreateModelResult> {
        let req = modelbox::CreateModelRequest {
            name: model.name,
            namespace: model.namespace,
            owner: model.owner,
            task: model.task,
            description: model.description,
        };
        let fut_model = self.client.as_mut().unwrap().create_model(req);
        let res = self
            .rt
            .as_ref()
            .unwrap()
            .block_on(fut_model)
            .map_err(|e| PyErr::new::<pyo3::exceptions::PyException, _>(format!("{}", e)))?;
        Ok(api_structs::CreateModelResult {
            id: res.into_inner().id,
        })
    }

    pub fn create_model_version(
        &mut self,
        model_version: api_structs::CreateModelVersion,
    ) -> PyResult<api_structs::CreateModelVersionResult> {
        let req = modelbox::CreateModelVersionRequest {
            model: model_version.model_id,
            name: model_version.name,
            version: model_version.version,
            description: model_version.description,
            namespace: model_version.namespace,
            framework: model_version.ml_framework.to_proto() as i32,
            unique_tags: model_version.tags,
        };
        let fut_model_version = self.client.as_mut().unwrap().create_model_version(req);
        let res = self
            .rt
            .as_ref()
            .unwrap()
            .block_on(fut_model_version)
            .map_err(|e| PyErr::new::<pyo3::exceptions::PyException, _>(format!("{}", e)))?;
        Ok(api_structs::CreateModelVersionResult {
            id: res.into_inner().model_version,
        })
    }

    pub fn log_event(
        &mut self,
        event: api_structs::LogEvent,
    ) -> PyResult<api_structs::LogEventResult> {
        let req = modelbox::LogEventRequest {
            parent_id: event.object_id,
            event: Some(modelbox::Event {
                name: event.name,
                source: Some(modelbox::EventSource { name: event.source }),
                wallclock_time: Some(prost_types::Timestamp {
                    seconds: event.timestamp,
                    nanos: 0,
                }),
                metadata: Some(modelbox::Metadata {
                    metadata: event.metadata,
                }),
            }),
        };
        let fut_event = self.client.as_mut().unwrap().log_event(req);
        let _res = self
            .rt
            .as_ref()
            .unwrap()
            .block_on(fut_event)
            .map_err(|e| PyErr::new::<pyo3::exceptions::PyException, _>(format!("{}", e)))?;
        Ok(api_structs::LogEventResult {})
    }
}

impl fmt::Display for ModelBoxRpcClient {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "ModelBoxRpcClient {{ addr: {} }}", self.addr)
    }
}

#[pymodule]
fn modelbox_rpc_client(py: Python, m: &PyModule) -> PyResult<()> {
    m.add_class::<ModelBoxRpcClient>()?;
    api_structs::register(py, m)?;
    Ok(())
}
