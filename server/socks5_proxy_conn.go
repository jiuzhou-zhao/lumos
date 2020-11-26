package server

import (
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/jiuzhou-zhao/lumos.git/server/socks5"
	"github.com/sirupsen/logrus"
)

type Socks5ProxyConn struct {
	conn  net.Conn
	proxy *Socks5Proxy
}

func NewSocks5ProxyConn(conn net.Conn, proxy *Socks5Proxy) *Socks5ProxyConn {
	return &Socks5ProxyConn{
		conn:  conn,
		proxy: proxy,
	}
}

func (conn *Socks5ProxyConn) Serve() {
	defer func() {
		_ = conn.conn.Close()
	}()

	method, err := conn.selectAuthMethod()
	if err != nil {
		logrus.Errorf("selectAuthMethod failed: %v", err)
		return
	}

	err = conn.checkAuthMethod(method)
	if err != nil {
		logrus.Errorf("checkAuthMethod failed: %v", err)
		return
	}

	req, err := conn.readRequest()
	if err != nil {
		logrus.Errorf("readRequest failed: %v", err)
		return
	}

	err = conn.handleRequest(req)
	if err != nil {
		logrus.Errorf("handleRequest failed: %v", err)
		return
	}
}

func (conn *Socks5ProxyConn) selectAuthMethod() (byte, error) {
	req, err := socks5.NewMethodSelectReqFrom(conn.conn)
	if err != nil {
		return 0, fmt.Errorf("NewMethodSelectReqFrom:%w", err)
	}

	if req.Ver != socks5.VerSocks5 {
		return 0, fmt.Errorf("invalid socks5 version: %v", req.Ver)
	}

	var method byte = socks5.MethodNone
	if conn.proxy.NeedAuth() {
		method = socks5.MethodUserPass
	}

	var exist bool
	for _, v := range req.Methods {
		if method == v {
			exist = true
			break
		}
	}

	if !exist {
		method = socks5.MethodNoAcceptable
	}

	_, err = conn.conn.Write(socks5.NewMethodSelectReply(method).ToBytes())
	if err != nil {
		return 0, fmt.Errorf("reply: %w", err)
	}

	if method == socks5.MethodNoAcceptable {
		return 0, fmt.Errorf("unsupported method: %v", req.Methods)
	}

	return method, nil
}

func (conn *Socks5ProxyConn) checkAuthMethod(method byte) error {
	switch method {
	case socks5.MethodNone:
		return nil
	case socks5.MethodUserPass:
		req, err := socks5.NewUserPassAuthReqFrom(conn.conn)
		if err != nil {
			return fmt.Errorf("NewUserPassAuthReqFrom:%w", err)
		}

		if req.Ver != socks5.VerAuthUserPass {
			return fmt.Errorf("ver not VerAuthUserPass: %v", req.Ver)
		}

		var status byte = socks5.AuthStatusFailure
		if conn.proxy.ValidateCredential(string(req.UserName), string(req.Password)) {
			status = socks5.AuthStatusSuccess
		}

		_, err = conn.conn.Write(socks5.NewUserPassAuthReply(status).ToBytes())
		if err != nil {
			return fmt.Errorf("reply: %w", err)
		}

		if status != socks5.AuthStatusSuccess {
			return fmt.Errorf("auth failed: %v", status)
		}
		return nil
	default:
		return fmt.Errorf("unsupported auth method: %v", method)
	}
}

func (conn *Socks5ProxyConn) readRequest() (*socks5.Request, error) {
	req, err := socks5.NewRequestFrom(conn.conn)
	if err != nil {
		return nil, fmt.Errorf("readRequest: %w", err)
	}
	return req, nil
}

func (conn *Socks5ProxyConn) handleRequest(req *socks5.Request) error {
	switch req.Cmd {
	case socks5.CmdConnect:
		return conn.handleConnect(req)
	case socks5.CmdUDP:
		logrus.Errorf("do not support udp socks5 proxy, who use it !?")
		fallthrough
	default:
		_, _ = conn.conn.Write(socks5.NewReply(socks5.RepCmdNotSupported, nil).ToBytes())
		return socks5.ErrCmdNotSupport
	}
}

func (conn *Socks5ProxyConn) handleConnect(req *socks5.Request) error {
	addr := req.Address()
	logrus.Debugf("tcp req: %v", addr)
	s, err := net.DialTimeout("tcp", addr, conn.proxy.cfg.DialTimeout)
	if err != nil {
		msg := err.Error()
		var rep byte = socks5.RepHostUnreachable
		if strings.Contains(msg, "refused") {
			rep = socks5.RepConnectionRefused
		} else if strings.Contains(msg, "network is unreachable") {
			rep = socks5.RepNetworkUnreachable
		}
		_, _ = conn.conn.Write(socks5.NewReply(rep, nil).ToBytes())
		return fmt.Errorf("connect to %v failed: %w", req.Address(), err)
	}
	defer func() {
		_ = s.Close()
	}()

	bAddr, err := socks5.NewAddrByteFromString(s.LocalAddr().(*net.TCPAddr).String())
	if err != nil {
		_, _ = conn.conn.Write(socks5.NewReply(socks5.RepServerFailure, nil).ToBytes())
		return fmt.Errorf("NewAddrByteFromString:%w", err)
	}

	_, err = conn.conn.Write(socks5.NewReply(socks5.RepSuccess, bAddr).ToBytes())
	if err != nil {
		return fmt.Errorf("reply:%w", err)
	}

	go func() {
		_, _ = io.Copy(conn.conn, s)
	}()

	_, _ = io.Copy(s, conn.conn)
	return nil
}
