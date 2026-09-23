-- +goose Up
CREATE TABLE room_settings (
    room_uuid         UUID PRIMARY KEY REFERENCES rooms(room_uuid) ON DELETE CASCADE,
    max_participants  INT NOT NULL CHECK (max_participants > 0),    -- максимальное количество участников
    recording_enabled BOOLEAN NOT NULL DEFAULT FALSE,               -- включена ли запись
    quality           VARCHAR(20) NOT NULL,                         -- "low", "medium", "high" / "HD","SD"
    auto_close        BOOLEAN NOT NULL DEFAULT FALSE,               -- авто-закрытие при выходе всех
    allowed_users     UUID[]                                        -- массив user_uuid, допущенных в комнату (пустой = публичная)

);

-- +goose Down
DROP TABLE IF EXISTS room_settings;
