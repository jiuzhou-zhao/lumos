package server

import (
	"crypto/tls"
	"errors"
	"net"
	"sync"

	"github.com/jiuzhou-zhao/lumos.git/config"
	"github.com/jiuzhou-zhao/lumos.git/utils"
	"github.com/sirupsen/logrus"
)

type TCPServer struct {
	wg         sync.WaitGroup
	listener   net.Listener
	clientChan chan net.Conn
}

func NewTCPServer() *TCPServer {
	return &TCPServer{
		clientChan: make(chan net.Conn, 10),
	}
}

func (svr *TCPServer) Start(address string) (<-chan net.Conn, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		logrus.Errorf("listen on %v failed: %v", address, err)
		return nil, err
	}
	svr.listener = listener

	return svr.startEntry()
}

func (svr *TCPServer) StartTLS(address string, cfg *config.TlsConfig) (<-chan net.Conn, error) {
	tlsConfig, err := utils.GetServerTLSConfig(cfg)
	if err != nil {
		logrus.Errorf("get client tls config [%+v] failed: %v", cfg, err)
		return nil, err
	}
	listener, err := tls.Listen("tcp", address, tlsConfig)
	if err != nil {
		logrus.Errorf("listen on %v failed: %v", address, err)
		return nil, err
	}

	svr.listener = listener
	return svr.startEntry()
}

func (svr *TCPServer) startEntry() (<-chan net.Conn, error) {
	listener := svr.listener
	if listener == nil {
		logrus.Error("no listener")
		return nil, errors.New("no listener")
	}
	svr.wg.Add(1)
	go func() {
		defer func() {
			svr.wg.Done()
			_ = listener.Close()
			svr.listener = nil
		}()
		for {
			conn, err := listener.Accept()
			if err != nil {
				logrus.Errorf("accept failed: %v", err)
				continue
			}
			svr.clientChan <- conn
		}
	}()
	return svr.clientChan, nil
}

func (svr *TCPServer) wait() {
	svr.wg.Wait()
}

func (svr *TCPServer) Wait() {
	svr.wait()
}

func (svr *TCPServer) TerminateAndWait() {
	listener := svr.listener
	if listener != nil {
		_ = listener.Close()
	}
	svr.wait()
}
