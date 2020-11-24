package utils

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/jiuzhou-zhao/lumos.git/config"
	"io/ioutil"
	"net"
	"time"
)

func GetTLSConfig(cfg *config.TlsConfig) (certs []tls.Certificate, cas *x509.CertPool, err error) {
	cert, err := tls.LoadX509KeyPair(cfg.Cert, cfg.Key)
	if err != nil {
		return
	}

	certPool := x509.NewCertPool()
	for _, ca := range cfg.RootCAs {
		var certBytes []byte
		certBytes, err = ioutil.ReadFile(ca)
		if err != nil {
			return
		}
		certPool.AppendCertsFromPEM(certBytes)
	}
	return []tls.Certificate{cert}, certPool, nil
}

func GetServerTLSConfig(cfg *config.TlsConfig) (*tls.Config, error) {
	certs, cas, err := GetTLSConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: certs,
		ClientCAs:    cas,
	}, nil
}

func GetClientTLSConfig(cfg *config.TlsConfig) (*tls.Config, error) {
	certs, cas, err := GetTLSConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: certs,
		RootCAs:      cas,
	}, nil
}

func EasyTCPConnectServer(enableTls bool, cfg *config.TlsConfig, address string,
	timeout time.Duration) (conn net.Conn, err error) {
	dialer := &net.Dialer{
		Timeout: timeout,
	}
	if enableTls {
		var tlsConfig *tls.Config
		tlsConfig, err = GetClientTLSConfig(cfg)
		if err != nil {
			return
		}
		conn, err = tls.DialWithDialer(dialer, "tcp", address, tlsConfig)
	} else {
		conn, err = dialer.Dial("tcp", address)
	}
	return conn, err
}
