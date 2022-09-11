# Dockerfile
FROM gcr.io/distroless/static-debian11
WORKDIR /app
COPY modelbox /app/modelbox
COPY cmd/modelbox/assets/modelbox_server.toml /app/modelbox_server.toml
EXPOSE 8085
