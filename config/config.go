package config

import (
	"log"
	"path"
	"strings"
	"time"
)

type ProxyMode int

const (
	ModeHTTPProxy   ProxyMode = 0
	ModeLocal       ProxyMode = 1
	ModeRelay       ProxyMode = 2
	ModeSocks5Proxy ProxyMode = 10
)

func ParseProxyMode(s string) ProxyMode {
	s = strings.ToLower(s)
	switch s {
	case "local":
		return ModeLocal
	case "relay":
		return ModeRelay
	case "socks5":
		return ModeSocks5Proxy
	default:
		return ModeHTTPProxy
	}
}

func (m ProxyMode) String() string {
	switch m {
	case ModeLocal:
		return "ModeLocal"
	case ModeRelay:
		return "ModeRelay"
	case ModeSocks5Proxy:
		return "ModeSocks5Proxy"
	default:
		return "ModeHTTPProxy"
	}
}

type TlsConfig struct {
	Cert    string
	Key     string
	RootCAs []string
}

type TLSEnableFlag struct {
	ConnectServerUseTLS bool
	ServerUseTLS        bool
}
type Config struct {
	Mode               string
	EffectMode         ProxyMode
	ProxyAddress       string
	Credentials        []string
	RemoteProxyAddress string
	Secure             struct {
		TLSEnableFlag          *TLSEnableFlag
		ConnectServerTLSConfig *TlsConfig
		ServerTLSConfig        *TlsConfig
	}

	DialTimeout time.Duration
	ExternalIP  string
}

func (cfg *Config) Fix(cfgDir string) {
	cfg.EffectMode = ParseProxyMode(cfg.Mode)
	if cfg.DialTimeout <= 0 {
		cfg.DialTimeout = 30 * time.Second
	}

	cfg.fixTlsConfigFilePath(cfg.Secure.ConnectServerTLSConfig, cfgDir)
	cfg.fixTlsConfigFilePath(cfg.Secure.ServerTLSConfig, cfgDir)

	if cfg.Secure.TLSEnableFlag == nil {
		switch cfg.EffectMode {
		case ModeLocal:
			cfg.Secure.TLSEnableFlag = &TLSEnableFlag{
				ServerUseTLS:        false,
				ConnectServerUseTLS: cfg.Secure.ConnectServerTLSConfig != nil,
			}
		case ModeRelay:
			cfg.Secure.TLSEnableFlag = &TLSEnableFlag{
				ServerUseTLS:        cfg.Secure.ServerTLSConfig != nil,
				ConnectServerUseTLS: cfg.Secure.ConnectServerTLSConfig != nil,
			}
		default:
			cfg.Secure.TLSEnableFlag = &TLSEnableFlag{
				ServerUseTLS:        cfg.Secure.ServerTLSConfig != nil,
				ConnectServerUseTLS: false,
			}
		}
	}

	switch cfg.EffectMode {
	case ModeLocal:
		if cfg.Secure.TLSEnableFlag.ServerUseTLS {
			cfg.Secure.TLSEnableFlag.ServerUseTLS = false
			log.Print("ModeLocal, ServerUseTLS should be false")
		}
	case ModeRelay:
	default:
		if cfg.Secure.TLSEnableFlag.ConnectServerUseTLS {
			cfg.Secure.TLSEnableFlag.ConnectServerUseTLS = false
			log.Printf("ModeProxy[%v], ConnectServerUseTLS should be false", cfg.EffectMode)
		}
	}

	if cfg.Secure.TLSEnableFlag.ServerUseTLS {
		if cfg.Secure.ServerTLSConfig == nil || cfg.Secure.ServerTLSConfig.Cert == "" || cfg.Secure.ServerTLSConfig.Key == "" {
			log.Fatal("ServerUseTLS config failed")
		}
	}

	if cfg.Secure.TLSEnableFlag.ConnectServerUseTLS {
		if cfg.Secure.ConnectServerTLSConfig == nil || cfg.Secure.ConnectServerTLSConfig.Cert == "" || cfg.Secure.ConnectServerTLSConfig.Key == "" {
			log.Fatal("ConnectServerUseTLS config failed")
		}
	}
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
