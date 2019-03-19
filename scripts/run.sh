#!/bin/bash
## How to run bitmarkNodeWatcher as a docker container
## Setup your nase mount directory
nodeDir=$HOME/bitmark-node-data-test
docker run -it --name bitmarkNodeWatcher \
-e DOCKER_HOST="unix:///var/run/docker.sock" \
-e NODE_IMAGE="bitmark/bitmark-node-test" \
-e NODE_NAME="bitmarkNodeTest" \
-e USER_HOME_BASE_DIR=$nodeDir \
-v /var/run/docker.sock:/var/run/docker.sock \
-v $nodeDir:/log \ 
-v $nodeDir/data:/.config/bitmark-node/bitmarkd/bitmark/data \
-v $nodeDir/data-test:/.config/bitmark-node/bitmarkd/testing/data \
bitmark-node-watcher
