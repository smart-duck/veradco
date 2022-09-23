# Overview

This plugin allows to implicitly pull docker images in harbor proxy caches. To do it, it implements the docker registry specification to mimic docker pull.

## Check that the plugin builds

```
go mod init github.com/smart-duck/veradco/harbor_proxy_cache_populator
go mod edit -replace github.com/smart-duck/veradco=../../veradco
go mod tidy
go build -buildmode=plugin -o /dev/null plug.go
```

# TODO

- Debug mode with long sleep
- Manage dryRun in plugin
- Queue with channels