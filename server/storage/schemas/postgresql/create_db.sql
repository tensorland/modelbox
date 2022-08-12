SELECT 'CREATE DATABASE modelbox'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'modelbox')\gexec