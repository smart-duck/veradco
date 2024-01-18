go build -o docker/dummy main.go

cd docker;sudo docker build -t smartduck/dummy-grpc-plugin:0.1 -f ./Dockerfile .;cd ..

sudo ~/go/src/veradco/veradco/demo/local_registry/push_local_image_to_local_registry.sh smartduck/dummy-grpc-plugin:0.1

k apply -f deploy/deploy_plugin_dummy.yaml

~/go/bin/stern -n default dummyplugin &

k describe po $(k get po | grep dummyplugin2 | cut -d" " -f1)