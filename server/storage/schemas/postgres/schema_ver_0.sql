CREATE TABLE IF NOT EXISTS schema_version (
   version INT,
   PRIMARY KEY (version)
);

INSERT INTO schema_version (version) VALUES(1) ON CONFLICT(version) DO UPDATE SET version=1;

CREATE TABLE IF NOT EXISTS experiments (
   id VARCHAR(40) PRIMARY KEY,
   external_id VARCHAR(40),
   name VARCHAR(50),
   owner VARCHAR(30),
   namespace VARCHAR(50),
   ml_framework INT,
   created_at BIGINT,
   updated_at BIGINT 
);

CREATE TABLE IF NOT EXISTS checkpoints (
   id VARCHAR(40) PRIMARY KEY,
   experiment VARCHAR(40),
   epoch int,
   metrics JSON,
   created_at BIGINT,
   updated_at BIGINT 
);

CREATE TABLE IF NOT EXISTS models (
   id VARCHAR(40) PRIMARY KEY,
   name VARCHAR(50),
   owner  VARCHAR(30),
   namespace VARCHAR(50),
   task VARCHAR(20),
   description TEXT,
   created_at BIGINT,
   updated_at BIGINT
);

CREATE TABLE IF NOT EXISTS model_versions (
   id VARCHAR(40) PRIMARY KEY,
   name VARCHAR(40),
   model_id VARCHAR(40),
   version VARCHAR(5),
   description TEXT,
   ml_framework INT,
   unique_tags JSON,
   created_at BIGINT,
   updated_at BIGINT
);

CREATE TABLE IF NOT EXISTS blobs (
   id VARCHAR(40) PRIMARY KEY,
   parent_id VARCHAR(40),
   metadata JSON
);

CREATE TABLE IF NOT EXISTS metadata (
   id VARCHAR(40) PRIMARY KEY,
   parent_id VARCHAR(40),
   metadata JSON,
   created_at BIGINT,
   updated_at BIGINT
);

CREATE TABLE IF NOT EXISTS mutation_events (
   mutation_id SERIAL PRIMARY KEY,
   mutation_time BIGINT,
   action VARCHAR(20),
   object_id VARCHAR(40),
   object_type VARCHAR(20),
   parent_id VARCHAR(40),
   namespace VARCHAR(40),
   payload JSON
);

CREATE TABLE IF NOT EXISTS events (
   id VARCHAR(40) PRIMARY KEY,
   parent_id VARCHAR(40) NOT NULL,
   name TEXT,
   source_name TEXT,
   wallclock BIGINT,
   metadata JSON
);