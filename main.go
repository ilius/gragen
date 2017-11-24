package main

import (
	"os"
)

func main() {
	if len(os.Args) > 1 {
		protoPath := os.Args[1]
		err := parseProtoFile(protoPath)
		if err != nil {
			panic(err)
		}
	}
}
