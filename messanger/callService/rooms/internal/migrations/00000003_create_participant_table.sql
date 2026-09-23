-- +goose Up
CREATE TABLE participants (
    participant_uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),    -- UUID "входа" участника
    room_uuid        UUID NOT NULL REFERENCES rooms(room_uuid) ON DELETE CASCADE,
    user_uuid        UUID NOT NULL,                                 -- пользователь
    joined_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),             -- время входа
    left_at          TIMESTAMPTZ                                     -- время выхода (NULL = ещё в комнате)
);

-- если нужно запрещать повторный вход с тем же joined_at
CREATE UNIQUE INDEX uniq_participants_room_user_joined
    ON participants (room_uuid, user_uuid, joined_at);

-- активные участники комнаты
CREATE INDEX idx_participants_room_active
    ON participants (room_uuid)
    WHERE left_at IS NULL;

-- поиск по пользователю
CREATE INDEX idx_participants_user_uuid ON participants (user_uuid);

-- +goose Down
DROP TABLE IF EXISTS participants;
