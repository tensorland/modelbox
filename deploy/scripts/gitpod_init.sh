#!/bin/bash

# Install goreleaser
go install github.com/goreleaser/goreleaser@latest

# Build modelbox
goreleaser release --snapshot --rm-dist