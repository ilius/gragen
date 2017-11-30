package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) > 1 {
		pbGoPath := os.Args[1]
		service, err := parsePbGoFile(pbGoPath)
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
