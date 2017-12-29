package main

import (
	"fmt"
	"os"
	"path/filepath"

	protoparser "github.com/emicklei/proto"
)

func parseProtoFile(service *Service, basePath string) error {
	protoPath := basePath + ".proto"
	file, err := os.Open(protoPath)
	if err != nil {
		return err
	}
	parser := protoparser.NewParser(file)
	protoFileParts := filepath.SplitList(protoPath)
	protoFilename := protoFileParts[len(protoFileParts)-1]
	parser.Filename(protoFilename)
	p, err := parser.Parse()
	if err != nil {
		return err
	}
	for _, element := range p.Elements {
		// fmt.Printf("type(element)=%T, element=%v\n", element, element)
		serviceElement, ok := element.(*protoparser.Service)
		if !ok {
			continue
		}
		for _, subElement := range serviceElement.Elements {
			// fmt.Printf("type(subElement)=%T, subElement=%v\n", subElement, subElement)
			methodElement, ok := subElement.(*protoparser.RPC)
			if !ok {
				continue
			}
			methodName := methodElement.Name
			method, ok := service.MethodByName[methodName]
			if !ok {
				return fmt.Errorf("Internal error: method %v not found in .pb.go file", methodName)
			}
			methodOptions := map[string]map[string]string{}
			for _, opt := range methodElement.Options {
				// opt.Name == "(google.api.http)"
				optMap := map[string]string{}
				for _, c := range opt.AggregatedConstants {
					optMap[c.Name] = c.Literal.Source
				}
				methodOptions[opt.Name] = optMap
				// fmt.Printf("Method %v, Option %v, Values: %v\n", methodName, opt.Name, optMap)
			}
			method.Options = methodOptions
		}
	}

	return nil
}
