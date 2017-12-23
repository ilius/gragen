#!/bin/bash

GRAGEN="github.com/ilius/gragen"

rm customtype.pb.go customtype_adaptor.go 2>/dev/null

protoc -I. \
	"-I${GOPATH}/src" \
	--go_out=plugins=grpc:. \
	types/types.proto \
	|| exit $?

protoc -I. \
	"-I${GOPATH}/src" \
	--go_out=plugins=grpc:. \
	customtype.proto \
	|| exit $?

go build "$GRAGEN" || exit $?


./gragen customtype.pb.go || exit $?

go build -o server.bin ./server || exit $?
chmod a+x server.bin