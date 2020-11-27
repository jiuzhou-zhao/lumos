package server

import (
	"net"
	"strings"

	"github.com/jiuzhou-zhao/lumos.git/config"
	"github.com/sirupsen/logrus"
)

type Socks5Proxy struct {
	cfg         *config.Config
	credentials map[string]string
}

func NewSocks5Proxy(cfg *config.Config) *Socks5Proxy {
	credentials := make(map[string]string)
	for _, credential := range cfg.Credentials {
		up := strings.SplitN(credential, ":", 2)
		if len(up) != 2 {
			logrus.Fatalf("invalid credential format: %v", credential)
		}
		credentials[up[0]] = up[1]
	}
	return &Socks5Proxy{
		cfg:         cfg,
		credentials: credentials,
	}
}

func (proxy *Socks5Proxy) Serve() {
	tcpServer := NewTCPServer()

	var clientChan <-chan net.Conn
	var err error

	if proxy.cfg.Secure.TLSEnableFlag.ServerUseTLS {
		clientChan, err = tcpServer.StartTLS(proxy.cfg.ProxyAddress, proxy.cfg.Secure.ServerTLSConfig)
	} else {
		clientChan, err = tcpServer.Start(proxy.cfg.ProxyAddress)
	}

	if err != nil {
		logrus.Fatalf("start tcp server failed: %v", err)
	}

	logrus.Infof("%v listen on: %v\n", proxy.cfg.Mode, proxy.cfg.ProxyAddress)

	for client := range clientChan {
		go NewSocks5ProxyConn(client, proxy).Serve()
	}
}

func (proxy *Socks5Proxy) NeedAuth() bool {
	return len(proxy.credentials) > 0
}

func (proxy *Socks5Proxy) ValidateCredential(userName, password string) bool {
	if pass, ok := proxy.credentials[userName]; ok {
		if pass == password {
			return true
		}
	}
	return false
}
