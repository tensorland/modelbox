<!-- Output copied to clipboard! -->

<!-----

Yay, no errors, warnings, or alerts!

Conversion time: 0.392 seconds.


Using this Markdown file:

1. Paste this output into your source file.
2. See the notes and action items below regarding this conversion run.
3. Check the rendered output (headings, lists, code blocks, tables) for proper
   formatting and use a linkchecker before you publish this page.

Conversion notes:

* Docs to Markdown version 1.0β33
* Fri Sep 02 2022 12:19:42 GMT-0700 (PDT)
* Source doc: Install
* This is a partial selection. Check to make sure intra-doc links work.
* Tables are currently converted to HTML tables.
----->


---

sidebar_position: 2

---


# Install and Operation 

ModelBox can be installed and run in several models depending on the use case. Here we review a few of those use cases and discuss possible ways to run and operate the service.


## Service Components

The service consists of the following components -



1. ModelBox Server - The central control plane of ModelBox stores metadata related to experiments and models.
2. Metadata Storage Backend - Storage service for experiment and model metadata.
3. Blob Server - The server can be optionally run in the blob serving mode, where it only offers APIs to download and upload artifacts.
4. Metrics Backend - Storage service for time series data of experiments and models.


## Evaluating ModelBox Locally

The best way to evaluate ModelBox is to run it locally using ephemeral storage. This mode allows users to train new models and learn how to log, read and compare metadata using the SDK without thinking about deploying in a cluster.


#### Configuration

Generate the default config for ModelBox server. The CLI has a command to generate the default config. It generates configuration to run the server with ephemeral storage.


```
$ modelbox server init-config
```


The config generated should be the following in a file called modelbox_server.toml -


```
artifact_storage = "filesystem"
metadata_storage = "ephemeral"
metrics_storage = "inmemory"
listen_addr = ":8085"

[artifact_storage_filesystem]
base_dir = "/tmp/modelboxblobs"

[artifact_storage_s3]
region = "us-east-1"
bucket = "modelbox-artifacts"

[metadata_storage_integrated]
path = "/tmp/modelbox.dat"

[metadata_storage_postgres]
host = "172.17.0.3"
port = 5432
username = "postgres"
password = "foo"
dbname   = "modelbox"

[metadata_storage_mysql]
host     = "172.17.0.2"
port     = 3306
username = "root"
password = "foo"
dbname   = "modelbox"

[metrics_storage_timescaledb]
host = "172.17.0.4"
port = 5432
username = "postgres"
password = "foo"
dbname   = "modelbox_metrics" 
```



#### Start the Server 


```
$ modelbox server start -config-path ./path/to/modelbox_server.toml
```



## Production Deployment Scenarios

In production, it is expected that HA data storage services are used for metadata and metrics storage, and the ModelBox server is also run in a HA mode.


### Custom Installation in Datacenter


```
artifact_storage = "s3"
metadata_storage = "mysql"
metrics_storage = "timescaledb"
```


Once the configuration points to the appropriate datastorage services, the sections for the appropriate backends needs to be changed with the right credentials, schema name, and such.

Multiple instances of the servers should be run for high availability. The service metrics should be monitored, and the appropriate number of instances of services should be chosen to keep the API latency and resource usage of the server to reasonable limits.

Once the configuration is changed start the servers -


```
$ modelbox server start -config-path ./path/to/modelbox_server.toml
```



### Kubernetes


### AWS


### GCP


## Build ModelBox 

ModelBox can be built from source or be downloaded from GitHub.


#### Building from Source using GoReleaser

Install goreleaser from here. After it’s installed, the binary can be built.


```
goreleaser build --rm-dist --snapshot
```