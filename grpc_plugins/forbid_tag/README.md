go build -o docker/forbidtag main.go

cd docker;sudo docker build -t smartduck/forbidtag-grpc-plugin:0.1 -f ./Dockerfile .;cd ..

sudo ~/go/src/veradco/veradco/demo/local_registry/push_local_image_to_local_registry.sh smartduck/forbidtag-grpc-plugin:0.1

k apply -f deploy/deploy_plugin_forbidtag.yaml

~/go/bin/stern -n default forbidtagplugin &
