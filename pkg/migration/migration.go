package migration

import (
	"bytes"
	_ "embed"
	"os"
	"text/template"

	"github.com/rakateja/repogen/pkg/common"
	"github.com/rakateja/repogen/pkg/entity"
)

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

//go:embed postgresql_migration.txt
var sqlMigrationTemplate string

func Handler(in []entity.Entity) error {
	out, err := toMigration(in)
	if err != nil {
		return err
	}
	if err = os.WriteFile("migrations/1.sql", []byte(out), 0644); err != nil {
		return err
	}
	return nil
}

func toMigration(in []entity.Entity) (string, error) {
	tableEntity := ToTableEntity(in)
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

func ToTableEntity(entityList []entity.Entity) []TableEntity {
	var tableEntityList []TableEntity
	for _, entity := range entityList {
		var columns []TableEntityColumn
		for _, field := range entity.Fields {
			columns = append(columns, TableEntityColumn{
				Name:     common.ToSnakeCase(field.ID),
				Type:     ToPostgreSQLColumnType(field.Type),
				Nullable: false,
			})
		}
		tableEntityList = append(tableEntityList, TableEntity{
			Name:    common.ToSnakeCase(entity.ID),
			Columns: columns,
		})
	}
	return tableEntityList
}

func ToPostgreSQLColumnType(t string) string {
	switch t {
	case "UUID":
		return "CHAR(36)"
	case "String":
		return "VARCHAR(100)"
	case "Text":
		return "TEXT"
	case "Timestamp":
		return "TIMESTAMPTZ"
	case "Bool":
		return "BOOLEAN"
	case "Int":
		return "INT"
	case "FLOAT":
		return "FLOAT"
	}
	return ""
}
