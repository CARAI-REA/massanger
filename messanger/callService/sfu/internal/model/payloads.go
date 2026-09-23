package model

type ICEServer struct {
	URLs       []string `json:"urls"`
	Username   string   `json:"username,omitempty"`
	Credential string   `json:"credential,omitempty"`
}

type WelcomePayload struct {
	SelfUserUUID string      `json:"self_user_uuid"`
	Peers        []string    `json:"peers"`
	ICEServers   []ICEServer `json:"ice_servers"`
	Role         string      `json:"role"`
}

type PeerEventPayload struct {
	UserUUID string `json:"user_uuid"`
}

type SDPPayload struct {
	SDP string `json:"sdp"`
}

type ICEPayload struct {
	Candidate     string  `json:"candidate"`
	SDPMID        string  `json:"sdp_mid,omitempty"`
	SDPMLineIndex *uint16 `json:"sdp_m_line_index,omitempty"`
}

type ErrorPayload struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	InstanceURL string `json:"instance_url,omitempty"`
}
