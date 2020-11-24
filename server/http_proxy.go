package server

import (
	"encoding/base64"
	"net"
	"strings"

	"github.com/jiuzhou-zhao/lumos.git/config"
	"github.com/sirupsen/logrus"
)

type HTTPProxy struct {
	cfg         *config.Config
	credentials map[string]interface{}
}

func NewHTTPProxy(cfg *config.Config) *HTTPProxy {
	if cfg.EffectMode != config.ModeLocal && cfg.EffectMode != config.ModeProxy {
		logrus.Fatalf("invalid mode: %v", cfg.Mode)
	}

	credentials := make(map[string]interface{})
	for _, credential := range cfg.Credentials {
		credential = base64.StdEncoding.EncodeToString([]byte(credential))
		logrus.Debugf("use credential: %v\n", credential)
		credentials[credential] = true
	}
	return &HTTPProxy{
		cfg:         cfg,
		credentials: credentials,
	}
}

func (proxy *HTTPProxy) Serve() {
	tcpServer := NewTCPServer()

	var clientChan <-chan net.Conn
	var err error

	if proxy.cfg.Secure.EnableTLSServer {
		clientChan, err = tcpServer.StartTLS(proxy.cfg.ProxyAddress, proxy.cfg.Secure.Server)
	} else {
		clientChan, err = tcpServer.Start(proxy.cfg.ProxyAddress)
	}

	if err != nil {
		logrus.Fatalf("start tcp server failed: %v", err)
	}

	logrus.Infof("%v listen on: %v\n", proxy.cfg.Mode, proxy.cfg.ProxyAddress)

	for client := range clientChan {
		go NewHTTPProxyConn(client, proxy).Server()
	}
}

func (proxy *HTTPProxy) NeedAuth() bool {
	return len(proxy.credentials) > 0
}

func (proxy *HTTPProxy) ValidateCredential(basicCredential string) bool {
	c := strings.Split(basicCredential, " ")
	if len(c) == 2 && strings.EqualFold(c[0], "Basic") {
		if _, ok := proxy.credentials[c[1]]; ok {
			return true
		}
	}
	return false
}
