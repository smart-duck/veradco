#!/bin/sh

source $(dirname $(readlink -f $0))/start_any_script.source

cd /go/src/veradco
rm go.mod go.sum

set -e

go mod init github.com/smart-duck/veradco
go mod tidy

echo "veradco go.mod:"
cat go.mod
echo "veradco go.sum:"
cat go.sum || true

mkdir -p /app

go build -o /app/veradcod cmd/serverd/main.go

source $(dirname $(readlink -f $0))/end_any_script.source