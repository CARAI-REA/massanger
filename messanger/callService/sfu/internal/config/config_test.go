package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	webrtcapi "sfu/internal/webrtc"
)

type stubEnv struct{ env string }

func (s stubEnv) Env() string { return s.env }

type stubHTTP struct {
	origins []string
	allowQT bool
}

func (s stubHTTP) Address() string                 { return ":8082" }
func (s stubHTTP) WSPath() string                  { return "/v1/ws" }
func (s stubHTTP) PingInterval() time.Duration     { return time.Second }
func (s stubHTTP) PongWait() time.Duration         { return time.Second }
func (s stubHTTP) MaxMessageBytes() int64          { return 1 }
func (s stubHTTP) MaxMessagesPerSec() int          { return 1 }
func (s stubHTTP) AllowedOrigins() []string        { return s.origins }
func (s stubHTTP) AllowQueryToken() bool           { return s.allowQT }

type stubRedis struct{ pass string }

func (s stubRedis) Host() string                      { return "localhost" }
func (s stubRedis) Port() string                      { return "6379" }
func (s stubRedis) Password() string                  { return s.pass }
func (s stubRedis) ConnectionTimeout() time.Duration  { return time.Second }
func (s stubRedis) MaxIdle() int                      { return 1 }
func (s stubRedis) IdleTimeout() time.Duration        { return time.Minute }
func (s stubRedis) AffinityTTL() time.Duration        { return time.Minute }

type stubJWT struct{ join, svc string }

func (s stubJWT) JoinTokenSecretKey() string    { return s.join }
func (s stubJWT) ServiceTokenSecretKey() string { return s.svc }
func (s stubJWT) ServiceName() string           { return "sfu" }

type stubInstance struct{ url string }

func (s stubInstance) ID() string          { return "sfu-1" }
func (s stubInstance) PublicWSURL() string { return s.url }

type stubWebRTC struct {
	nat  []string
	turn string
}

func (s stubWebRTC) ICEServers() []webrtcapi.ICEServer       { return nil }
func (s stubWebRTC) ClientICEServers() []webrtcapi.ICEServer { return nil }
func (s stubWebRTC) UDPPortMin() uint16                      { return 10000 }
func (s stubWebRTC) UDPPortMax() uint16                      { return 10100 }
func (s stubWebRTC) NAT1To1IPs() []string                    { return s.nat }
func (s stubWebRTC) NAT1To1CandidateType() string            { return "host" }
func (s stubWebRTC) TURNCredential() string                  { return s.turn }
func (s stubWebRTC) SimulcastEnabled() bool                  { return false }
func (s stubWebRTC) RecordingEnabled() bool                  { return false }
func (s stubWebRTC) RecordingDir() string                    { return "" }

func TestValidateProductionRejectsQueryToken(t *testing.T) {
	c := &config{
		AppEnv:   stubEnv{env: "production"},
		HTTP:     stubHTTP{origins: []string{"https://app.example"}, allowQT: true},
		Redis:    stubRedis{pass: "redis-password"},
		JWT:      stubJWT{join: "join-secret-long-enough-for-prod!!", svc: "svc-secret-long-enough-for-prod!!"},
		Instance: stubInstance{url: "wss://calls.example/sfu/v1/ws"},
		WebRTC:   stubWebRTC{nat: []string{"203.0.113.10"}, turn: "turn-credential-ok"},
	}
	err := c.ValidateProduction()
	require.Error(t, err)
	require.Contains(t, err.Error(), "WS_ALLOW_QUERY_TOKEN")
}

func TestValidateProductionOK(t *testing.T) {
	c := &config{
		AppEnv:   stubEnv{env: "production"},
		HTTP:     stubHTTP{origins: []string{"https://app.example"}, allowQT: false},
		Redis:    stubRedis{pass: "redis-password"},
		JWT:      stubJWT{join: "join-secret-long-enough-for-prod!!", svc: "svc-secret-long-enough-for-prod!!"},
		Instance: stubInstance{url: "wss://calls.example/sfu/v1/ws"},
		WebRTC:   stubWebRTC{nat: []string{"203.0.113.10"}, turn: "turn-credential-ok"},
	}
	require.NoError(t, c.ValidateProduction())
}
