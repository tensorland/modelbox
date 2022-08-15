SELECT 'CREATE DATABASE modelbox_metrics'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'modelbox_metrics')\gexec