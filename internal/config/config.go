package config

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var cfg *App

func Get() *App {
	if cfg != nil {
		return cfg
	}

	path := flag.String("config", "/etc/kube-finder/config.yaml", "the absolute path of config.yaml")
	flag.Parse()

	content, err := ioutil.ReadFile(*path)
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(content, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	return cfg
}
