# ModelBox Python API

This package contains the client library for the ModelBox API for managing Deep Learning models, checkpoints from experiments, and other model operations related services.

## Concepts and Understanding the ModelBox API

### Namespace
A Namespace is a mechanism to organize related models or models published by a team. They are also use for access control and such to the metadata of uploaded models, invoking benchmarks or other model transformation work. Namespaces are automatically created when a new model or experieemnt specifies the namespace it wants to be associated with.

### Model
A model is an object to track common metadata, and to apply policies on models created by experiments to solve a machine learning task. For ex. datasets to evaluate all trained models of a task can be tracked using this object. Users can also add rules around retention policies of trained versions, setup policies for labelling a trained model if it has better metrics on a dataset, and meets all other criterion.

### Model Version
A model version is a trained model, it includes the model binary, related files that a user wants to track - dataset file handles, any other metadata, model metrics, etc. Model versions are always related to a Model and all the policies created for a Model are applied to Model Versions.

### Experiment and Checkpoints
Experiments are used to injest model checkpoints created during a training run. ModelBox is not an experiment metadata tracker so there is no support for rich experiment management which are available on experiment trackers such as Weights and Biases, the experiment abstraction here exists so that we can track and injest model checkpoints which eventually become model versions if they have good metrics and does well in benchmarks.

## Example

```
from modelbox import ModelBoxClient, MLFramework

client = ModelBoxClient(SERVER_ADDR)

model = self._client.create_model(
            "yolo",
            "owner@email.com",
            "ai/vision/",
            "object_detection",
            "yolo_des",
            {"meta": "foo"},
        )
model_version = self._client.create_model_version(
            model.id, 
            "yolo4_v1",
            "v1",
            "A Yolo v4 trained with custom dataset", 
            ["s3://path/to/bucket/model.pt],
            {"model_hyperparam_1": "value"},
            MLFramework.PYTORCH,
        )

client.close()
```


## Local Development and Installation
The modelbox client library can be installed locally in the following way -
```
cd <project-root>/client-py/
pip install .
```
This installs the version of the client checked out with the repo.

Build the client and create distribution packages
```
cd <project-root>/client-py/
python -m build .
```

Run Tests 
```
cd <project-root>/client-py/
python tests/test_modelbox_api.py
```