package affinity

import "fmt"

func ownerKey(roomUUID string) string {
	return fmt.Sprintf("sfu:room:%s:owner", roomUUID)
}

func encodeOwner(instanceID, publicURL string) string {
	return instanceID + "|" + publicURL
}

func decodeOwner(v string) (instanceID, publicURL string) {
	for i := 0; i < len(v); i++ {
		if v[i] == '|' {
			return v[:i], v[i+1:]
		}
	}
	return v, ""
}
