package server

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
)

type HTTPProxyConn struct {
	proxy *HTTPProxy

	conn   net.Conn
	reader *bufio.Reader
}

func NewHTTPProxyConn(conn net.Conn, proxy *HTTPProxy) *HTTPProxyConn {
	return &HTTPProxyConn{
		conn:   conn,
		reader: bufio.NewReader(conn),
		proxy:  proxy,
	}
}

func (conn *HTTPProxyConn) Server() {
	defer func() {
		_ = conn.conn.Close()
	}()

	rawHttpRequestHeader, remote, credential, proxyFlag, err := conn.getTunnelInfo()
	if err != nil {
		logrus.Errorf("parse tunnel info failed: %v\n", err)
		return
	}

	if !conn.auth(credential) {
		logrus.Errorf("auth failed: " + credential)
		return
	}

	logrus.Infof("connecting to %v\n", remote)
	remoteConn, err := net.Dial("tcp", remote)
	if err != nil {
		logrus.Errorf("dial %v failed: %v\n", remote, err)
		return
	}

	if proxyFlag {
		// if https, should sent 200 to client
		_, err = conn.conn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
		if err != nil {
			logrus.Errorf("write failed: %v\n", err)
			return
		}
	} else {
		// if not https, should sent the request header to remote
		_, err = rawHttpRequestHeader.WriteTo(remoteConn)
		if err != nil {
			logrus.Errorf("write failed: %v\n", err)
			return
		}
	}

	// build bidirectional-streams
	logrus.Infof("begin tunnel %v <-> %v\n", conn.conn.RemoteAddr(), remote)
	conn.tunnel(remoteConn)
	logrus.Infof("stop tunnel %v <-> %v\n", conn.conn.RemoteAddr(), remote)
}

func (conn *HTTPProxyConn) tunnel(remoteConn net.Conn) {
	go func() {
		defer func() {
			_ = remoteConn.Close()
		}()
		_, err := conn.reader.WriteTo(remoteConn)
		if err != nil {
			logrus.Warningf("write failed: %v\n", err)
		}
	}()
	_, err := io.Copy(conn.conn, remoteConn)
	if err != nil {
		logrus.Warningf("write failed: %v\n", err)
	}
}

func (conn *HTTPProxyConn) getTunnelInfo() (rawReqHeader bytes.Buffer, host, credential string, proxyFlag bool, err error) {
	tp := textproto.NewReader(conn.reader)

	// CONNECT /index.html HTTP/1.0
	var requestLine string
	if requestLine, err = tp.ReadLine(); err != nil {
		return
	}

	method, requestURI, _, ok := parseRequestLine(requestLine)
	if !ok {
		err = errors.New("malformed HTTP request")
		return
	}

	if method == "CONNECT" {
		proxyFlag = true
		requestURI = "http://" + requestURI
	}

	uriInfo, err := url.ParseRequestURI(requestURI)
	if err != nil {
		logrus.Errorf("parse request uri failed: %v, %v", err, requestURI)
		return
	}

	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil {
		logrus.Errorf("read mime header failed: %v\n", err)
		return
	}

	credential = mimeHeader.Get("Proxy-Authorization")

	if uriInfo.Host == "" {
		host = mimeHeader.Get("Host")
	} else {
		if !strings.Contains(uriInfo.Host, ":") {
			host = uriInfo.Host + ":80"
		} else {
			host = uriInfo.Host
		}
	}

	// rebuild http request header
	rawReqHeader.WriteString(requestLine + "\r\n")
	for k, vs := range mimeHeader {
		for _, v := range vs {
			rawReqHeader.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
		}
	}
	rawReqHeader.WriteString("\r\n")
	return
}

func (conn *HTTPProxyConn) auth(credential string) bool {
	if !conn.proxy.NeedAuth() || conn.proxy.ValidateCredential(credential) {
		return true
	}
	// 407
	_, err := conn.conn.Write(
		[]byte("HTTP/1.1 407 Proxy Authentication Required\r\nProxy-Authenticate: Basic realm=\"*\"\r\n\r\n"))
	if err != nil {
		logrus.Errorf("write failed: %v\n", err)
	}
	return false
}

func parseRequestLine(line string) (method, requestURI, proto string, ok bool) {
	s1 := strings.Index(line, " ")
	s2 := strings.Index(line[s1+1:], " ")
	if s1 < 0 || s2 < 0 {
		logrus.Errorf("unknown first line: %v\n", line)
		return
	}
	s2 += s1 + 1
	return line[:s1], line[s1+1 : s2], line[s2+1:], true
}
