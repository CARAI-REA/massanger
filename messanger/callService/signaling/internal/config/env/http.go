package env

import (
	"net"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
)

type httpEnvConfig struct {
	Host              string        `env:"HTTP_HOST" envDefault:"0.0.0.0"`
	Port              string        `env:"HTTP_PORT,required"`
	WSPath            string        `env:"WS_PATH" envDefault:"/v1/ws"`
	PingInterval      time.Duration `env:"WS_PING_INTERVAL" envDefault:"20s"`
	PongWait          time.Duration `env:"WS_PONG_WAIT" envDefault:"30s"`
	MaxMessageBytes   int64         `env:"WS_MAX_MESSAGE_BYTES" envDefault:"65536"`
	MaxMessagesPerSec int           `env:"WS_MAX_MESSAGES_PER_SEC" envDefault:"50"`
	AllowedOrigins    string        `env:"WS_ALLOWED_ORIGINS" envDefault:""`
	AllowQueryToken   bool          `env:"WS_ALLOW_QUERY_TOKEN" envDefault:"true"`
}

type httpConfig struct {
	raw httpEnvConfig
}

func NewHTTPConfig() (*httpConfig, error) {
	var raw httpEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}
	return &httpConfig{raw: raw}, nil
}

func (cfg *httpConfig) Address() string {
	return net.JoinHostPort(cfg.raw.Host, cfg.raw.Port)
}

func (cfg *httpConfig) WSPath() string              { return cfg.raw.WSPath }
func (cfg *httpConfig) PingInterval() time.Duration { return cfg.raw.PingInterval }
func (cfg *httpConfig) PongWait() time.Duration     { return cfg.raw.PongWait }
func (cfg *httpConfig) MaxMessageBytes() int64      { return cfg.raw.MaxMessageBytes }
func (cfg *httpConfig) MaxMessagesPerSec() int      { return cfg.raw.MaxMessagesPerSec }
func (cfg *httpConfig) AllowQueryToken() bool       { return cfg.raw.AllowQueryToken }

func (cfg *httpConfig) AllowedOrigins() []string {
	if cfg.raw.AllowedOrigins == "" {
		return nil
	}
	parts := strings.Split(cfg.raw.AllowedOrigins, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
