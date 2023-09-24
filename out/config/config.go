package config

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Host         string `envconfig:"host" default:"localhost"`
	Port         int    `envconfig:"port" default:"5434"`
	Database     string `envconfig:"database" default:"repogen"`
	User         string `envconfig:"user" default:"root"`
	Password     string `envconfig:"password" default:"root-is-not-used"`
	MaxOpenConns int    `envconfig:"max_open_conn" default:"100"`
	MaxIdleConns int    `envconfig:"max_idle_conn" default:"10"`
}

func NewConfig() Config {
	var conf Config
	err := envconfig.Process("ORDER", &conf)
	if err != nil {
		log.Fatalf("%v", err)
	}
	return conf
}
