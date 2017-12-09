package main

import (
	"encoding/json"
)

type Service struct {
	Name       string // starts with lowercase
	ServerName string // name of grpc server interface
	ClientName string // name of grpc client interface
	Methods    []*Method
	DirPath    string

	// the following 2 maps: map[namespace] -> {alias, import_path}
	// alias can be empty, and is usually empty
	// alias can also be "." and "_", in these case, key(namespace) is the same as import_path
	Imports        map[string][2]string
	AdaptorImports map[string][2]string
}

type Param struct {
	JsonKey string
	Name    string
	Type    string
}

type Method struct {
	Name          string
	RequestName   string
	RequestParams []Param
	ResponseName  string
}

func (m Method) String() string {
	jsonBytes, _ := json.MarshalIndent(m, "", "    ")
	return string(jsonBytes)
}
