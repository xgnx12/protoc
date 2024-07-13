package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/xgnx12/protoc/protos"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

func main() {
	protogen.Options{}.Run(func(gen *protogen.Plugin) error {
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			generateFile(gen, f)
		}
		return nil
	})
}

func generateFile(gen *protogen.Plugin, file *protogen.File) {
	if len(file.Services) == 0 {
		return
	}
	// 获取生成文件的相对路径
	parts := strings.Split(file.GeneratedFilenamePrefix, "/")
	fileName := parts[(len(parts)-1)] + "_echo.go"
	fileName = filepath.Join(filepath.Dir(file.Desc.Path()), fileName)
	g := gen.NewGeneratedFile(fileName, file.GoImportPath)

	g.P("// Code generated by protoc-gen-echo. DO NOT EDIT.")
	g.P()
	g.P("package ", file.GoPackageName)
	g.P()
	g.P(`import (`)
	g.P(`    "net/http"`)
	g.P(`    "github.com/labstack/echo/v4"`)
	g.P(`    "context"`)
	g.P(`)`)
	g.P()

	for _, service := range file.Services {
		generateServiceInterface(g, service)
		generateHandlerRouter(g, service)
	}
}

func generateServiceInterface(g *protogen.GeneratedFile, service *protogen.Service) {
	g.P("type ", service.GoName, "Interface interface {")
	for _, method := range service.Methods {
		g.P(method.GoName, "(ctx context.Context, req *", method.Input.GoIdent, ") ", "(", method.Output.GoIdent, ",error)")
	}
	g.P("}")
	g.P()
}

func generateHandlerRouter(g *protogen.GeneratedFile, service *protogen.Service) {
	g.P("func Register", service.GoName, "(e *echo.Echo, server ", service.GoName, "Interface) {")
	for _, method := range service.Methods {
		httpMethod := proto.GetExtension(method.Desc.Options(), protos.E_HttpMethod).(string)
		if httpMethod == "" {
			logError("must specify http_method option for rpb method")
		}
		if !isValidHttpMethod(httpMethod) {
			logError(fmt.Sprintf("`%s` is no valid http method", httpMethod))
		}
		g.P(`e.`, strings.ToUpper(httpMethod), `("`, camelToSnake(service.GoName), "/", camelToSnake(method.GoName), `", func(c echo.Context) error {`)
		g.P("    req := new(", method.Input.GoIdent, ")")
		g.P("    if err := c.Bind(req); err != nil {")
		g.P("       return err")
		g.P("    }")
		g.P("    resp, err := server.", method.GoName, "(c.Request().Context(), req)")
		g.P("    if err != nil {")
		g.P("   	 return err")
		g.P("    }")
		g.P("    return c.JSON(http.StatusOK, resp)")
		g.P("})")
	}
	g.P("}")
}

func isValidHttpMethod(method string) bool {
	for _, m := range []string{"get", "post", "put", "delete", "head", "option"} {
		if strings.ToLower(method) == m {
			return true
		}
	}
	return false

}
func logError(err string) {
	log.SetOutput(os.Stderr)
	log.Println(err)
	panic(err)
}

func camelToSnake(s string) string {
	var result string
	for i, c := range s {
		if i > 0 && c >= 'A' && c <= 'Z' {
			result += "_"
		}
		result += string(c)
	}
	return strings.ToLower(result)
}
