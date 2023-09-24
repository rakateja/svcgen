package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"

	"github.com/rakateja/repogen/pkg/entity"
	"github.com/rakateja/repogen/pkg/migration"
	"github.com/rakateja/repogen/pkg/repo"
	"github.com/rakateja/repogen/pkg/twirprpc"
)

//go:embed product.yaml
var input string

type CmdInput struct {
	List []entity.Entity `yaml:"list"`
}

var cmdOutString string
var pkgName string
var targetDir string
var protoDir string
var rootModule string
var yamlFileInput string

func init() {
	flag.StringVar(&cmdOutString, "out", "rpc,repo", "the generated files whether they are repo, rpc or migration files")
	flag.StringVar(&pkgName, "pkg", "FooBar", "package name")
	flag.StringVar(&targetDir, "target", "out", "target directory")
	flag.StringVar(&protoDir, "proto_dir", "protos", "target directory of protobuf files")
	flag.StringVar(&rootModule, "root_module", "github.com/rakateja/foo", "The root module of the target service")
	flag.StringVar(&yamlFileInput, "input", "entity.yaml", "YAML file input")
	flag.Parse()
}

func main() {
	bt, err := os.ReadFile(yamlFileInput)
	if err != nil {
		log.Fatalf("%v", err)
	}

	var cmdInput CmdInput
	err = yaml.Unmarshal(bt, &cmdInput)
	if err != nil {
		log.Fatalf("%v", err)
	}
	if err := validateEntityList(cmdInput.List); err != nil {
		log.Fatalf("%v", err)
	}
	cmdOutFiles := strings.Split(cmdOutString, ",")
	fmt.Printf("cmd out files: %v %s\n", cmdOutFiles, targetDir)
	for _, outFile := range cmdOutFiles {
		switch outFile {
		case "twirp":
			if err := twirprpc.Handler(targetDir, protoDir, pkgName, cmdInput.List); err != nil {
				log.Fatalf("%v", err)
			}
		case "repo":
			if err := migration.Handler(cmdInput.List); err != nil {
				log.Fatalf("%v", err)
			}
			if err := repo.Handler(rootModule, pkgName, cmdInput.List); err != nil {
				log.Fatalf("%v", err)
			}
			log.Printf("repo is selected.")
		}
	}
}

func validateEntityList(list []entity.Entity) error {
	entityNames := make(map[string]entity.Entity, 0)
	for _, entity := range list {
		entityNames[entity.ID] = entity
	}
	for _, entity := range list {
		for _, c := range entity.Childs {
			if _, exist := entityNames[c.Type]; !exist {
				return fmt.Errorf("%s type couldn't be found", c.Type)
			}
		}
	}
	return nil
}

func writeToFile(fileName, pkgName, content string) error {
	if err := os.WriteFile(fmt.Sprintf("out/%s/%s", pkgName, fileName), []byte(content), 0644); err != nil {
		return err
	}
	return nil
}

func mainPrev() {
	var cmdInput CmdInput
	err := yaml.Unmarshal([]byte(input), &cmdInput)
	if err != nil {
		log.Fatalf("%v", err)
	}
	pkgName := "product"
	pkgs := packagesFromEntityList(cmdInput.List)
	modelOut, err := ModelGen(cmdInput.List, pkgName, pkgs)
	if err != nil {
		log.Fatalf("%v", err)
	}

	err = os.MkdirAll(fmt.Sprintf("out/domains/%s", pkgName), 0777)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = os.WriteFile(fmt.Sprintf("out/domains/%s/model.go", pkgName), []byte(modelOut), 0644)
	if err != nil {
		log.Fatalf("%v", err)
	}
	repoOut, err := RepoGen(cmdInput.List, pkgName)
	if err != nil {
		log.Fatalf("%v", err)
	}
	err = os.WriteFile(fmt.Sprintf("out/domains/%s/repository.go", pkgName), []byte(repoOut), 0644)
	if err != nil {
		log.Fatalf("%v", err)
	}
	sqlRepoOut, err := SQLRepoGen(cmdInput.List, pkgName)
	if err != nil {
		log.Fatalf("%v", err)
	}
	err = os.WriteFile(fmt.Sprintf("out/domains/%s/sql.go", pkgName), []byte(sqlRepoOut), 0644)
	if err != nil {
		log.Fatalf("%v", err)
	}
	sqlMigrationOut, err := ToPostgreSQLMigration(cmdInput.List)
	if err != nil {
		log.Fatalf("%v", err)
	}
	err = os.WriteFile("out/migrations/1.sql", []byte(sqlMigrationOut), 0644)
	if err != nil {
		log.Fatalf("%v", err)
	}
}

// Model generator
//
//go:embed templates/model.txt
var modelTemplate string

//go:embed templates/repo.txt
var repoTemplate string

