package env_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"sfu/internal/config/env"
)

func TestWebRTCConfigTURNAndNAT(t *testing.T) {
	t.Setenv("WEBRTC_ICE_SERVERS", "stun:stun.l.google.com:19302,turn:coturn:3478?transport=udp")
	t.Setenv("WEBRTC_CLIENT_ICE_SERVERS", "stun:stun.l.google.com:19302,turn:localhost:3478?transport=udp")
	t.Setenv("WEBRTC_TURN_USERNAME", "sfu")
	t.Setenv("WEBRTC_TURN_CREDENTIAL", "secret")
	t.Setenv("WEBRTC_NAT_1TO1_IPS", "203.0.113.10,203.0.113.11")
	t.Setenv("WEBRTC_NAT_1TO1_CANDIDATE_TYPE", "host")
	t.Setenv("WEBRTC_UDP_PORT_MIN", "10000")
	t.Setenv("WEBRTC_UDP_PORT_MAX", "10100")

	cfg, err := env.NewWebRTCConfig()
	require.NoError(t, err)

	servers := cfg.ICEServers()
	require.Len(t, servers, 2)
	require.Equal(t, []string{"stun:stun.l.google.com:19302"}, servers[0].URLs)
	require.Equal(t, []string{"turn:coturn:3478?transport=udp"}, servers[1].URLs)
	require.Equal(t, "sfu", servers[1].Username)
	require.Equal(t, "secret", servers[1].Credential)

	client := cfg.ClientICEServers()
	require.Equal(t, []string{"turn:localhost:3478?transport=udp"}, client[1].URLs)
	require.Equal(t, []string{"203.0.113.10", "203.0.113.11"}, cfg.NAT1To1IPs())
	require.Equal(t, "host", cfg.NAT1To1CandidateType())
}
