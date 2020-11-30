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

	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.SetLevel(logrus.DebugLevel)

	logrus.Infof("Using configuration file '%s'", viper.ConfigFileUsed())

	var cfg config.Config
	err = viper.Unmarshal(&cfg)
	if err != nil {
		logrus.Fatalf("error unmarshal config: %s", err)
	}

	dir, _ := path.Split(viper.ConfigFileUsed())
	cfg.Fix(dir)

	logrus.Info("lumos mode:")
	logrus.Infof("  %v", cfg.EffectMode.String())
	logrus.Infof("  Server Use TLS: %v", cfg.Secure.TLSEnableFlag.ServerUseTLS)
	logrus.Infof("  Connect Remote Server Use TLS: %v", cfg.Secure.TLSEnableFlag.ConnectServerUseTLS)

	if cfg.EffectMode == config.ModeLocal || cfg.EffectMode == config.ModeRelay {
		transProxy := server.NewTransProxy(&cfg)
		transProxy.Serve()
	} else if cfg.EffectMode == config.ModeSocks5Proxy {
		socks5Proxy := server.NewSocks5Proxy(&cfg)
		socks5Proxy.Serve()
	} else {
		httpProxyServer := server.NewHTTPProxy(&cfg)
		httpProxyServer.Serve()
	}
}
