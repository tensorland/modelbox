# Getting Started

This guide aims to get an instance of ModelBox started and have you train a model, and then look at the metadata and metrics collected by ModelBox to analyze the experiment, models and other artifacts.

There are three ways to get started with ModelBox. The docker-compose method is preferred if you have Docker installed on your machine. If you don't have Docker, GitPod would be the second best alternative. 
Lastly, you could down the ModelBox binary and run it locally as well, either with ephemeral storage or use the various storage dependencies.

## Docker Compose

This is the quickest way to get started if you have docker and docker-compose installed. 

```
docker compose --profile local up
```

This starts the ModelBox server with all the dependencies and a container with a Jupyter Notebook that demonstrates how to integrate a Pytorch trainer with ModelBox.

The ModelBox server hosts the API at the address - `172.21.0.2:8085`

The Jupyter notebook with the tutorials is available at the address - `https://localhost:8888`

Train a Pytorch Model by following the [notebook](https://github.com/tensorland/modelbox/blob/main/tutorials/Pytorch_Lightning_Integration_Tutorial.ipynb)


## Gitpod

Gitpod provides an ephemeral development environment which is ideal for -
1. Getting started with ModelBox and evaluate the service if you don't have Docker available locally.
2. Contributing code to ModelBox without setting up the development environment locally.

Click the following button to have a GitPod workspace. When the workspace starts, it will automatically bring up a docker environment in a terminal and the PyTorch notebook can be run inside the workspace.

[![Open in Gitpod](https://gitpod.io/button/open-in-gitpod.svg)](https://gitpod.io/#https://github.com/tensorland/modelbox)

## Local Server with Ephemeral Metadata Storage

1. Install the dependencies

2. Build ModelBox

3. Generate the Server and client configs.

4. Train a Model

## Local Server with Local Datastores

1. Follow steps 1-3 from the above section which demonstrates how to run modelbox locally.

2. Install Postgres/MySQL Server.

3. Decide which metrics backend to use.

4. Decide which blob storage backend to use.

5. Start the server and train a model.