# Dockerfile
#FROM gcr.io/distroless/static-debian11
FROM ubuntu
WORKDIR /app
COPY modelbox /app/modelbox
COPY cmd/modelbox/assets/modelbox_server.yaml /app/modelbox_server.yaml
EXPOSE 8085
