package twirprpc

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/rakateja/repogen/pkg/entity"
)

// protobuff
type ProtoEntity struct {
	Name   string
	Fields []ProtoEntityField
}

type ProtoEntityField struct {
	Name string
	Type string
}

type TemplateData struct {
	PackageName      string
	GoPackageName    string
	EntityName       string
	ImportedPackages []string
	Protos           []ProtoEntity
}

type TwirpRpcTemplateData struct {
	PackageName string
	EntityName  string
}

//go:embed protobuf.txt
var protobuffTemplate string

//go:embed twirprpc.txt
var twirpRpcTemplate string

//go:embed parser.txt
var parserTemplate string

func Handler(targetDir, protoDir, pkgName string, entityList []entity.Entity) error {
	path := fmt.Sprintf("domains/%s", pkgName)
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.Printf("[WARN] %v", err)
	}
	path = fmt.Sprintf("protos/%s", pkgName)
	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.Printf("[WARN] %v", err)
	}

	t, err := template.New("base").
		Funcs(template.FuncMap{
			"plus_one": func(n int) int {
				return n + 1
			},
		}).
		Parse(protobuffTemplate)
	if err != nil {
		return err
	}
	entityName, err := entityFromList(entityList)
	if err != nil {
		return err
	}
	var out bytes.Buffer
	if err = t.Execute(&out, TemplateData{
		PackageName:   "dev.raka.langit101",
		GoPackageName: strings.ToLower(pkgName),
		EntityName:    entityName,
		ImportedPackages: []string{
			"google/protobuf/timestamp.proto",
		},
		Protos: toProtoBuffEntity(entityList),
	}); err != nil {
		return err
	}
	fileName := fmt.Sprintf("%s/%s.proto", protoDir, strings.ToLower(pkgName))
	if err := os.WriteFile(fileName, out.Bytes(), 0644); err != nil {
		return err
	}
	if err = toTwirpRpc(entityName, pkgName); err != nil {
		return err
	}
	if err = toParser(entityName, pkgName); err != nil {
		return err
	}
	return nil
}

func toTwirpRpc(entityName, packageName string) error {
	t, err := template.New("base").
		Parse(twirpRpcTemplate)
	if err != nil {
		return err
	}
	var out bytes.Buffer
	if err := t.Execute(&out, TwirpRpcTemplateData{
		PackageName: packageName,
		EntityName:  entityName,
	}); err != nil {
		return err
	}
	fileName := fmt.Sprintf("%s/%s/rpc.go", "domains", strings.ToLower(packageName))
	if err := os.WriteFile(fileName, out.Bytes(), 0644); err != nil {
		return err
	}
	return nil
}

func toParser(entityName, packageName string) error {
	t, err := template.New("base").
		Parse(parserTemplate)
	if err != nil {
		return err
	}
	var out bytes.Buffer
	if err := t.Execute(&out, TwirpRpcTemplateData{
		PackageName: packageName,
		EntityName:  entityName,
	}); err != nil {
		return err
	}
	fileName := fmt.Sprintf("%s/%s/parser.go", "domains", strings.ToLower(packageName))
	if err := os.WriteFile(fileName, out.Bytes(), 0644); err != nil {
		return err
	}
	return nil
}

func toProtoBuffEntity(in []entity.Entity) []ProtoEntity {
	var pb []ProtoEntity
	for _, entity := range in {
		var fields []ProtoEntityField
		for _, field := range entity.Fields {
			fields = append(fields, ProtoEntityField{
				Name: field.ID,
				Type: ToProtoBuffType(field.Type),
			})
		}
		pb = append(pb, ProtoEntity{
			Name:   entity.ID,
			Fields: fields,
		})
	}
	return pb
}

func ToProtoBuffType(str string) string {
	switch str {
	case "UUID":
		return "string"
	case "String":
		return "string"
	case "Float":
		return "float32"
	case "Int":
		return "int32"
	case "Timestamp":
		return "google.protobuf.Timestamp"
	case "Bool":
		return "bool"
	case "Text":
		return "string"
	default:
		return ""
	}
}

func entityFromList(ls []entity.Entity) (string, error) {
	for _, e := range ls {
		if e.IsParent {
			return e.ID, nil
		}
	}
	return "", errors.New("NotFound")
}
