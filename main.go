package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) > 1 {
		argPath := os.Args[1]
		var basePath string
		if strings.HasSuffix(argPath, ".proto") {
			basePath = argPath[:len(argPath)-len(".proto")]
		} else if strings.HasSuffix(argPath, ".pb.go") {
			basePath = argPath[:len(argPath)-len(".pb.go")]
		} else {
			panic(fmt.Errorf("uknown file extention for file %%v, must be .proto or .pb.go", argPath))
		}
		service, err := parsePbGoFile(basePath)
		if err != nil {
			panic(err)
		}
		err = parseProtoFile(service, basePath)
		if err != nil {
			panic(err)
		}
		code, err := generateServiceCode(service)
		if err != nil {
			panic(err)
		}
		outFilePath := filepath.Join(service.DirPath, service.Name+"_adaptor.go")
		err = ioutil.WriteFile(outFilePath, []byte(code), 0644)
		if err != nil {
			panic(err)
		}
	}
}
