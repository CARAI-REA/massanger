# Техническое задание на разработку микросервиса Room Service

---

## Оглавление

1. [Введение](#1-введение)
2. [Общие сведения о системе](#2-общие-сведения-о-системе)
3. [Роль Room Service](#3-роль-room-service)
4. [Функциональные требования](#4-функциональные-требования)
5. [Нефункциональные требования](#5-нефункциональные-требования)
6. [Архитектура и технологии](#6-архитектура-и-технологии)
7. [Детали реализации](#7-детали-реализации)
8. [Тестирование](#8-тестирование)
9. [Развертывание](#9-развертывание)
10. [Мониторинг и логирование](#10-мониторинг-и-логирование)
11. [Дорожная карта разработки](#11-дорожная-карта-разработки)
12. [Приложение](#12-приложение)
13. [Детальная спецификация структур данных и методов](#13-детальная-спецификация-структур-данных-и-методов)

---

## 1. Введение

Настоящий документ описывает требования к разработке микросервиса **Room Service**, предназначенного для управления комнатами и участниками в системе групповых видеозвонков на базе WebRTC. Документ предназначен для разработчиков, тестировщиков и технических специалистов.

---

## 2. Общие сведения о системе

Room Service является частью распределённой системы микросервисов, включающей также:

| Сервис | Описание |
|--------|----------|
| **API Gateway** | Внешний шлюз, принимающий запросы от клиентов |
| **Signaling Service** | Сервис для обмена сигналами WebRTC |
| **Media Server (SFU)** | Сервис ретрансляции медиапотоков на базе Pion WebRTC |
| **Kafka** | Брокер сообщений для асинхронных событий |

Все внутренние взаимодействия между сервисами выполняются по **gRPC**.

---

## 3. Роль Room Service

Room Service является **источником истины** для данных о комнатах и участниках. Он отвечает за:

- создание, редактирование, удаление комнат;
- учёт участников (вход/выход);
- проверку прав доступа;
- хранение метаданных комнат и истории событий;
- публикацию событий для других сервисов.

---

## 4. Функциональные требования

### 4.1 Управление комнатами

| ID | Функция | Описание |
|----|---------|----------|
| R1 | Создание комнаты | Пользователь (владелец) создаёт новую комнату. Указывает название, настройки (макс. участников, тип доступа, качество, запись). |
| R2 | Получение информации о комнате | Возвращает данные комнаты по её идентификатору (название, владелец, настройки, статус). |
| R3 | Обновление настроек комнаты | Владелец может изменить настройки комнаты (кроме идентификатора). |
| R4 | Удаление комнаты | Помечает комнату как удалённую (soft delete) или полностью удаляет данные (hard delete – опционально). |
| R5 | Поиск комнат | Поиск по названию, владельцу, статусу (активные/завершённые). |
| R6 | Завершение комнаты | Принудительное завершение комнаты (например, когда последний участник вышел или истекло время). |

### 4.2 Управление участниками

| ID | Функция | Описание |
|----|---------|----------|
| P1 | Добавление участника | Регистрирует вход пользователя в комнату. Проверяет лимиты и права доступа. |
| P2 | Удаление участника | Фиксирует выход участника (добровольный или принудительный кик). |
| P3 | Получение списка участников | Возвращает список активных участников комнаты (с метаданными). |
| P4 | Проверка присутствия | Проверяет, находится ли пользователь в комнате. |
| P5 | Обновление метаданных участника | Например, информация о клиенте (браузер, версия, имя). |

### 4.3 Проверка прав и лимитов

| ID | Функция | Описание |
|----|---------|----------|
| A1 | Проверка доступа к комнате | Для приватных комнат проверяет, разрешён ли пользователю вход (по списку `allowed_users`). |
| A2 | Проверка лимита участников | Не даёт войти, если комната заполнена. |
| A3 | Проверка владельца | Только владелец может удалять/изменять комнату. |
| A4 | Генерация токена для входа | При успешном добавлении участника создаёт временный токен (JWT) для подключения к Signaling Service. |

### 4.4 Событийность

| ID | Функция | Описание |
|----|---------|----------|
| E1 | Публикация событий комнаты | При создании, изменении, удалении комнаты отправлять событие в Kafka (топик `room-events`). |
| E2 | Публикация событий участников | При входе/выходе участника отправлять событие в Kafka (топик `participant-events`). |

---

## 5. Нефункциональные требования

### 5.1 Производительность

- Время ответа на запрос (95-й перцентиль) **не более 100 мс** (без учёта задержек сети)
- Поддержка до **10 000** одновременных комнат
- Поддержка до **100 000** активных участников (записи в БД и кэше)
- Пропускная способность: **не менее 1000 запросов в секунду** на инстанс

### 5.2 Надёжность и отказоустойчивость

- Сервис должен быть **stateless** (вся state хранится в БД и Redis), допускается горизонтальное масштабирование
- При падении инстанса запросы должны обрабатываться другими инстансами
- Данные не должны теряться: подтверждение записи в БД до ответа клиенту
- Kafka-события должны быть доставлены **at least once**

### 5.3 Безопасность

- Все входящие запросы (gRPC) должны проходить **аутентификацию** (JWT, переданный из API Gateway)
- Неавторизованные запросы отклоняются с кодом `UNAUTHENTICATED`
- Проверка прав доступа к ресурсам (только владелец может изменять комнату)
- Конфиденциальные данные (токены) **не логируются**

### 5.4 Масштабируемость

- Возможность запуска нескольких экземпляров сервиса за балансировщиком
- Кэширование в Redis для снижения нагрузки на БД
- Асинхронная обработка через Kafka для фоновых задач

### 5.5 Наблюдаемость

- Логирование в структурированном формате (JSON) с уровнем `info`/`error`/`debug`
- Метрики в Prometheus: количество запросов, ошибки, время ответа, количество активных комнат/участников
- Трассировка (опционально OpenTelemetry) для отслеживания цепочек вызовов

---

## 6. Архитектура и технологии

### 6.1 Технологический стек

| Компонент | Технология |
|-----------|------------|
| Язык | Go (версия 1.21+) |
| База данных | PostgreSQL 14+ (основное хранилище) |
| Кэш | Redis 7+ (кэш активных комнат и Pub/Sub) |
| Брокер сообщений | Kafka (через segmentio/kafka-go или confluent-kafka-go) |
| Межсервисное взаимодействие | gRPC (protobuf) |
| Конфигурация | Переменные окружения + файлы (или consul/etcd опционально) |
| Контейнеризация | Docker, оркестрация Kubernetes (в перспективе) |

### 6.2 Взаимодействие с другими сервисами

*(Схема взаимодействия)*

### 6.3 Модель данных

#### Таблица `rooms`

```sql
CREATE TABLE rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    owner_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,  -- soft delete
    settings JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(20) NOT NULL DEFAULT 'active'  -- active, ended
);

CREATE INDEX idx_rooms_owner_id ON rooms(owner_id);
CREATE INDEX idx_rooms_status ON rooms(status);
```

#### Таблица `participants`

```sql
CREATE TABLE participants (
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    left_at TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}',
    PRIMARY KEY (room_id, user_id, joined_at)  -- позволяет несколько входов одного пользователя
);

CREATE INDEX idx_participants_room_id ON participants(room_id) WHERE left_at IS NULL;
CREATE INDEX idx_participants_user_id ON participants(user_id);
```

#### Кэш в Redis

| Ключ | Описание |
|------|----------|
| `room:{roomId}:info` | JSON с данными комнаты (срок жизни, например, 1 час) |
| `room:{roomId}:participants` | Set (user_id) активных участников |
| `room:{roomId}:metadata:{userId}` | Hash с метаданными участника |

### 6.4 Структура gRPC API

**Файл:** `proto/room_service.proto`

```protobuf
syntax = "proto3";

package room;

option go_package = "./proto;roompb";

service RoomService {
    // Управление комнатами
    rpc CreateRoom(CreateRoomRequest) returns (CreateRoomResponse);
    rpc GetRoom(GetRoomRequest) returns (GetRoomResponse);
    rpc UpdateRoom(UpdateRoomRequest) returns (UpdateRoomResponse);
    rpc DeleteRoom(DeleteRoomRequest) returns (DeleteRoomResponse);
    rpc ListRooms(ListRoomsRequest) returns (ListRoomsResponse);
    rpc EndRoom(EndRoomRequest) returns (EndRoomResponse);

    // Управление участниками
    rpc AddParticipant(AddParticipantRequest) returns (AddParticipantResponse);
    rpc RemoveParticipant(RemoveParticipantRequest) returns (RemoveParticipantResponse);
    rpc GetParticipants(GetParticipantsRequest) returns (GetParticipantsResponse);
    rpc IsParticipant(IsParticipantRequest) returns (IsParticipantResponse);
    rpc UpdateParticipantMetadata(UpdateParticipantMetadataRequest) returns (UpdateParticipantMetadataResponse);
}

// Сообщения для комнат
message Room {
    string id = 1;
    string name = 2;
    string owner_id = 3;
    string created_at = 4;  // RFC3339
    string updated_at = 5;
    string deleted_at = 6;
    RoomSettings settings = 7;
    string status = 8;
}

message RoomSettings {
    int32 max_participants = 1;
    bool recording_enabled = 2;
    repeated string allowed_users = 3;  // UUID
    string quality = 4;  // "HD", "SD"
}

message CreateRoomRequest {
    string name = 1;
    string owner_id = 2;  // из JWT
    RoomSettings settings = 3;
}

message CreateRoomResponse {
    Room room = 1;
    string join_token = 2;  // JWT для подключения к signaling
}

message GetRoomRequest {
    string room_id = 1;
}

message GetRoomResponse {
    Room room = 1;
}

// Сообщения для участников
message Participant {
    string user_id = 1;
    string joined_at = 2;
    string left_at = 3;
    bytes metadata = 4;  // JSON
}

message AddParticipantRequest {
    string room_id = 1;
    string user_id = 2;
    bytes metadata = 3;  // опционально
}

message AddParticipantResponse {
    bool success = 1;
    string join_token = 2;  // JWT
}

message RemoveParticipantRequest {
    string room_id = 1;
    string user_id = 2;
}

message RemoveParticipantResponse {
    bool success = 1;
}

message GetParticipantsRequest {
    string room_id = 1;
}

message GetParticipantsResponse {
    repeated Participant participants = 1;
}
```

---

## 7. Детали реализации

### 7.1 Обработка запросов

Каждый gRPC-метод должен:

1. Извлечь контекст с метаданными (`user_id` из JWT, переданный через заголовки)
2. Выполнить авторизацию (если требуется)
3. Вызвать соответствующий метод сервисного слоя
4. Обработать ошибки (преобразовать в gRPC status codes)
5. При успехе — вернуть ответ

**Пример структуры хендлера:**

```go
func (s *grpcServer) CreateRoom(ctx context.Context, req *pb.CreateRoomRequest) (*pb.CreateRoomResponse, error) {
    // 1. Получить user_id из контекста (устанавливается в интерцепторе)
    userID := ctx.Value("user_id").(string)
    if userID == "" {
        return nil, status.Error(codes.Unauthenticated, "missing user id")
    }
    // 2. Валидация входных данных
    if req.Name == "" {
        return nil, status.Error(codes.InvalidArgument, "room name required")
    }
    // 3. Вызов бизнес-логики
    room, token, err := s.service.CreateRoom(ctx, req.Name, userID, req.Settings)
    if err != nil {
        return nil, err
    }
    return &pb.CreateRoomResponse{Room: room, JoinToken: token}, nil
}
```

### 7.2 Бизнес-логика (service layer)

#### Создание комнаты

1. Сгенерировать UUID комнаты
2. Сохранить запись в PostgreSQL в транзакции
3. Сохранить в Redis кэш комнаты
4. Опубликовать событие `room_created` в Kafka
5. Сгенерировать JWT для входа (для владельца)

#### Добавление участника

1. Проверить существование и статус комнаты
2. Проверить, не превышен ли лимит
3. Для приватной комнаты — проверить, разрешён ли пользователь
4. Записать в таблицу `participants` (начать транзакцию)
5. Добавить `user_id` в Redis Set активных участников
6. Опубликовать событие `participant_joined` в Kafka
7. Сгенерировать и вернуть JWT для Signaling

#### Удаление участника

1. Обновить запись в `participants`, установив `left_at = NOW()`
2. Удалить `user_id` из Redis Set
3. Опубликовать событие `participant_left` в Kafka
4. Если комната пуста и включена опция `auto-close` — инициировать завершение комнаты

### 7.3 Кэширование в Redis

- **Инвалидация:** при любом изменении комнаты (update, delete) удалять ключ `room:{roomId}:info`
- **Активные участники:** использовать Redis Set для быстрой проверки присутствия
- **Срок жизни:** для информации о комнате — 1 час, для Set участников — без ограничения (удаляется при завершении комнаты)

### 7.4 Интеграция с Kafka

- Использовать продюсер с записью в синхронном режиме
- События должны содержать: тип, комната, пользователь, timestamp
- Для гарантии доставки: `acks=all` и ретраи
- При ошибке записи — логировать и, возможно, сохранять в отдельную таблицу для повторной отправки

### 7.5 Безопасность

- **Аутентификация:** gRPC interceptors проверяют JWT (authorization bearer), извлекают `user_id`
- **Авторизация:** проверка соответствия `user_id` владельцу для операций изменения/удаления
- **Токены для Signaling:** короткий срок жизни (5 минут), подпись общим секретом

### 7.6 Обработка ошибок

| Код gRPC | Использование |
|----------|---------------|
| `InvalidArgument` | Неверные входные данные |
| `NotFound` | Комната/участник не найдены |
| `PermissionDenied` | Недостаточно прав |
| `ResourceExhausted` | Комната заполнена |
| `AlreadyExists` | Участник уже в комнате |
| `Internal` | Внутренняя ошибка (БД, Kafka) |

---

## 8. Тестирование

### 8.1 Unit-тесты

- Покрыть бизнес-логику (service layer) моками репозиториев и продюсера
- Тестировать валидацию, проверку прав, лимиты

### 8.2 Интеграционные тесты

- Поднять PostgreSQL и Redis в Docker (testcontainers)
- Проверить все gRPC методы с реальной БД и кэшем
- Проверить публикацию событий в Kafka

### 8.3 Нагрузочное тестирование

- Инструменты: ghz, k6
- **Цель:** 1000 RPS с временем ответа < 100 мс

---

## 9. Развертывание

### 9.1 Конфигурация через переменные окружения

```bash
DB_HOST=postgres
DB_PORT=5432
DB_USER=room
DB_PASSWORD=xxx
DB_NAME=roomdb

REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=

KAFKA_BROKERS=kafka:9092
KAFKA_ROOM_EVENTS_TOPIC=room-events
KAFKA_PARTICIPANT_EVENTS_TOPIC=participant-events

JWT_SECRET=super-secret
SIGNALING_TOKEN_TTL=5m

GRPC_PORT=50051
METRICS_PORT=9090
```

### 9.2 Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o room-service ./cmd/room-service

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/room-service .
EXPOSE 50051 9090
CMD ["./room-service"]
```

### 9.3 Kubernetes манифесты

- Deployment с репликами (HPA по CPU/памяти)
- Service (ClusterIP) для gRPC
- ConfigMap с переменными окружения
- Secrets для паролей

---

## 10. Мониторинг и логирование

### 10.1 Метрики Prometheus

| Метрика | Описание |
|---------|----------|
| `room_service_requests_total{method, code}` | Количество запросов |
| `room_service_request_duration_seconds{method}` | Время ответа |
| `room_service_active_rooms` | Активные комнаты |
| `room_service_active_participants` | Активные участники |
| `room_service_kafka_errors_total` | Ошибки Kafka |

### 10.2 Логирование

- Использовать `log/slog` (структурированные логи)
- Поля: timestamp, level, service, method, request_id, user_id, room_id, error
- **Не логировать** токены и чувствительные данные

---

## 11. Дорожная карта разработки

| Этап | Срок | Описание |
|------|------|----------|
| 1 | 2 недели | Каркас сервиса, окружение, базовые CRUD для комнат |
| 2 | 2 недели | Управление участниками, лимиты, интеграция с Redis |
| 3 | 1 неделя | Интеграция с Kafka, публикация событий |
| 4 | 1 неделя | Аутентификация/авторизация, генерация JWT для Signaling |
| 5 | 1 неделя | Тестирование, оптимизация, документация |
| 6 | 1 неделя | Нагрузочное тестирование, подготовка к деплою |

---

## 12. Приложение

### 12.1 Пример JWT для Signaling

```json
{
  "user_id": "123",
  "room_id": "456",
  "exp": 1700000000,
  "iat": 1699996400
}
```

Подпись HMAC-SHA256.

### 12.2 Пример события Kafka (participant_joined)

```json
{
  "event_type": "participant_joined",
  "room_id": "456",
  "user_id": "123",
  "timestamp": "2024-01-01T12:00:00Z",
  "metadata": {
    "user_agent": "Mozilla/5.0..."
  }
}
```

---

## 13. Детальная спецификация структур данных и методов

### 13.1 Модели данных (domain)

#### Структура Room

```go
type Room struct {
    ID        string         `db:"id" json:"id"`
    Name      string         `db:"name" json:"name"`
    OwnerID   string         `db:"owner_id" json:"owner_id"`
    CreatedAt time.Time      `db:"created_at" json:"created_at"`
    UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
    DeletedAt *time.Time     `db:"deleted_at" json:"deleted_at,omitempty"`
    Settings  RoomSettings   `db:"settings" json:"settings"`
    Status    string         `db:"status" json:"status"`  // "active", "ended"
}

type RoomSettings struct {
    MaxParticipants  int      `json:"max_participants"`
    RecordingEnabled bool     `json:"recording_enabled"`
    AllowedUsers     []string `json:"allowed_users,omitempty"`
    Quality          string   `json:"quality"`     // "low", "medium", "high"
    AutoClose        bool     `json:"auto_close"`
}
```

#### Структура Participant

```go
type Participant struct {
    RoomID    string     `db:"room_id" json:"room_id"`
    UserID    string     `db:"user_id" json:"user_id"`
    JoinedAt  time.Time  `db:"joined_at" json:"joined_at"`
    LeftAt    *time.Time `db:"left_at" json:"left_at,omitempty"`
    Metadata  []byte     `db:"metadata" json:"metadata,omitempty"`
}

type ParticipantMeta struct {
    Name          string `json:"name,omitempty"`
    UserAgent     string `json:"user_agent,omitempty"`
    ClientVersion string `json:"client_version,omitempty"`
}
```

### 13.2 Репозитории

#### RoomRepository (PostgreSQL)

```go
type RoomRepository interface {
    Create(ctx context.Context, room *Room) error
    GetByID(ctx context.Context, id string) (*Room, error)
    Update(ctx context.Context, room *Room) error
    Delete(ctx context.Context, id string) error
    ListByOwner(ctx context.Context, ownerID string, limit, offset int) ([]*Room, error)
    ListActive(ctx context.Context) ([]*Room, error)
    EndRoom(ctx context.Context, id string) error
}
```

#### ParticipantRepository (PostgreSQL)

```go
type ParticipantRepository interface {
    Add(ctx context.Context, participant *Participant) error
    Remove(ctx context.Context, roomID, userID string) error
    GetActiveByRoom(ctx context.Context, roomID string) ([]*Participant, error)
    IsActive(ctx context.Context, roomID, userID string) (bool, error)
    UpdateMetadata(ctx context.Context, roomID, userID string, metadata []byte) error
    CountActiveInRoom(ctx context.Context, roomID string) (int, error)
}
```

#### CacheRepository (Redis)

```go
type CacheRepository interface {
    SetRoom(ctx context.Context, room *Room, ttl time.Duration) error
    GetRoom(ctx context.Context, roomID string) (*Room, error)
    DeleteRoom(ctx context.Context, roomID string) error

    AddParticipantToRoom(ctx context.Context, roomID, userID string) error
    RemoveParticipantFromRoom(ctx context.Context, roomID, userID string) error
    GetRoomParticipants(ctx context.Context, roomID string) ([]string, error)
    IsParticipantInRoom(ctx context.Context, roomID, userID string) (bool, error)
    CountRoomParticipants(ctx context.Context, roomID string) (int64, error)

    SetParticipantMeta(ctx context.Context, roomID, userID string, meta []byte) error
    GetParticipantMeta(ctx context.Context, roomID, userID string) ([]byte, error)
}
```

### 13.3 Интерфейс RoomService

```go
type RoomService interface {
    CreateRoom(ctx context.Context, name, ownerID string, settings *RoomSettings) (*Room, string, error)
    GetRoom(ctx context.Context, roomID string) (*Room, error)
    UpdateRoom(ctx context.Context, roomID, ownerID string, settings *RoomSettings) error
    DeleteRoom(ctx context.Context, roomID, ownerID string) error
    ListRooms(ctx context.Context, ownerID string, limit, offset int) ([]*Room, error)
    EndRoom(ctx context.Context, roomID string, ownerID string) error

    AddParticipant(ctx context.Context, roomID, userID string, metadata []byte) (string, error)
    RemoveParticipant(ctx context.Context, roomID, userID string) error
    GetParticipants(ctx context.Context, roomID string) ([]*Participant, error)
    IsParticipant(ctx context.Context, roomID, userID string) (bool, error)
    UpdateParticipantMetadata(ctx context.Context, roomID, userID string, metadata []byte) error

    ValidateAccess(ctx context.Context, roomID, userID string) error
}
```

### 13.4 Конфигурация сервиса

```go
type Config struct {
    Postgres PostgresConfig `envconfig:"DB"`
    Redis    RedisConfig    `envconfig:"REDIS"`
    Kafka    KafkaConfig    `envconfig:"KAFKA"`
    JWT      JWTConfig      `envconfig:"JWT"`
    GRPC     GRPCConfig     `envconfig:"GRPC"`
    Metrics  MetricsConfig  `envconfig:"METRICS"`
}

type PostgresConfig struct {
    Host     string `envconfig:"HOST" default:"localhost"`
    Port     int    `envconfig:"PORT" default:"5432"`
    User     string `envconfig:"USER" required:"true"`
    Password string `envconfig:"PASSWORD" required:"true"`
    DBName   string `envconfig:"NAME" required:"true"`
    SSLMode  string `envconfig:"SSLMODE" default:"disable"`
    MaxConns int    `envconfig:"MAX_CONNS" default:"20"`
}

type RedisConfig struct {
    Host     string `envconfig:"HOST" default:"localhost"`
    Port     int    `envconfig:"PORT" default:"6379"`
    Password string `envconfig:"PASSWORD"`
    DB       int    `envconfig:"DB" default:"0"`
}

type KafkaConfig struct {
    Brokers                []string `envconfig:"BROKERS" default:"localhost:9092"`
    RoomEventsTopic        string   `envconfig:"ROOM_EVENTS_TOPIC" default:"room-events"`
    ParticipantEventsTopic string   `envconfig:"PARTICIPANT_EVENTS_TOPIC" default:"participant-events"`
    Acks                   string   `envconfig:"ACKS" default:"all"`
}

type JWTConfig struct {
    Secret       string        `envconfig:"SECRET" required:"true"`
    SignalingTTL time.Duration `envconfig:"SIGNALING_TTL" default:"5m"`
}

type GRPCConfig struct {
    Port int `envconfig:"PORT" default:"50051"`
}

type MetricsConfig struct {
    Port int `envconfig:"PORT" default:"9090"`
}
```

### 13.5 EventProducer (Kafka)

```go
type EventProducer interface {
    PublishRoomCreated(ctx context.Context, roomID, ownerID string) error
    PublishRoomUpdated(ctx context.Context, roomID string) error
    PublishRoomDeleted(ctx context.Context, roomID string) error
    PublishRoomEnded(ctx context.Context, roomID string) error
    PublishParticipantJoined(ctx context.Context, roomID, userID string, metadata []byte) error
    PublishParticipantLeft(ctx context.Context, roomID, userID string) error
    Close() error
}
```

### 13.6 Итоговая структура проекта

```
cmd/
  room-service/
    main.go                 # точка входа, инициализация зависимостей

internal/
  config/                   # загрузка конфигурации
  domain/                   # модели данных (Room, Participant)
  repository/
    postgres/               # RoomRepository, ParticipantRepository
    redis/                  # CacheRepository
  service/                  # RoomService
  handler/                  # gRPC handlers
  kafka/                    # продюсер событий
  middleware/               # интерцепторы (auth, logging)

pkg/
  utils/                    # утилиты (uuid, time)

proto/                      # сгенерированный код из protobuf

docker-compose.yml
Dockerfile
```
