package presence

import "fmt"

func peersKey(roomUUID string) string {
	return fmt.Sprintf("sig:room:%s:peers", roomUUID)
}

func peerKey(roomUUID, userUUID string) string {
	return fmt.Sprintf("sig:peer:%s:%s", roomUUID, userUUID)
}

func pubsubChannel(roomUUID string) string {
	return fmt.Sprintf("sig:pubsub:room:%s", roomUUID)
}
