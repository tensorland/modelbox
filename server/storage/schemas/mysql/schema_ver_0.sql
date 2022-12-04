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
   metadata JSON
);

CREATE TABLE IF NOT EXISTS mutation_events (
   mutation_id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
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

CREATE TABLE IF NOT EXISTS actions (
   id VARCHAR(40) PRIMARY KEY,
   parent_id VARCHAR(40) NOT NULL,
   name VARCHAR(100) NOT NULL,
   arch VARCHAR(20) NOT NULL,
   params JSON,
   trigger_predicate TEXT,
   created_at BIGINT NOT NULL,
   updated_at BIGINT NOT NULL,
   finished_at BIGINT
);

CREATE TABLE IF NOT EXISTS action_instances (
   id VARCHAR(45) PRIMARY KEY,
   action_id VARCHAR(40) NOT NULL,
   attempt BIGINT NOT NULL,
   status INT NOT NULL,
   outcome INT NOT NULL,
   outcome_reason VARCHAR(20) NOT NULL,
   created_at BIGINT,
   updated_at BIGINT,
   finished_at BIGINT
);

CREATE TABLE IF NOT EXISTS cluster_members (
   id VARCHAR(40) PRIMARY KEY,
   info JSON,
   heartbeat_time BIGINT
);

CREATE TABLE IF NOT EXISTS action_evals (
   id VARCHAR(40) PRIMARY KEY,
   parent_id VARCHAR(40),
   parent_type INT,
   eval_type INT,
   created_at BIGINT NOT NULL,
   processed_at BIGINT
);
