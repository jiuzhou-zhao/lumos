package ft

import (
	"bytes"
	"encoding/gob"
	"github.com/stretchr/testify/assert"
	"path"
	"testing"

	"github.com/jiuzhou-zhao/lumos.git/config"
	"github.com/jiuzhou-zhao/lumos.git/server"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func TestAllProxy(t *testing.T) {
	t.SkipNow()

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")

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

	//
	fnStarServer := func(cfg *config.Config) {
		if cfg.EffectMode == config.ModeLocal || cfg.EffectMode == config.ModeRelay {
			transProxy := server.NewTransProxy(cfg)
			transProxy.Serve()
		} else {
			httpProxyServer := server.NewHTTPProxy(cfg)
			httpProxyServer.Serve()
		}
	}

	fnDeepCopy := func(dst, src interface{}) error {
		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(src); err != nil {
			return err
		}
		return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
	}

	//
	cfg.Secure.EnableTLSClient = true
	cfg.Secure.EnableTLSServer = true

	//
	cfg.EffectMode = config.ModeLocal
	cfg.ProxyAddress = ":8000"
	cfg.RemoteProxyAddress = "127.0.0.1:8001"

	var cfgLocal config.Config
	err = fnDeepCopy(&cfgLocal, &cfg)
	assert.Nil(t, err)
	go fnStarServer(&cfgLocal)

	//
	cfg.EffectMode = config.ModeRelay
	cfg.ProxyAddress = ":8001"
	cfg.RemoteProxyAddress = "127.0.0.1:8002"

	var cfgRelay config.Config
	err = fnDeepCopy(&cfgRelay, &cfg)
	assert.Nil(t, err)
	go fnStarServer(&cfgRelay)

	//
	cfg.EffectMode = config.ModeHTTPProxy
	cfg.ProxyAddress = ":8002"
	cfg.RemoteProxyAddress = ""

	fnStarServer(&cfg)
}
