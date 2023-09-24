package service

import (
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/rakateja/repogen/pkg/entity"
)

//go:embed query.txt
var queryTemplate string

//go:embed manager.txt
var managerTemplate string

type TemplateData struct {
	PackageName      string
	StructName       string
	ImportedPackages []string
}

func Handler(rootModule, pkgName string, in []entity.Entity) error {
	// Query
	t, err := template.New("base").Parse(queryTemplate)
	if err != nil {
		return err
	}
	structName, err := parentName(in)
	if err != nil {
		return err
	}
	var out bytes.Buffer
	if err := t.Execute(&out, TemplateData{
		PackageName: pkgName,
		StructName:  structName,
		ImportedPackages: []string{
			"\"context\"",
			fmt.Sprintf("\"%s/page\"", rootModule),
		},
	}); err != nil {
		return err
	}
	loc := fmt.Sprintf("domains/%s/query.go", pkgName)
	err = os.WriteFile(loc, out.Bytes(), 0644)
	if err != nil {
		log.Fatalf("%v", err)
	}

	// Manager
	tM, err := template.New("base").Parse(managerTemplate)
	if err != nil {
		return err
	}
	var outManager bytes.Buffer
	if err := tM.Execute(&outManager, TemplateData{
		PackageName: pkgName,
		StructName:  structName,
		ImportedPackages: []string{
			"\"errors\"",
			"\"context\"",
		},
	}); err != nil {
		return err
	}
	locManager := fmt.Sprintf("domains/%s/manager.go", pkgName)
	err = os.WriteFile(locManager, outManager.Bytes(), 0644)
	if err != nil {
		log.Fatalf("%v", err)
	}
	return nil
}

func parentName(in []entity.Entity) (string, error) {
	for _, entity := range in {
		if entity.IsParent {
			return entity.ID, nil
		}
	}
	return in[0].ID, nil
}
