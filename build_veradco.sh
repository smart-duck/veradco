#!/bin/sh

source $(dirname $(readlink -f $0))/start_any_script.source

cd /go/src/veradco
rm go.mod go.sum

set -e

go mod init github.com/smart-duck/veradco
go mod tidy

mkdir -p /release

go build -o /release/veradcod cmd/serverd/main.go

source $(dirname $(readlink -f $0))/end_any_script.source