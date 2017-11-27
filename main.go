package main

import (
	"fmt"
	"os"
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
		fmt.Println(code)
	}
}
