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

func decodeMethodFieldIdent(fieldType *ast.Ident) (*MethodField, error) {
	if fieldType.Name == "error" {
		return nil, nil
	}
	if fieldType.Obj == nil {
		return nil, fmt.Errorf("fieldType.Obj is nil, name=%#v, fieldType = %T %v", fieldType.Name, fieldType, fieldType)
	}
	switch decl := fieldType.Obj.Decl.(type) {
	case *ast.TypeSpec:
		switch declType := decl.Type.(type) {
		case *ast.InterfaceType:
			methods := map[string]*InterfaceMethod{}
			for _, m := range declType.Methods.List {
				if len(m.Names) == 0 {
					continue
				}
				name := m.Names[0].Name
				mymethod := &InterfaceMethod{Name: name}
				for _, arg := range m.Type.(*ast.FuncType).Params.List {
					argType, err := formatTypeExpr(arg.Type)
					if err != nil {
						return nil, err
					}
					mymethod.InputTypes = append(mymethod.InputTypes, argType)
				}
				for _, result := range m.Type.(*ast.FuncType).Results.List {
					resultType, err := formatTypeExpr(result.Type)
					if err != nil {
						return nil, err
					}
					mymethod.OutputTypes = append(mymethod.OutputTypes, resultType)
				}
				methods[name] = mymethod
			}
			return &MethodField{
				Name:             fieldType.Name,
				Kind:             MF_interface,
				InterfaceMethods: methods,
			}, nil
		default:
			return nil, fmt.Errorf("unexpected type for declType: %T : %v", declType, declType)
		}
	default:
		return nil, fmt.Errorf("unexpected type for decl: %T : %v", decl, decl)
	}
}

// returns (name, isInterface, interfaceMethods, err)
func decodeMethodField(field *ast.Field) (*MethodField, error) {
	switch fieldType := field.Type.(type) {
	case *ast.StarExpr:
		return &MethodField{
			Name: fieldType.X.(*ast.Ident).Name,
			Kind: MF_structPointer,
		}, nil
	case *ast.Ident:
		return decodeMethodFieldIdent(fieldType)
	case *ast.SelectorExpr:
		// fieldType.X is the package
		// fieldType.Sel is the type inside that package
		pkgName, err := formatTypeExpr(fieldType.X)
		if err != nil {
			return nil, err
		}
		if fieldType.Sel.Obj == nil { // for context.Context
			selName, err := formatTypeExpr(fieldType.Sel)
			if err != nil {
				return nil, err
			}
			if pkgName == "context" && selName == "Context" {
				return &MethodField{
					Name: pkgName + "." + selName,
					Kind: MF_context,
				}, nil
			}
			return &MethodField{
				Name: pkgName + "." + selName,
				Kind: MF_unknown,
			}, nil
		}
		fieldDecoded, err := decodeMethodFieldIdent(fieldType.Sel)
		if err != nil {
			return nil, err
		}
		fieldDecoded.Name = pkgName + fieldDecoded.Name
		return fieldDecoded, nil
	default:
		return nil, fmt.Errorf("unexpected type for fieldType: %T : %v", fieldType, fieldType)
	}
	return nil, fmt.Errorf("unexpected type for field: %T : %v", field, field)
}

func getServerMethods(fileScope *ast.Scope, serverObj *ast.Object) ([]*Method, error) {
	methods := []*Method{}
	typeSpec := serverObj.Decl.(*ast.TypeSpec)
	interfaceType := typeSpec.Type.(*ast.InterfaceType)
	for _, method := range interfaceType.Methods.List {
		methodName := method.Names[0].Name
		methodType := method.Type.(*ast.FuncType)
		args := methodType.Params.List
		var requestName string
		for _, arg := range args {
			argDecoded, err := decodeMethodField(arg)
			if err != nil {
				return nil, err
			}
			if argDecoded == nil {
				continue
			}
			if argDecoded.Kind == MF_structPointer {
				requestName = argDecoded.Name
			}
			// {
			// 	b, _ := json.MarshalIndent(argDecoded, ">", "    ")
			// 	fmt.Println(string(b))
			// }
		}
		if requestName == "" {
			return nil, fmt.Errorf("could not determine request type")
		}

		var responseName string

		for _, result := range methodType.Results.List {
			resultDecoded, err := decodeMethodField(result)
			if err != nil {
				return nil, err
			}
			if resultDecoded == nil {
				continue
			}
			if resultDecoded.Kind == MF_structPointer {
				responseName = resultDecoded.Name
			}
			// {
			// 	b, _ := json.MarshalIndent(resultDecoded, "<", "    ")
			// 	fmt.Println(string(b))
			// }
		}

		if responseName == "" {
			return nil, fmt.Errorf("could not determine response type")
		}

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
