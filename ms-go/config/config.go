package config

import (
	"flag"
	"os"

	msLog "github.com/gaolaoge/ms-go/log"

	"github.com/BurntSushi/toml"
)

type Config struct {
	logger *msLog.Logger
	Log    map[string]any
}

var Conf = &Config{
	logger: msLog.Default(),
}

func init() {
	loadToml()
}

func loadToml() {
	configFile := flag.String("conf", "conf/app.toml", "app config file")
	flag.Parse()
	if _, err := os.Stat(*configFile); err != nil {
		Conf.logger.Info("app config file not exist")
		return
	}
	_, err := toml.DecodeFile(*configFile, Conf)
	if err != nil {
		Conf.logger.Info("app config file decode failed")
		return
	}
}
