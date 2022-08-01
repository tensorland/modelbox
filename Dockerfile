# Dockerfile
FROM gcr.io/distroless/static-debian11
WORKDIR /app
COPY modelbox /app/modelbox
EXPOSE 8085
ENTRYPOINT ["/app/modelbox"]

