# Developing ModelBox


## Push latest snapshot containers of modelbox to DockerHub

1. Build the server and the docker containers.
```
goreleaser release --snapshot --rm-dist
```

2. Push the images to dockerhub. *Replace the versions*
```
docker push modelboxdotio/modelbox:0.0.1-next-arm64v8
docker push modelboxdotio/modelbox:0.0.1-next-amd64
```

3. Create a manifest to link the containers of various arch under one tag
```
docker manifest create modelboxdotio/modelbox:latest modelboxdotio/modelbox:0.0.1-next-arm64v8 modelboxdotio/modelbox:0.0.1-next-amd64
```