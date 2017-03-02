#!/usr/bin/env bash
env GOOS=linux GOARCH=amd64 go build -v cmd/main.go
mv main dnsbench
