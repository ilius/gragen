#!/bin/bash

GRAGEN="github.com/ilius/gragen"

protoc -I. \
	"-I${GOPATH}/src" \
	--go_out=. \
	hello.proto \
	|| exit $?


cp hello.pb.go "$GOPATH/src/$GRAGEN/proto_registry/" || exit $?

go build "$GRAGEN" || exit $?

./gragen hello.proto > hello_adaptor.go || exit $?
go fmt hello_adaptor.go || exit $?