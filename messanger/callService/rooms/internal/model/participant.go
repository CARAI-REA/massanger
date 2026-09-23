package model

import "time"

// Метаданные участника (для удобства, но в proto храним как bytes)
type ParticipantMetadata struct {
	// отображаемое имя
	DisplayName string

	// информация о клиенте
	UserAgent string

	// версия клиента
	ClientVersion string

	// заглушен ли микрофон
	AudioMuted bool

	// выключена ли камера
	VideoMuted bool

	// роль: "participant", "moderator", "owner"
	Role string
}

// Информация об участнике
type Participant struct {
	// идентификатор участника
	UserUUID string

	// комната, в которой участвует
	RommUUID string

	// время входа
	JoinedAt *time.Time

	// время выхода (если вышел)
	LeftAt *time.Time

	// метаданные (имя, клиент, состояние)
	Metadata ParticipantMetadata
}

// Запрос на добавление участника в комнату
type AddParticipantRequest struct {
	// в какую комнату входит
	RoomUUID string

	// кто входит
	UserUUID string

	// дополнительная информация (имя, браузер и т.п.)
	Metadata ParticipantMetadata
}

// Ответ на запрос добавления участника в комнату
type AddParticipantResponse struct {
	// успешно ли добавлен участник
	Success bool

	// JWT для подключения к signaling сервису
	JoinToken string

	// информация о добавленном участнике
	Participant Participant
}

// Запрос на удаление участника из комнаты (выход или кик)
type RemoveParticipantRequest struct {
	// из какой комнаты выходит
	RoomUUID string

	// кто выходит
	UserUUID string

	// кто инициировал удаление (для кика)
	RemovedBy string
}

// Ответ на запрос удаления участника из комнаты
type RemoveParticipantResponse struct {
	// успешно ли удаление
	Success bool
}

// Запрос на получение списка участников комнаты
type GetParticipantsRequest struct {
	// комната
	RoomUUID string

	// только активные (не вышедшие)
	OnlyActive bool
}

// Ответ со списком участников комнаты и количеством активных
type GetParticipantsResponse struct {
	// список участников
	Partisipants []Participant

	// количество активных
	ActiveCount int32
}

// Запрос на проверку, активен ли участник в комнате
type IsParticipantRequest struct {
	// комната
	RoomUUID string

	// пользователь
	UserUUID string
}

// Ответ на запрос проверки активности участника в комнате
type IsParticipantResponse struct {
	// true если активен в комнате
	IsActive bool

	// информация об участнике (если активен)
	Participant Participant
}

// Запрос на обновление метаданных участника в комнате
type UpdateParticipantMetadataRequest struct {
	// комната
	RoomUUID string

	// пользователь
	UserUUID string

	// новые метаданные в JSON
	Metadata ParticipantMetadata
}

// Ответ на запрос обновления метаданных участника
type UpdateParticipantMetadataResponse struct {
	// успешно ли обновление
	Seccess bool

	// обновлённый участник
	Participant Participant
}

// AssertCanJoinRequest is used by signaling/sfu to verify join eligibility.
type AssertCanJoinRequest struct {
	RoomUUID string
	UserUUID string
}

// AssertCanJoinResponse reports whether the user may join the room.
type AssertCanJoinResponse struct {
	Ok         bool
	RoomStatus string
}

// RefreshJoinTokenRequest asks for a new join JWT for an active participant.
type RefreshJoinTokenRequest struct {
	RoomUUID string
	UserUUID string
}

// RefreshJoinTokenResponse contains a freshly issued join JWT.
type RefreshJoinTokenResponse struct {
	JoinToken string
}

// GetTURNCredentialsRequest requests short-lived TURN credentials for a participant.
type GetTURNCredentialsRequest struct {
	RoomUUID string
	UserUUID string
}

// ICEServer describes a STUN/TURN endpoint with optional credentials.
type ICEServer struct {
	URLs       []string
	Username   string
	Credential string
}

// GetTURNCredentialsResponse returns ICE servers with ephemeral TURN credentials.
type GetTURNCredentialsResponse struct {
	ICEServers []ICEServer
	TTLSeconds int64
}
