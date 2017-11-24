package main

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"

	_ "github.com/ilius/gragen/proto_registry"
)

func getMessageType(msgName string) (reflect.Type, error) {
	msgName = strings.TrimLeft(msgName, ".")
	msgTypePtr := proto.MessageType(msgName)
	if msgTypePtr == nil {
		return nil, fmt.Errorf("<< Error: Could not find proto message type %#v\n", msgName)
	}
	msgType := msgTypePtr
	if msgType.Kind() == reflect.Ptr {
		msgType = msgType.Elem()
	}
	return msgType, nil
}

type Param struct {
	Name      string
	GRPCName  string
	Type      reflect.Type
	OmitEmpty bool
}

type Method struct {
	Name         string
	Descriptor   *descriptor.MethodDescriptorProto
	RequestType  reflect.Type
	ResponseType reflect.Type
	Params       []*Param
}

func processProtoService(service *descriptor.ServiceDescriptorProto) ([]*Method, error) {
	methods := []*Method{}
	for _, methodDesc := range service.GetMethod() {
		requestType, err := getMessageType(methodDesc.GetInputType())
		if err != nil {
			return nil, err
		}
		responseType, err := getMessageType(methodDesc.GetOutputType())
		if err != nil {
			return nil, err
		}
		if requestType.Kind() != reflect.Struct {
			return nil, fmt.Errorf("unknown request type %v with kind %v", requestType, requestType.Kind())
		}

		params := []*Param{}
		for i := 0; i < requestType.NumField(); i++ {
			field := requestType.Field(i)
			jsonTagParts := strings.Split(field.Tag.Get("json"), ",")
			if len(jsonTagParts) == 0 || jsonTagParts[0] == "" {
				continue // no json tag for this field
			}
			fieldParamName := jsonTagParts[0]
			isOmitEmpty := tagHasFlag(jsonTagParts, "omitempty")

			params = append(params, &Param{
				Name:      fieldParamName,
				GRPCName:  field.Name,
				Type:      field.Type,
				OmitEmpty: isOmitEmpty,
			})
		}

		methods = append(methods, &Method{
			Name:         methodDesc.GetName(),
			Descriptor:   methodDesc,
			RequestType:  requestType,
			ResponseType: responseType,
			Params:       params,
		})
	}
	return methods, nil
}

var (
	t_string      = reflect.TypeOf("")
	t_int         = reflect.TypeOf(int(0))
	t_int32       = reflect.TypeOf(int32(0))
	t_int64       = reflect.TypeOf(int64(0))
	t_float64     = reflect.TypeOf(float64(0))
	t_float32     = reflect.TypeOf(float32(0))
	t_bool        = reflect.TypeOf(false)
	t_stringSlice = reflect.TypeOf([]string{})
)

func generateMethodCode(method *Method) (string, error) {
	code := "func " + method.Name + "Handler(req ripo.Request) (*ripo.Response, error) {\n"
	requestTypeParts := strings.Split(method.RequestType.String(), ".")
	if len(requestTypeParts) != 2 {
		return "", fmt.Errorf("unexpected RequestType = %v", method.RequestType)
	}
	code += fmt.Sprintf("grpcReq := &%v{}\n", requestTypeParts[1])
	for _, param := range method.Params {
		callCode := ""
		varName := param.Name
		valueExpr := "*" + varName
		typ := param.Type
		kind := typ.Kind()
		// isPointer := false
		// if kind == reflect.Ptr {
		// 	isPointer = true
		// }
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
		default:
			switch kind {
			case reflect.Slice:
			case reflect.Struct:
			}
		}
		if callCode == "" {
			log.Printf("unrecognized type %v for param %#v", typ, param.Name)
			continue
		}
		callCode = fmt.Sprintf(callCode, param.Name)

		// TODO: fix varName to make sure it's a valid var name
		code += "\t{\n"
		code += fmt.Sprintf("\t\t%v, err := %v\n", varName, callCode)
		code += "\t\tif err != nil {return nil, err}\n"
		code += fmt.Sprintf("\t\tgrpcReq.%v = %v\n", param.GRPCName, valueExpr)
		code += "\t}\n"
	}
	code += "\tlog.Println(\"grpcReq =\", grpcReq)" + "\n"

	// TODO: call server.{method}
	code += "\treturn nil, nil // FIXME\n"
	code += "}"

	return code, nil
}

func generateServiceCode(service *descriptor.ServiceDescriptorProto, methods []*Method) (string, error) {
	imports := []string{
		"log",
		"net/http",
		"github.com/ilius/ripo",
	}
	code := "package " + service.GetName() + "\n"
	code += "import (\n"
	for _, imp := range imports {
		code += "\t" + `"` + imp + `"` + "\n"
	}
	code += ")"

	for _, method := range methods {
		methodCode, err := generateMethodCode(method)
		if err != nil {
			return "", err
		}
		code += "\n" + methodCode + "\n\n"
	}

	code += "func main() {\n"
	for _, method := range methods {
		pattern := method.Name // FIXME
		code += fmt.Sprintf("http.HandleFunc(%#v, ripo.TranslateHandler(%vHandler))\n", pattern, method.Name)
	}
	code += "}"

	return code, nil
}

func parseProtoFile(protoPath string) error {
	fd := &descriptor.FileDescriptorProto{}
	protoGzBytes := proto.FileDescriptor(protoPath)
	if protoGzBytes == nil {
		return fmt.Errorf("proto file %#v not found or not registered", protoPath)
	}
	protoBytes, err := gunzipBytes(protoGzBytes)
	if err != nil {
		return err
	}
	err = proto.Unmarshal(protoBytes, fd)
	if err != nil {
		return err
	}
	for _, service := range fd.GetService() {
		methods, err := processProtoService(service)
		if err != nil {
			return err
		}
		goText, err := generateServiceCode(service, methods)
		if err != nil {
			return err
		}
		fmt.Println(goText)
		// ioutil.WriteFile(
	}
	return nil
}
