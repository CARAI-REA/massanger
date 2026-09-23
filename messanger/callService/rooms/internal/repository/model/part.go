package model

import "time"

type Room struct {
	// уникальный идентификатор комнаты (UUID)
	RoomUUID string

	// название комнаты
	Name string

	// кто создал комнату
	OwnerUUID string

	// время создания
	CreatedAt *time.Time

	// время последнего обновления
	UpdatedAt *time.Time

	// время удаления (если есть)
	DeletedAt *time.Time

	// настройки комнаты (метаданные)
	RoomSettings RoomSettings

	// статуc
	Status string
}

// Настройки комнаты
type RoomSettings struct {
	// максимальное количество участников
	MaxParticipants int32

	// список пользователей, допущенных в приватную комнату (пустой = публичная)
	AllowedUsers []string

	// качество видеj
	Quality string

	// автоматически завершать комнату, когда все вышли
	AutoClose bool

	// включить запись звонка (SFU egress)
	RecordingEnabled bool
}

// Запрос на создание комнаты
type CreateRoomRequest struct {
	// название комнаты
	Name string

	// идентификатор создателя (из JWT)
	OwnerUUID string

	// настройки комнаты
	RoomSettings RoomSettings
}

// Ответ о создании комнаты, включающий в себя структуру room и токен
type CreateRoomResponse struct {
	// созданная комната со всеми полями
	Room Room

	// JWT для немедленного подключения к signaling
	JoinToken string
}

// Запрос на получение UUID комнаты
type GetRoomRequest struct {
	// идентификатор комнаты
	RoomUUID string
}

// Ответ с информацией о комнате в структуре room
type GetRoomResponse struct {
	// информация о комнате
	Room Room
}

// Запрос на обновление комнаты
type UpdateRoomRequest struct {
	// идентификатор комнаты
	RoomUUID string

	// идентификатор владельца (для проверки прав)
	OwnerUUID string

	// новые настройки
	RoomSettings RoomSettings
}

// Ответ на запрос обновления комнаты с флагом успеха и новой структурой комнаты
type UpdateRoomResponse struct {
	// успешно ли обновление
	Success bool

	// обновлённая комната
	Room Room
}

// Запрос на удаление комнаты
type DeleteRoomRequest struct {
	// идентификатор комнаты
	RoomUUID string

	// идентификатор владельца (для проверки прав)
	OwnerUUID string

	// true = полное удаление, false = soft delete (мягкое удаление, это когда мы не удаляем а просто ставим время удаления, что бы могли обратиться к комнате в будущем)
	Permanent bool
}

// Ответ на запрос удаления комнаты
// Ответ на запрос о удалении комнаты содержащий флаг успеха
type DeleteRoomResponse struct {
	// успешно ли удаление
	Success bool
}

//list rooms нужен для того что бы вернуть список комнат созданных пользователем
//
// Запрос на получение списка комнат пользователя с поддержкой фильтрации и пагинации
type ListRoomRequest struct {
	// чьи комнаты ищем
	OwnerUUID string

	// фильтр по статусу: "active", "ended", "all"
	Status string

	// количество записей (пагинация)
	Limit int32

	// смещение (пагинация)
	Offset int32
}

// Ответ со списком комнат и общим количеством для пагинации
type ListRoomResponse struct {
	// список комнат
	Rooms []Room

	// общее количество (для пагинации)
	Total int32
}

// Запрос на принудительное завершение активной комнаты
type EndRoomRequest struct {
	// идентификатор комнаты
	RoomUUID string

	// идентификатор владельца или модератора
	OwnerUUID string
}

// Ответ на запрос завершения комнаты
type EndRoomResponse struct {
	// успешно ли завершение
	Success bool
}
