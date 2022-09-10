---
sidebar_position: 1
---

# Configuring ModelBox Server
ModelBox server reads a toml configuration file to initialize the various service dependencies. A sample configuration is here - 

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

## Generate Server Configuration
```
modelbox server init-config
```

## Server Parameters
* `artifact_storage` -  
    - `s3` -  Artifacts are stored in AWS S3. `artifact_storage_s3` section is read for S3 specific configuration.
    - `filesystem` -  Artifacts are stored in filesystem. `artifact_storage_filesystem` section in read for file-system specific configuration.

* `metadata_storage` - Backend to use for storing experiment and model metadata.
    - `mysql` - `metadata_storage_mysql` is read to configure mysql.
    - `postgres` - `metadata_storage_postgres` is read to configure postgres.
    - `ephemeral` -  `metadata_storage_ephemeral1 is read to configure filesystem based storage.

* `metrics_storage` - Backend to use for metrics storage. Possible options -
    - `inmemory` - Metrics are stored in memory. No further configuration is required.
    - `timescaledb` - Metrics are stored in timescaledb. `metrics_storage_timescaledb` section is read to configure timescaledb.

* `listen_addr` -  Network interfaces and ports on which the modelbox server binds to.

* `artifact_storage_filesystem` - Configuration to store artifacts in the filesystem.
    - `base_dir` - Base directory in the filesystem where artifacts are stored.

* `artifact_storage_s3` - Configuration to store artifacts in S3
    - `region` - Region of the bucket
    - `bucket` - Bucket to store the artifacts.

* `metadata_storage_integrated` - Configuration to store server metadata in filesystem
    - `path` - Path of the file where boltdb stores data.

* `metadata_storage_postgres` - Configuration of Postgres database for storing server metadata
    - `host` - host or dns address of the Postgres service
    - `port` - Port on which the database is listening 
    - `username` - username of the database.
    - `password` - password of the database.
    - `dbname` - name of the database.

* `metadata_storage_mysql` - Configuration of MySQL database for storing server metadata
    - `host` - host or dns address of the MySQL service
    - `port` - port on which the database is listening 
    - `username` - service username to access the database.
    - `password` - password to access the database.
    - `dbname` - name of the database.

* `metrics_storage_timescaledb` - Configuration for TimescaleDB.
    - `host` - host or dns address of the Postgres service
    - `port` - port on which the database is listening 
    - `username` - service username to access the database.
    - `password` - password to access the database.
    - `dbname` - name of the database.

# Configuring ModelBox Client
ModelBox Client binary requires a config to interact with the API endpoint. An example client configuration -

```
server_addr = ":8085"
```

## Generate Client Configuration
```
modelbox client init-config
```

## Client Parameters
* `server_addr` : Address of the server API endpoint