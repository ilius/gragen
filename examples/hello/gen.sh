#!/bin/bash

GRAGEN="github.com/ilius/gragen"

rm hello.pb.go

protoc -I. \
	"-I${GOPATH}/src" \
	--go_out=plugins=grpc:. \
	hello.proto \
	|| exit $?

go build "$GRAGEN" || exit $?

./gragen hello.pb.go

# ./gragen hello.pb.go > hello_adaptor.go || exit $?