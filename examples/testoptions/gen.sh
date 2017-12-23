#!/bin/bash

GRAGEN="github.com/ilius/gragen"
GOOGLE_APIS="github.com/ilius/gragen/googleapis"

rm testoptions.pb.go testoptions_adaptor.go 2>/dev/null

protoc -I. \
	"-I${GOPATH}/src" \
	"-I${GOPATH}/src/${GOOGLE_APIS}" \
	--go_out=Mgoogle/api/annotations.proto=${GOOGLE_APIS}/google/api,plugins=grpc:. \
	testoptions.proto \
	|| exit $?

go build "$GRAGEN" || exit $?


./gragen testoptions.pb.go || exit $?

go build -o server.bin ./server || exit $?
chmod a+x server.bin