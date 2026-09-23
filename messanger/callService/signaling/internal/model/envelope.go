package model

import "encoding/json"

type MessageType string

const (
	TypeWelcome      MessageType = "welcome"
	TypePeerJoined   MessageType = "peer_joined"
	TypePeerLeft     MessageType = "peer_left"
	TypeOffer        MessageType = "offer"
	TypeAnswer       MessageType = "answer"
	TypeICECandidate MessageType = "ice_candidate"
	TypePing         MessageType = "ping"
	TypePong         MessageType = "pong"
	TypeBye          MessageType = "bye"
	TypeError        MessageType = "error"
)

type Envelope struct {
	Type         MessageType     `json:"type"`
	RoomUUID     string          `json:"room_uuid"`
	FromUserUUID string          `json:"from_user_uuid"`
	ToUserUUID   string          `json:"to_user_uuid"`
	Payload      json.RawMessage `json:"payload"`
	TS           string          `json:"ts"`
}
