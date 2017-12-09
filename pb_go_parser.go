package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// returns (clientName, serverName, serverObj)
func findClientServerInterfaces(f *ast.File) (string, string, *ast.Object) {
	registerFuncName := ""
	for objName, _ := range f.Scope.Objects {
		if strings.HasPrefix(objName, "Register") && strings.HasSuffix(objName, "Server") {
			registerFuncName = objName
			break
		}
	}
	if registerFuncName == "" {
		return "", "", nil
	}
	serverName := registerFuncName[len("Register"):]
	serverObj := f.Scope.Objects[serverName]
	baseName := serverName[:len(serverName)-len("Server")]
	clientName := baseName + "Client"
	_, ok := f.Scope.Objects[clientName]
	if !ok {
		panic(fmt.Sprintf("did not find client interface: %v", clientName))
	}

	return clientName, serverName, serverObj
}

func formatTypeExpr(expr interface{}) (string, error) {
	switch exprTyped := expr.(type) {
	case *ast.Ident:
		return exprTyped.Name, nil
	case *ast.StarExpr:
		nonPtr, err := formatTypeExpr(exprTyped.X)
		if err != nil {
			return "", err
		}
		return "*" + nonPtr, nil
	case *ast.ArrayType:
		item, err := formatTypeExpr(exprTyped.Elt)
		if err != nil {
			return "", err
		}
		return "[]" + item, nil
	case *ast.SelectorExpr:
		pkgName, err := formatTypeExpr(exprTyped.X)
		if err != nil {
			return "", err
		}
		localName, err := formatTypeExpr(exprTyped.Sel)
		if err != nil {
			return "", err
		}
		return pkgName + "." + localName, nil
	}
	return "", fmt.Errorf("could not detect type name from %v with type %T", expr, expr)
}

func getJsonKeyFromTag(tag string) string {
	parts := strings.Split(tag, `json:"`)
	if len(parts) != 2 {
		return ""
	}
	parts = strings.Split(parts[1], `"`)
	if len(parts) < 1 {
		return ""
	}
	jsonTag := parts[0]
	if jsonTag == "" {
		return ""
	}
	parts = strings.Split(jsonTag, ",")
	if len(parts) < 1 {
		return ""
	}
	jsonKey := parts[0]
	if jsonKey == "-" {
		return ""
	}
	return jsonKey
}

func getServerMethods(fileScope *ast.Scope, serverObj *ast.Object) ([]*Method, error) {
	methods := []*Method{}
	typeSpec := serverObj.Decl.(*ast.TypeSpec)
	interfaceType := typeSpec.Type.(*ast.InterfaceType)
	for _, method := range interfaceType.Methods.List {
		methodName := method.Names[0].Name
		methodType := method.Type.(*ast.FuncType)
		args := methodType.Params.List
		if len(args) != 2 {
			return nil, fmt.Errorf("unexpected number of method arguments: %v", args)
		}
		requestName := args[1].Type.(*ast.StarExpr).X.(*ast.Ident).Name
		results := methodType.Results.List
		if len(results) != 2 {
			return nil, fmt.Errorf("unexpected number of method return values: %v", results)
		}
		responseName := results[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name

		requestObj := fileScope.Lookup(requestName)
		if requestObj == nil {
			return nil, fmt.Errorf("request struct %v was not found", requestName)
		}
		requestParams := []Param{}
		for _, field := range requestObj.Decl.(*ast.TypeSpec).Type.(*ast.StructType).Fields.List {
			Type, err := formatTypeExpr(field.Type)
			if err != nil {
				return nil, err
			}
			jsonKey := getJsonKeyFromTag(field.Tag.Value)
			if jsonKey == "" {
				return nil, fmt.Errorf("invalid or unexpected struct tag %#v", field.Tag.Value)
			}
			requestParams = append(requestParams, Param{
				Name:    field.Names[0].Name,
				JsonKey: jsonKey,
				Type:    Type,
			})
		}

		methods = append(methods, &Method{
			Name:          methodName,
			RequestName:   requestName,
			RequestParams: requestParams,
			ResponseName:  responseName,
		})
	}
	return methods, nil
}

func parsePbGoFile(pbGoPath string) (*Service, error) {
	dirPath, filename := filepath.Split(pbGoPath)
	if !strings.HasSuffix(filename, ".pb.go") {
		return nil, fmt.Errorf("filename must end with .pb.go")
	}
	pkgName := filename[:len(filename)-len(".pb.go")]

	srcBytes, err := ioutil.ReadFile(pbGoPath)
	if err != nil {
		return nil, err
	}
	src := string(srcBytes)
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, pbGoPath, src, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	imports := map[string][2]string{}
	for _, imp := range f.Imports {
		alias := imp.Name.Name
		path := strings.Trim(imp.Path.Value, `"`)
		name := alias
		switch alias {
		case "":
			name = path // FIXME
		case ".", "_":
			// there is no usable namespace, so we just use the `path` to make it unique
			name = path
		}
		imports[name] = [2]string{alias, path}
	}

	clientName, serverName, serverObj := findClientServerInterfaces(f)
	if serverObj == nil {
		return nil, fmt.Errorf("could not find Server interface")
	}
	methods, err := getServerMethods(f.Scope, serverObj)
	if err != nil {
		return nil, err
	}
	service := &Service{
		Name:       pkgName,
		ClientName: clientName,
		ServerName: serverName,
		Methods:    methods,
		DirPath:    dirPath,
		Imports:    imports,
	}

	return service, nil
}
