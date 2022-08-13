CREATE TABLE IF NOT EXISTS metrics(
   time TIMESTAMPTZ  NOT NULL,
   parent_id TEXT NOT NULL,
   name TEXT NOT NULL,
   double_value DOUBLE PRECISION NULL,
   tensor TEXT NULL,
   step BIGINT NULL,
   wallclock BIGINT NULL
);

SELECT create_hypertable('metrics', 'time', if_not_exists => true);