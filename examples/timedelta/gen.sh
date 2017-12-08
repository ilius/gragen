#!/bin/bash

GRAGEN="github.com/ilius/gragen"

rm timedelta.pb.go timedelta_adaptor.go 2>/dev/null

protoc -I. \
	"-I${GOPATH}/src" \
	--go_out=plugins=grpc:. \
	timedelta.proto \
	|| exit $?

go build "$GRAGEN" || exit $?


./gragen timedelta.pb.go || exit $?

go build -o server.bin ./server || exit $?
chmod a+x server.bin