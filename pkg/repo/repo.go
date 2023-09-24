package repo

import (
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/rakateja/repogen/pkg/entity"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

//go:embed model.txt
var modelTemplate string

//go:embed repo.txt
var repoTemplate string

//go:embed postgresql.txt
var postgreSQLRepoTemplate string

type ModelTemplateData struct {
	PackageName      string
	ImportedPackages []string
	Parent           EntityData
	Child            []EntityData
}

type RepoTemplateData struct {
	PackageName      string
	ImportedPackages []string
	Parent           EntityData
	Childs           []EntityData
}

type EntityFieldData struct {
	ID      string
	Type    string
	JsonTag string
	DBTag   string
}

type EntityData struct {
	StructName    string
	TableName     string
	RootTableName string
	Fields        []EntityFieldData
}

func Handler(rootModule, pkgName string, in []entity.Entity) error {
	err := os.MkdirAll(fmt.Sprintf("domains/%s", pkgName), 0777)
	if err != nil {
		fmt.Println(err)
		return err
	}
	modelOut, err := modelGen(in, pkgName)
	if err != nil {
		return err
	}
	err = os.WriteFile(fmt.Sprintf("domains/%s/model.go", pkgName), []byte(modelOut), 0644)
	if err != nil {
		log.Fatalf("%v", err)
	}
	repoOut, err := repoGen(rootModule, pkgName, in)
	if err != nil {
		log.Fatalf("%v", err)
	}
	err = os.WriteFile(fmt.Sprintf("domains/%s/repository.go", pkgName), []byte(repoOut), 0644)
	if err != nil {
		log.Fatalf("%v", err)
	}
	sqlRepoOut, err := SQLRepoGen(rootModule, in, pkgName)
	if err != nil {
		log.Fatalf("%v", err)
	}
	err = os.WriteFile(fmt.Sprintf("domains/%s/sql.go", pkgName), []byte(sqlRepoOut), 0644)
	if err != nil {
		log.Fatalf("%v", err)
	}
	return nil
}

func repoGen(rootModule, pkgName string, in []entity.Entity) (string, error) {
	parent, childs, err := toStructs(in)
	if err != nil {
		return "", err
	}
	t, err := template.New("base").Parse(repoTemplate)
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	err = t.Execute(&out, RepoTemplateData{
		PackageName: pkgName,
		ImportedPackages: []string{
			"\"context\"",
			fmt.Sprintf("\"%s/page\"", rootModule),
		},
		Parent: parent,
		Childs: childs,
	})
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func SQLRepoGen(rootModule string, entityList []entity.Entity, pkgName string) (string, error) {
	parent, childs, err := toStructs(entityList)
	if err != nil {
		return "", err
	}
	t, err := template.New("base").
		Funcs(template.FuncMap{
			"sqlvalues": func(arg int) string {
				var values []string
				for i := 1; i <= arg; i++ {
					values = append(values, fmt.Sprintf("$%d", i))
				}
				return strings.Join(values, ", ")
			},
			"sqlcolumns": func(arg []EntityFieldData) string {
				var columns []string
				for _, f := range arg {
					columns = append(columns, f.DBTag)
				}
				return strings.Join(columns, ", \n")
			},
			"title": func(str string) string {
				return cases.Title(language.Und, cases.NoLower).String(str)
			},
		}).
		Parse(postgreSQLRepoTemplate)
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	err = t.Execute(&out, RepoTemplateData{
		PackageName: pkgName,
		ImportedPackages: []string{
			"\"context\"",
			fmt.Sprintf("\"%s/page\"", rootModule),
			fmt.Sprintf("\"%s/database\"", rootModule),
		},
		Parent: parent,
		Childs: childs,
	})
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func packagesFromEntityList(entityList []entity.Entity) (res []string) {
	m := make(map[string]bool, 0)
	for _, entity := range entityList {
		for _, f := range entity.Fields {
			if f.Type == "Timestamp" {
				m["\"time\""] = true
			}
		}
	}
	for s, _ := range m {
		res = append(res, s)
	}
	return
}

func modelGen(entityList []entity.Entity, pkgName string) (string, error) {
	pkgs := packagesFromEntityList(entityList)
	t, err := template.New("base").Parse(modelTemplate)
	if err != nil {
		return "", err
	}
	var child []EntityData
	var parent *EntityData
	for _, entity := range entityList {
		var fields []EntityFieldData
		for _, field := range entity.Fields {
			fields = append(fields, EntityFieldData{
				ID:      cases.Title(language.Und, cases.NoLower).String(field.ID),
				Type:    toGoType(field.Type),
				JsonTag: field.ID,
				DBTag:   toSnakeCase(field.ID),
			})
		}
		if entity.IsParent {
			parent = &EntityData{
				StructName: entity.ID,
				Fields:     fields,
			}
			continue
		}
		child = append(child, EntityData{
			StructName: entity.ID,
			Fields:     fields,
		})
	}
	if parent == nil {
		return "", fmt.Errorf("entity parent couldn't be found")
	}
	tmplData := ModelTemplateData{
		PackageName:      pkgName,
		ImportedPackages: pkgs,
		Parent:           *parent,
		Child:            child,
	}
	var bt bytes.Buffer
	err = t.Execute(&bt, tmplData)
	if err != nil {
		return "", err
	}
	return bt.String(), nil
}

func toGoType(t string) string {
	switch t {
	case "UUID":
		return "string"
	case "String":
		return "string"
	case "Timestamp":
		return "time.Time"
	case "Bool":
		return "bool"
	case "Int":
		return "int32"
	case "Long":
		return "int64"
	case "Float":
		return "float32"
	case "Text":
		return "string"
	}
	return ""
}

func toSnakeCase(str string) string {
	var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func toStructs(entityList []entity.Entity) (p EntityData, childs []EntityData, err error) {
	var parent *EntityData
	for _, entity := range entityList {
		var fields []EntityFieldData
		for _, field := range entity.Fields {
			fields = append(fields, EntityFieldData{
				ID:      cases.Title(language.Und, cases.NoLower).String(field.ID),
				Type:    toGoType(field.Type),
				JsonTag: field.ID,
				DBTag:   toSnakeCase(field.ID),
			})
		}
		if entity.IsParent {
			parent = &EntityData{
				StructName: entity.ID,
				TableName:  toSnakeCase(entity.ID),
				Fields:     fields,
			}
			continue
		}
		childs = append(childs, EntityData{
			StructName:    entity.ID,
			RootTableName: parent.TableName,
			TableName:     toSnakeCase(entity.ID),
			Fields:        fields,
		})
	}
	if parent == nil {
		err = fmt.Errorf("entity parent couldn't be found")
		return
	}
	return *parent, childs, nil
}
