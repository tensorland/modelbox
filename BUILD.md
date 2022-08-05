# Building ModelBox 

This guide is intended for users installing Modelbox and developers who wish to contribute to the project to build modelbox from source.

## Build using the go toolchain

```
go build -o build/modelbox ./cmd/modelbox
````

## Build using goreleaser for development
```
goreleaser build --single-target --snapshot
```

## Build using goreleaser for all supported platforms

This is the preferred way to build ModelBox as it creates binaries for all the versions we support -

Install goreleaser first - https://goreleaser.com/install/

Once gorealeaser is installed, run the following command -
```
goreleaser build --snapshot --rm-dist
```

## Building a Docker Container to serve ModelBox server locally
This will build the modelbox server binary from source and create a Docker container to run the server. This is only possible if goreleaser is invoked in the realease mode.
```
goreleaser release --snapshot --rm-dist
```