CREATE TABLE IF NOT EXISTS schema_version (
   version INT,
   PRIMARY KEY (version)
);

INSERT INTO schema_version (version) VALUES(1) ON DUPLICATE KEY UPDATE version=1;

CREATE TABLE IF NOT EXISTS experiments (
   id VARCHAR(40) PRIMARY KEY,
   external_id VARCHAR(40),
   name VARCHAR(50),
   owner VARCHAR(30),
   namespace VARCHAR(50),
   ml_framework INT,
   metadata JSON,
   created_at BIGINT,
   updated_at BIGINT 
);

CREATE TABLE IF NOT EXISTS checkpoints (
   id VARCHAR(40) PRIMARY KEY,
   experiment VARCHAR(40),
   epoch int,
   metrics JSON,
   metadata JSON,
   created_at BIGINT,
   updated_at BIGINT 
);

CREATE TABLE IF NOT EXISTS models (
   id VARCHAR(40) PRIMARY KEY,
   name VARCHAR(50),
   owner  VARCHAR(30),
   namespace VARCHAR(50),
   task VARCHAR(20),
   metadata JSON,
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
   metadata JSON,
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
   metadata JSON
);