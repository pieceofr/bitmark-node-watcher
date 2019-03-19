#!/bin/bash
## How to run bitmarkNodeWatcher as a docker container

docker run -it --name bitmarkNodeWatcher \
-e DOCKER_HOST="unix:///var/run/docker.sock" \
-e NODE_IMAGE="bitmark/bitmark-node-test" \
-e NODE_NAME="bitmarkNodeTest" \
-v /var/run/docker.sock:/var/run/docker.sock \
-v $nodeDir/data:/.config/bitmark-node/bitmarkd/bitmark/data \
-v $nodeDir/data-test:/.config/bitmark-node/bitmarkd/testing/data \
bitmark-node-watcher
