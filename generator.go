package main

import (
	"fmt"
	"go/format"
	"log"
	"strings"
)

const (
	t_string      = "string"
	t_int         = "int"
	t_int32       = "int32"
	t_int64       = "int64"
	t_float64     = "float64"
	t_float32     = "float32"
	t_bool        = "bool"
	t_stringSlice = "[]string"
)

func generateServiceMethodsCode(service *Service) (string, error) {
	code := ""
	for _, method := range service.Methods {
		methodCode, err := generateMethodCode(method)
		if err != nil {
			return "", err
		}
		code += "\n" + methodCode + "\n\n"
	}
	return code, nil
}

func generateServiceCode(service *Service) (string, error) {
	methodsCode, err := generateServiceMethodsCode(service)
	if err != nil {
		return "", err
	}

	imports := []string{
		"log",
		"net/http",
		"github.com/ilius/ripo",
	}

	if strings.Contains(methodsCode, "reflect.") {
		imports = append(imports, "reflect")
	}

	code := "package " + service.Name + "\n"
	code += "import (\n"
	for _, imp := range imports {
		code += "\t" + `"` + imp + `"` + "\n"
	}
	code += ")"

	code += methodsCode

	code += "func RegisterRestHandlers() {\n"
	for _, method := range service.Methods {
		pattern := method.Name // FIXME
		code += fmt.Sprintf("http.HandleFunc(%#v, ripo.TranslateHandler(%vHandler))\n", pattern, method.Name)
	}
	code += "}"

	{
		formattedCodeBytes, err := format.Source([]byte(code))
		if err != nil {
			return "", err
		}
		code = string(formattedCodeBytes)
	}

	return code, nil
}

func generateMethodCode(method *Method) (string, error) {
	code := "func " + method.Name + "Handler(req ripo.Request) (*ripo.Response, error) {\n"
	code += fmt.Sprintf("grpcReq := &%v{}\n", method.RequestName)
	for _, param := range method.RequestParams {
		callCode := ""
		varName := "value" // isolated in a block
		varNameNil := "valueNil"
		valueExpr := "*" + varName
		typ := param.Type
		// isPointer := false
		// if kind == reflect.Ptr {
		// 	isPointer = true
		// }
		enableVarStatement := false
		switch typ {
		case t_string:
			callCode = "req.GetString(%#v)"
		case t_int:
			callCode = "req.GetInt(%#v)"
		case t_int64:
			callCode = "req.GetInt(%#v)"
			valueExpr = fmt.Sprintf("int64(*%v)", varName)
		case t_int32:
			callCode = "req.GetInt(%#v)"
			valueExpr = fmt.Sprintf("int32(*%v)", varName)
		case t_float64:
			callCode = "req.GetFloat(%#v)"
		case t_float32:
			callCode = "req.GetFloat(%#v)"
			valueExpr = fmt.Sprintf("float32(*%v)", varName)
		case t_bool:
			callCode = "req.GetBool(%#v)"
		case t_stringSlice:
			callCode = "req.GetStringList(%#v)"
			valueExpr = varName
		default:
			enableVarStatement = true
			typeExpr := fmt.Sprintf("reflect.TypeOf(%v)", varNameNil) // correct?
			callCode = "req.GetObject(%#v, " + typeExpr + ")"
			valueExpr = varName + ".(" + typ + ")"
			// if strings.HasPrefix(typ, "[]")
			// if strings.HasPrefix(typ, "*")
		}
		if callCode == "" {
			log.Printf("unrecognized type %v for param %#v", typ, param.Name)
			continue
		}
		callCode = fmt.Sprintf(callCode, param.Name)

		// TODO: fix varName to make sure it's a valid var name
		code += "\t{\n"
		if enableVarStatement {
			code += fmt.Sprintf("\t\tvar %v %v\n", varNameNil, typ)
		}
		code += fmt.Sprintf("\t\t%v, err := %v\n", varName, callCode)
		code += "\t\tif err != nil {return nil, err}\n"
		code += fmt.Sprintf("\t\tgrpcReq.%v = %v\n", param.Name, valueExpr)
		code += "\t}\n"
	}
	code += "\tlog.Println(\"grpcReq =\", grpcReq)" + "\n"

	// TODO: call server.{method}
	code += "\treturn nil, nil // FIXME\n"
	code += "}"

	return code, nil
}
