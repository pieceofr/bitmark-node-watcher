#/bin/bash
# script to run after docker is started
/go/bin/bitmark-node-watcher --host=$DOCKER_HOST --image=$NODE_IMAGE --name=$NODE_NAME