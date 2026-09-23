package webrtcapi

import (
	"fmt"
	"strings"

	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v4"
)

// ICEServer is a STUN/TURN entry shared by PeerConnections and welcome payloads.
type ICEServer struct {
	URLs       []string
	Username   string
	Credential string
}

// Config configures the Pion API used for all SFU PeerConnections.
type Config struct {
	ICEServers           []ICEServer
	UDPPortMin           uint16
	UDPPortMax           uint16
	NAT1To1IPs           []string
	NAT1To1CandidateType string // "host" (default) or "srflx"
}

type Factory struct {
	api        *webrtc.API
	iceServers []webrtc.ICEServer
}

func NewFactory(cfg Config) (*Factory, error) {
	se := webrtc.SettingEngine{}
	if cfg.UDPPortMin > 0 && cfg.UDPPortMax >= cfg.UDPPortMin {
		if err := se.SetEphemeralUDPPortRange(cfg.UDPPortMin, cfg.UDPPortMax); err != nil {
			return nil, err
		}
	}
	if len(cfg.NAT1To1IPs) > 0 {
		candType, err := parseNATCandidateType(cfg.NAT1To1CandidateType)
		if err != nil {
			return nil, err
		}
		se.SetNAT1To1IPs(cfg.NAT1To1IPs, candType)
		// Host 1:1 replaces private/container IPs; keep loopback usable for local tests.
		if candType == webrtc.ICECandidateTypeHost {
			se.SetIncludeLoopbackCandidate(true)
		}
	}

	m := &webrtc.MediaEngine{}
	if err := m.RegisterDefaultCodecs(); err != nil {
		return nil, err
	}
	i := &interceptor.Registry{}
	if err := webrtc.RegisterDefaultInterceptors(m, i); err != nil {
		return nil, err
	}
	api := webrtc.NewAPI(
		webrtc.WithSettingEngine(se),
		webrtc.WithMediaEngine(m),
		webrtc.WithInterceptorRegistry(i),
	)

	servers := toPionICEServers(cfg.ICEServers)
	return &Factory{api: api, iceServers: servers}, nil
}

func (f *Factory) NewPeerConnection() (*webrtc.PeerConnection, error) {
	return f.api.NewPeerConnection(webrtc.Configuration{
		ICEServers: f.iceServers,
	})
}

func (f *Factory) ICEServers() []ICEServer {
	out := make([]ICEServer, 0, len(f.iceServers))
	for _, s := range f.iceServers {
		cred, _ := s.Credential.(string)
		out = append(out, ICEServer{
			URLs:       append([]string(nil), s.URLs...),
			Username:   s.Username,
			Credential: cred,
		})
	}
	return out
}

func toPionICEServers(in []ICEServer) []webrtc.ICEServer {
	out := make([]webrtc.ICEServer, 0, len(in))
	for _, s := range in {
		if len(s.URLs) == 0 {
			continue
		}
		out = append(out, webrtc.ICEServer{
			URLs:       append([]string(nil), s.URLs...),
			Username:   s.Username,
			Credential: s.Credential,
		})
	}
	return out
}

func parseNATCandidateType(raw string) (webrtc.ICECandidateType, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "host":
		return webrtc.ICECandidateTypeHost, nil
	case "srflx":
		return webrtc.ICECandidateTypeSrflx, nil
	default:
		return webrtc.ICECandidateTypeHost, fmt.Errorf("invalid WEBRTC_NAT_1TO1_CANDIDATE_TYPE %q (want host|srflx)", raw)
	}
}
