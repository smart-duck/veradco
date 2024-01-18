go build -o docker/valid main.go

cd docker;sudo docker build -t smartduck/valid-grpc-plugin:0.1 -f ./Dockerfile .;cd ..

sudo ~/go/src/veradco/veradco/demo/local_registry/push_local_image_to_local_registry.sh smartduck/valid-grpc-plugin:0.1

k apply -f deploy/deploy_plugin_valid.yaml

~/go/bin/stern -n default validplugin &
