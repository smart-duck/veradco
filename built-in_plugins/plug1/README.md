# Init plugin golang environment

```
cd plug1
go mod init github.com/smart-duck/veradco/plug1
go mod edit -replace github.com/smart-duck/veradco=../../veradco
go mod tidy
```
# Setup golang environment

```
go mod init github.com/smart-duck/veradco/plug1
go mod tidy
go mod edit -replace github.com/smart-duck/veradco=../../veradco
go mod tidy
```


# Build plugin

```
cd plug1
go build -buildmode=plugin -o plug.so plug.go
```

# Package plugins

```
sudo docker build -t smartduck/veradco_plugins:0.1 .
sudo ../veradco/demo/local_registry/push_local_image_to_local_registry.sh smartduck/veradco_plugins:0.1
```
