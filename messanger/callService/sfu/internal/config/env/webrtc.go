package env

import (
	"strings"

	"github.com/caarlos0/env/v11"

	webrtcapi "sfu/internal/webrtc"
)

type webrtcEnvConfig struct {
	ICEServers           string `env:"WEBRTC_ICE_SERVERS" envDefault:"stun:stun.l.google.com:19302"`
	ClientICEServers     string `env:"WEBRTC_CLIENT_ICE_SERVERS"` // optional override for welcome payloads
	TURNUsername         string `env:"WEBRTC_TURN_USERNAME"`
	TURNCredential       string `env:"WEBRTC_TURN_CREDENTIAL"`
	UDPPortMin           uint16 `env:"WEBRTC_UDP_PORT_MIN" envDefault:"10000"`
	UDPPortMax           uint16 `env:"WEBRTC_UDP_PORT_MAX" envDefault:"10100"`
	NAT1To1IPs           string `env:"WEBRTC_NAT_1TO1_IPS"` // comma-separated public/host IPs for Docker NAT
	NAT1To1CandidateType string `env:"WEBRTC_NAT_1TO1_CANDIDATE_TYPE" envDefault:"host"`
	SimulcastEnabled     bool   `env:"WEBRTC_SIMULCAST_ENABLED" envDefault:"false"`
	RecordingEnabled     bool   `env:"WEBRTC_RECORDING_ENABLED" envDefault:"false"`
	RecordingDir         string `env:"RECORDING_DIR" envDefault:"./recordings"`
}

type webrtcConfig struct{ raw webrtcEnvConfig }

func NewWebRTCConfig() (*webrtcConfig, error) {
	var raw webrtcEnvConfig
	if err := env.Parse(&raw); err != nil {
		return nil, err
	}
	if raw.UDPPortMax < raw.UDPPortMin {
		raw.UDPPortMax = raw.UDPPortMin
	}
	return &webrtcConfig{raw: raw}, nil
}

func (c *webrtcConfig) ICEServers() []webrtcapi.ICEServer {
	return buildICEServers(c.raw.ICEServers, c.raw.TURNUsername, c.raw.TURNCredential)
}

func (c *webrtcConfig) ClientICEServers() []webrtcapi.ICEServer {
	if strings.TrimSpace(c.raw.ClientICEServers) == "" {
		return c.ICEServers()
	}
	return buildICEServers(c.raw.ClientICEServers, c.raw.TURNUsername, c.raw.TURNCredential)
}

func (c *webrtcConfig) UDPPortMin() uint16 { return c.raw.UDPPortMin }
func (c *webrtcConfig) UDPPortMax() uint16 { return c.raw.UDPPortMax }

func (c *webrtcConfig) NAT1To1IPs() []string {
	return splitCSV(c.raw.NAT1To1IPs)
}

func (c *webrtcConfig) NAT1To1CandidateType() string {
	return strings.TrimSpace(c.raw.NAT1To1CandidateType)
}

func (c *webrtcConfig) TURNCredential() string  { return c.raw.TURNCredential }
func (c *webrtcConfig) SimulcastEnabled() bool   { return c.raw.SimulcastEnabled }
func (c *webrtcConfig) RecordingEnabled() bool   { return c.raw.RecordingEnabled }
func (c *webrtcConfig) RecordingDir() string     { return c.raw.RecordingDir }

func buildICEServers(raw, turnUser, turnCred string) []webrtcapi.ICEServer {
	urls := splitCSV(raw)
	if len(urls) == 0 {
		return nil
	}

	var stunURLs, turnURLs []string
	for _, u := range urls {
		lower := strings.ToLower(u)
		if strings.HasPrefix(lower, "turn:") || strings.HasPrefix(lower, "turns:") {
			turnURLs = append(turnURLs, u)
			continue
		}
		stunURLs = append(stunURLs, u)
	}

	out := make([]webrtcapi.ICEServer, 0, 2)
	if len(stunURLs) > 0 {
		out = append(out, webrtcapi.ICEServer{URLs: stunURLs})
	}
	if len(turnURLs) > 0 {
		out = append(out, webrtcapi.ICEServer{
			URLs:       turnURLs,
			Username:   turnUser,
			Credential: turnCred,
		})
	}
	return out
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
