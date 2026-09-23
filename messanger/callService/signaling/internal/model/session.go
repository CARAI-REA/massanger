package model

type Session struct {
	RoomUUID string
	UserUUID string
}

type PeerInfo struct {
	UserUUID string `json:"user_uuid"`
}
