package config

import (
	"path"
	"strings"
	"time"
)

type ProxyMode int

const (
	ModeProxy ProxyMode = 0
	ModeLocal ProxyMode = 1
	ModeRelay ProxyMode = 2
)

func (mode ProxyMode) Parse(s string) ProxyMode {
	s = strings.ToLower(s)
	switch s {
	case "local":
		return ModeLocal
	case "relay":
		return ModeRelay
	default:
		return ModeProxy
	}
}

type TlsConfig struct {
	Cert    string
	Key     string
	RootCAs []string
}

type Config struct {
	Mode               string
	EffectMode         ProxyMode
	ProxyAddress       string
	Credentials        []string
	RemoteProxyAddress string
	Secure             struct {
		EnableTLSClient bool
		EnableTLSServer bool
		Client          *TlsConfig
		Server          *TlsConfig
	}

	DialTimeout time.Duration
}

func (cfg *Config) Fix(cfgDir string) {
	cfg.EffectMode = ModeProxy.Parse(cfg.Mode)
	if cfg.DialTimeout <= 0 {
		cfg.DialTimeout = 30 * time.Second
	}

	cfg.fixTlsConfigFilePath(cfg.Secure.Client, cfgDir)
	cfg.fixTlsConfigFilePath(cfg.Secure.Server, cfgDir)
}

func (cfg *Config) fixTlsConfigFilePath(tlsConfig *TlsConfig, dir string) {
	if tlsConfig == nil {
		return
	}
	tlsConfig.Cert = cfg.fixFilePath(tlsConfig.Cert, dir)
	tlsConfig.Key = cfg.fixFilePath(tlsConfig.Key, dir)
	for idx := 0; idx < len(tlsConfig.RootCAs); idx++ {
		tlsConfig.RootCAs[idx] = cfg.fixFilePath(tlsConfig.RootCAs[idx], dir)
	}
}

func (cfg *Config) fixFilePath(file, dir string) string {
	if !strings.HasPrefix(file, "./") {
		return file
	}
	return path.Join(dir, file[2:])
}
