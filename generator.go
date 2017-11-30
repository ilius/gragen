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
		methodCode, err := generateMethodCode(service, method)
		if err != nil {
			return "", err
		}
		code += "\n" + methodCode + "\n\n"
	}
	return code, nil
}

func genClientFromServerFunc(service *Service) (string, error) {
	clientName := service.ClientName
	clientNameLower := strings.ToLower(string(clientName[0])) + clientName[1:]
	serverName := service.ServerName
	code := ""
	structName := clientNameLower + "ByServerImp"
	code += fmt.Sprintf(
		`type %v struct {
			srv %s
		}`,
		structName,
		serverName,
	) + "\n\n"
	for _, method := range service.Methods {
		code += fmt.Sprintf(
			`func (c *%s) %s(ctx context.Context, in *%s, opts ...grpc.CallOption) (*%s, error) {
				return c.srv.%s(ctx, in)
			}`,
			structName,
			method.Name,
			method.RequestName,
			method.ResponseName,
			method.Name,
		) + "\n\n"
	}
	code += fmt.Sprintf(
		`func New%vFromServer(srv %s) %s {
			return &%s{srv: srv}
		}`,
		clientName,
		serverName,
		clientName,
		structName,
	) + "\n\n"
	return code, nil
}

// func genClientOrServerSwitchBlock(service *Service, code string) string {
// 	return fmt.Sprintf(
// 		`switch client := clientArg.(type) {
// 	case %v: // , %v
// 		%v
// 	default:
// 		panic(fmt.Sprintf("invalid client type %%T, must be %v or %v", client))
// 	}
// 	`,
// 		service.ClientName, service.ServerName,
// 		code,
// 		service.ClientName, service.ServerName,
// 	)
// }

func generateServiceCode(service *Service) (string, error) {
	methodsCode, err := generateServiceMethodsCode(service)
	if err != nil {
		return "", err
	}

	imports := []string{
		// "fmt",
		"log",
		"net/http",
		"golang.org/x/net/context",
		"google.golang.org/grpc",
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

	code += fmt.Sprintf("func RegisterRestHandlers(client %v) {\n", service.ClientName)
	for _, method := range service.Methods {
		pattern := method.Name // FIXME
		code += fmt.Sprintf("http.HandleFunc(%#v, ripo.TranslateHandler(NewRestHandler_%v(client)))\n", pattern, method.Name)
	}
	code += "}\n\n"

	s2c_code, err := genClientFromServerFunc(service)
	if err != nil {
		return "", err
	}
	code += s2c_code
	code += "\n\n"

	{
		formattedCodeBytes, err := format.Source([]byte(code))
		if err != nil {
			// FIXME
			return code, nil
		}
		code = string(formattedCodeBytes)
	}

	return code, nil
}

func generateMethodCode(service *Service, method *Method) (string, error) {
	headerCode := fmt.Sprintf(
		"func NewRestHandler_%v(client %v) ripo.Handler {\n",
		method.Name,
		service.ClientName,
	)
	code := "\treturn func(req ripo.Request) (*ripo.Response, error) {\n"
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
		code += "\t\t{\n"
		if enableVarStatement {
			code += fmt.Sprintf("\t\tvar %v %v\n", varNameNil, typ)
		}
		code += fmt.Sprintf("\t\t%v, err := %v\n", varName, callCode)
		code += "\t\t\tif err != nil {return nil, err}\n"
		code += fmt.Sprintf("\t\tgrpcReq.%v = %v\n", param.Name, valueExpr)
		code += "\t\t}\n"
	}
	code += "\tlog.Println(\"grpcReq =\", grpcReq)" + "\n"

	code += fmt.Sprintf("\tgrpcRes, err := client.%v(context.Background(), grpcReq)\n", method.Name)
	code += "\t\tif err != nil { return nil, err }\n"
	code += "\t\treturn &ripo.Response{Data: grpcRes}, nil\n"
	code += "\t}"

	code = headerCode + code + "}"

	return code, nil
}
