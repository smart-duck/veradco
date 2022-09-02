#!/bin/sh

cd /go/src/veradco
rm go.mod go.sum

go mod init github.com/smart-duck/veradco
go mod tidy

mkdir -p /release

go build -o /release/veradcod cmd/serverd/main.go 