//go:embed templates/sql.txt
var sqlRepoTemplate string

//go:embed templates/migration.txt
var sqlMigrationTemplate string

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

func ModelGen(entityList []entity.Entity, pkgName string, pkgs []string) (string, error) {
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
				Type:    ToGoType(field.Type),
				JsonTag: field.ID,
				DBTag:   ToSnakeCase(field.ID),
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

func RepoGen(entityList []entity.Entity, pkgName string) (string, error) {
	parent, childs, err := ToStructs(entityList)
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
			"\"github.com/rakateja/repogen/out/page\"",
		},
		Parent: parent,
		Childs: childs,
	})
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func ToPostgreSQLMigration(entityList []entity.Entity) (string, error) {
	tableEntity := ToTableEntity(entityList)
	t, err := template.New("base").
		Funcs(template.FuncMap{
			"sqlnullable": func(arg bool) string {
				if arg {
					return "NULL"
				}
				return "NOT NULL"
			},
		}).
		Parse(sqlMigrationTemplate)
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	if err = t.Execute(&out, tableEntity); err != nil {
		return "", err
	}
	return out.String(), nil
}

func SQLRepoGen(entityList []entity.Entity, pkgName string) (string, error) {
	parent, childs, err := ToStructs(entityList)
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
		Parse(sqlRepoTemplate)
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	err = t.Execute(&out, RepoTemplateData{
		PackageName: pkgName,
		ImportedPackages: []string{
			"\"context\"",
			"\"github.com/rakateja/repogen/out/page\"",
			"\"github.com/rakateja/repogen/out/database\"",
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
	for _, entity := range entityList {
		for _, f := range entity.Fields {
			if f.Type == "Timestamp" {
				res = append(res, "\"time\"")
			}
		}
	}
	return
}

func ToGoType(t string) string {
	switch t {
	case "UUID":
		return "string"
	case "String":
		return "string"
	case "Timestamp":
		return "time.Time"
	case "Bool":
		return "bool"
	}
	return ""
}

func ToPostgreSQLColumnType(t string) string {
	switch t {
	case "UUID":
		return "CHAR(36)"
	case "String":
		return "VARCHAR(100)"
	case "Timestamp":
		return "TIMESTAMPTZ"
	case "Bool":
		return "BOOL"
	}
	return ""
}

func ToSnakeCase(str string) string {
	var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func ToColumnNames(fields []EntityFieldData) []string {
	res := make([]string, 0)
	for _, f := range fields {
		res = append(res, f.DBTag)
	}
	return res
}

func ValueArguments(fields []EntityFieldData) string {
	return ""
}

func ToNullablePQ(arg bool) string {
	if arg {
		return "NULL"
	}
	return "NOT NULL"
}

func ToTableEntity(entityList []entity.Entity) []TableEntity {
	var tableEntityList []TableEntity
	for _, entity := range entityList {
		var columns []TableEntityColumn
		for _, field := range entity.Fields {
			columns = append(columns, TableEntityColumn{
				Name:     ToSnakeCase(field.ID),
				Type:     ToPostgreSQLColumnType(field.Type),
				Nullable: false,
			})
		}
		tableEntityList = append(tableEntityList, TableEntity{
			Name:    ToSnakeCase(entity.ID),
			Columns: columns,
		})
	}
	return tableEntityList
}

func ToStructs(entityList []entity.Entity) (p EntityData, childs []EntityData, err error) {
	var parent *EntityData
	for _, entity := range entityList {
		var fields []EntityFieldData
		for _, field := range entity.Fields {
			fields = append(fields, EntityFieldData{
				ID:      cases.Title(language.Und, cases.NoLower).String(field.ID),
				Type:    ToGoType(field.Type),
				JsonTag: field.ID,
				DBTag:   ToSnakeCase(field.ID),
			})
		}
		if entity.IsParent {
			parent = &EntityData{
				StructName: entity.ID,
				TableName:  ToSnakeCase(entity.ID),
				Fields:     fields,
			}
			continue
		}
		childs = append(childs, EntityData{
			StructName:    entity.ID,
			RootTableName: parent.TableName,
			TableName:     ToSnakeCase(entity.ID),
			Fields:        fields,
		})
	}
	if parent == nil {
		err = fmt.Errorf("entity parent couldn't be found")
		return
	}
	return *parent, childs, nil
}

func IsTimestampFields(field string) bool {
	tsFields := map[string]bool{
		"createdAt": true,
		"updatedAt": true,
		"createdBy": true,
		"updatedBy": true,
	}
	return tsFields[field]
}

// Table migration
type TableEntity struct {
	Name    string
	Columns []TableEntityColumn
}

type TableEntityColumn struct {
	Name         string
	Type         string
	Nullable     bool
	DefaultValue string
}
