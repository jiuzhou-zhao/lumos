package main

import (
	"github.com/jiuzhou-zhao/lumos.git/config"
	"github.com/jiuzhou-zhao/lumos.git/server"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"path"
)

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		logrus.Fatalf("error reading config: %s", err)
	}

	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetLevel(logrus.DebugLevel)

	logrus.Infof("Using configuration file '%s'\n", viper.ConfigFileUsed())

	var cfg config.Config
	err = viper.Unmarshal(&cfg)
	if err != nil {
		logrus.Fatalf("error unmarshal config: %s", err)
	}

	dir, _ := path.Split(viper.ConfigFileUsed())
	cfg.Fix(dir)

	if cfg.EffectMode == config.ModeLocal || cfg.EffectMode == config.ModeRelay {
		transProxy := server.NewTransProxy(&cfg)
		transProxy.Serve()
	} else {
		httpProxyServer := server.NewHTTPProxy(&cfg)
		httpProxyServer.Serve()
	}
}
