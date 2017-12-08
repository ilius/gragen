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

var code_restResponseWrapper = `
var restJsonMarshaler = jsonpb.Marshaler{}

type restResponseWrapper struct {
	grpcRes interface{}
}

func (rw *restResponseWrapper) MarshalJSON() ([]byte, error) {
	protoMsg, ok := rw.grpcRes.(proto.Message)
	if !ok {
		return json.Marshal(rw.grpcRes)
	}
	buf := bytes.NewBuffer(nil)
	err := restJsonMarshaler.Marshal(buf, protoMsg)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}`

var code_GontextFromRest = `const restHeaderToContextPrefix = "rest-header-"


func GontextFromRest(req ripo.Request) (context.Context, error) {
	headerMap := map[string]string{}
	for _, key := range req.HeaderKeys() {
		value := req.Header(key)
		headerMap[restHeaderToContextPrefix+key] = value
	}
	md := metadata.New(headerMap)
	ctx := context.Background()
	ctx = metadata.NewIncomingContext(ctx, md)
	ctx = metadata.NewOutgoingContext(ctx, md)
	return ctx, nil
}`

var code_getRestError = `// getRestError: convert grpc error to rest
func getRestError(err error) ripo.RPCError {
	st, ok := status.FromError(err)
	if !ok {
		return ripo.NewError(ripo.Unknown, "", err)
	}
	return ripo.NewError(ripo.Code(int32(st.Code())), st.Message(), err)
}`

var code_handleRest = `func handleRest(router *httprouter.Router, method string, path string, handler ripo.Handler) {
	handlerFunc := ripo.TranslateHandler(handler)
	router.Handle(
		method,
		path,
		func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
			r.ParseForm()
			for _, p := range params {
				r.Form.Add(p.Key, p.Value)
			}
			handlerFunc(w, r)
		},
	)
}`

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
		// "log",
		"bytes",
		"encoding/json",
		"net/http",
		"golang.org/x/net/context",
		"google.golang.org/grpc",
		"google.golang.org/grpc/status",
		"google.golang.org/grpc/metadata",
		"github.com/golang/protobuf/jsonpb",
		"github.com/golang/protobuf/proto",
		"github.com/ilius/ripo",
		"github.com/julienschmidt/httprouter",
	}

	if strings.Contains(methodsCode, "reflect.") {
		imports = append(imports, "reflect")
	}

	code := "package " + service.Name + "\n\n"
	code += "import (\n"
	for _, imp := range imports {
		code += "\t" + `"` + imp + `"` + "\n"
	}
	code += ")\n\n"

	code += code_restResponseWrapper + "\n\n"

	code += code_GontextFromRest + "\n\n"

	code += code_getRestError + "\n\n"

	code += code_handleRest + "\n\n"

	code += methodsCode + "\n\n"

	code += fmt.Sprintf("func RegisterRestHandlers(client %v, router *httprouter.Router) {\n", service.ClientName)
	for _, method := range service.Methods {
		pattern := method.Name // FIXME
		pattern = strings.ToLower(pattern)
		code += fmt.Sprintf(`handleRest(router, "GET", %#v, NewRest_%v(client))`,
			"/"+pattern,
			method.Name,
		) + "\n"
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
		"func NewRest_%v(client %v) ripo.Handler {\n",
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
		callCode = fmt.Sprintf(callCode, param.JsonKey)

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
	code += "\t\tctx, err := GontextFromRest(req)\n"
	code += "\t\tif err != nil { return nil, err }\n"
	code += fmt.Sprintf("\t\tgrpcRes, err := client.%v(ctx, grpcReq)\n", method.Name)
	code += "\t\tif err != nil { return nil, getRestError(err) }\n"
	code += "\t\treturn &ripo.Response{Data: &restResponseWrapper{grpcRes}}, nil\n"
	code += "\t}"

	code = headerCode + code + "}"

	return code, nil
}
