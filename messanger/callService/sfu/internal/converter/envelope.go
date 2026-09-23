package converter

import (
	"encoding/json"
	"fmt"
	"time"

	"sfu/internal/model"
)

func MarshalEnvelope(msg model.Envelope) ([]byte, error) {
	if msg.TS == "" {
		msg.TS = time.Now().UTC().Format(time.RFC3339Nano)
	}
	return json.Marshal(msg)
}

func UnmarshalEnvelope(data []byte) (model.Envelope, error) {
	var msg model.Envelope
	if err := json.Unmarshal(data, &msg); err != nil {
		return model.Envelope{}, fmt.Errorf("invalid envelope: %w", err)
	}
	if msg.Type == "" {
		return model.Envelope{}, fmt.Errorf("missing type")
	}
	return msg, nil
}

func MustPayload(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return b
}

func ErrorEnvelope(roomUUID, code, message string) model.Envelope {
	return model.Envelope{
		Type:     model.TypeError,
		RoomUUID: roomUUID,
		Payload: MustPayload(model.ErrorPayload{
			Code:    code,
			Message: message,
		}),
		TS: time.Now().UTC().Format(time.RFC3339Nano),
	}
}

func RedirectEnvelope(roomUUID, instanceURL string) model.Envelope {
	return model.Envelope{
		Type:     model.TypeError,
		RoomUUID: roomUUID,
		Payload: MustPayload(model.ErrorPayload{
			Code:        model.CodeRoomOnOtherInstance,
			Message:     "room hosted on another SFU instance",
			InstanceURL: instanceURL,
		}),
		TS: time.Now().UTC().Format(time.RFC3339Nano),
	}
}
