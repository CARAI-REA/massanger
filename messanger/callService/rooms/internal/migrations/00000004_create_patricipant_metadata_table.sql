-- +goose Up
CREATE TABLE participant_metadata (
    participant_uuid UUID PRIMARY KEY
        REFERENCES participants(participant_uuid) ON DELETE CASCADE,

    display_name   VARCHAR(255),                -- отображаемое имя
    user_agent     TEXT,                        -- user-agent клиента
    client_version VARCHAR(50),                 -- версия клиента
    audio_muted    BOOLEAN NOT NULL DEFAULT FALSE, -- заглушен ли микрофон
    video_muted    BOOLEAN NOT NULL DEFAULT FALSE, -- выключена ли камера
    role           VARCHAR(32) NOT NULL DEFAULT 'participant' -- "participant", "moderator", "owner"
);

-- +goose Down
DROP TABLE IF EXISTS participant_metadata;
