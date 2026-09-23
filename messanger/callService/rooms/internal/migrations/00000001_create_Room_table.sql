-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE rooms (
    room_uuid   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,                  -- название комнаты
    owner_uuid  UUID NOT NULL,                          -- кто создал комнату
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),      -- время создания
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),      -- время последнего обновления
    deleted_at  TIMESTAMPTZ,                             -- время удаления (если есть)
    status      VARCHAR(20) NOT NULL DEFAULT 'active'    -- 'active', 'ended'
);

CREATE INDEX idx_rooms_owner_uuid ON rooms (owner_uuid);
CREATE INDEX idx_rooms_status     ON rooms (status);

-- +goose Down
DROP TABLE IF EXISTS rooms;
