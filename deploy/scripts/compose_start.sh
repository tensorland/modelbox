#!/bin/bash

# Setup schema
until /app/modelbox server create-schema --schema-dir /app/schemas/ --config-path=/app/config/modelbox_server_compose.toml
do
  echo "Trying to apply schema again in 5 seconds...."
  sleep 5s
done

# Srart server
/app/modelbox server start --config-path=/app/config/modelbox_server_compose.toml