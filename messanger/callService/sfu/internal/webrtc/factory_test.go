package webrtcapi_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	webrtcapi "sfu/internal/webrtc"
)

func TestFactoryWithTURNCredentials(t *testing.T) {
	f, err := webrtcapi.NewFactory(webrtcapi.Config{
		ICEServers: []webrtcapi.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
			{URLs: []string{"turn:turn.example:3478"}, Username: "u", Credential: "p"},
		},
		NAT1To1IPs:           []string{"198.51.100.1"},
		NAT1To1CandidateType: "srflx",
	})
	require.NoError(t, err)
	servers := f.ICEServers()
	require.Len(t, servers, 2)
	require.Equal(t, "u", servers[1].Username)
	require.Equal(t, "p", servers[1].Credential)
}
