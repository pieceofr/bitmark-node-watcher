FROM bitmark/go-env:go12 as go-env

ENV GO111MODULE on

RUN cd /go/src && \
git clone https://github.com/pieceofr/bitmark-node-watcher && \
cd /go/src/bitmark-node-watcher && go mod download && \
go install && cd /go/bin

ADD dockerAssets/startwatcher.sh /
RUN cd / && chmod +x startwatcher.sh
CMD /startwatcher.sh

