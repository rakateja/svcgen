package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/rakateja/repogen/pkg/entity"
	"github.com/rakateja/repogen/pkg/migration"
	"github.com/rakateja/repogen/pkg/repo"
	"github.com/rakateja/repogen/pkg/service"
	"github.com/rakateja/repogen/pkg/twirprpc"
)

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
			if err := migration.Handler(cmdInput.List, pkgName); err != nil {
				log.Fatalf("%v", err)
			}
			if err := repo.Handler(rootModule, pkgName, cmdInput.List); err != nil {
				log.Fatalf("%v", err)
			}
			if err := service.Handler(rootModule, pkgName, cmdInput.List); err != nil {
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

func validateFilter() error {
	return nil
}
