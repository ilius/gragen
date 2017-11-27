package main

import (
	"encoding/json"
)

type Service struct {
	Name    string // starts with lowercase
	Methods []*Method
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
