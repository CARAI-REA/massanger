package model

type SFUHint struct {
	Enabled bool   `json:"enabled"`
	URL     string `json:"url"`
}

type WelcomePayload struct {
	SelfUserUUID string  `json:"self_user_uuid"`
	Peers        []string `json:"peers"`
	SFU          SFUHint `json:"sfu"`
}

type PeerEventPayload struct {
	UserUUID string `json:"user_uuid"`
}

type SDPPayload struct {
	SDP string `json:"sdp"`
}

type ICEPayload struct {
	Candidate     string `json:"candidate"`
	SDPMID        string `json:"sdp_mid,omitempty"`
	SDPMLineIndex *int   `json:"sdp_m_line_index,omitempty"`
}

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ByePayload struct {
	Reason string `json:"reason,omitempty"`
}
