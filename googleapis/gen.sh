GOOGLE_APIS="github.com/ilius/gragen/googleapis"

protoc --proto_path=${GOPATH}/src/${GOOGLE_APIS} --go_out=${GOPATH}/src ${GOPATH}/src/${GOOGLE_APIS}/google/api/*.proto
cp ${GOPATH}/src/google.golang.org/genproto/googleapis/api/annotations/*.go ${GOPATH}/src/${GOOGLE_APIS}/google/api/
