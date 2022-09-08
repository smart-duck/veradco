#!/bin/bash

for item in $(sudo docker images | grep "localhost:5001/smartduck/veradco" | grep -v " 0.1 " |  sed -E 's/^[^ ]+ +[^ ]+ +([^ ]+).+$/\1/g')
do
  sudo docker image rm $item
done

# sudo docker container prune

# sudo docker image prune -a

# sudo docker volume prune