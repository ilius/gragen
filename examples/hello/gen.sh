#!/bin/bash

GRAGEN="github.com/ilius/gragen"

rm hello.pb.go hello_adaptor.go 2>/dev/null

protoc -I. \
	"-I${GOPATH}/src" \
	--go_out=plugins=grpc:. \
	hello.proto \
	|| exit $?

go build "$GRAGEN" || exit $?


./gragen hello.pb.go || exit $?

go build -o server.bin ./server || exit $?
chmod a+x server.bin