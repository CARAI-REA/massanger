# ============================
# Stage 1: Build
# ============================
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src

COPY platform/go.mod platform/go.sum ./platform/
COPY shared/go.mod shared/go.sum ./shared/
COPY rooms/go.mod rooms/go.sum ./rooms/

WORKDIR /src/rooms
RUN go mod download

WORKDIR /src
COPY platform ./platform
COPY shared ./shared
COPY rooms ./rooms

WORKDIR /src/rooms
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /room-service ./cmd

ADD https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.37/grpc_health_probe-linux-amd64 /grpc-health-probe
RUN chmod +x /grpc-health-probe

# ============================
# Stage 2: Runtime
# ============================
FROM alpine:3.21.3

RUN addgroup -S appgroup && adduser -S appuser -G appgroup \
  && apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /room-service ./room-service
COPY --from=builder /grpc-health-probe /bin/grpc-health-probe
COPY --from=builder /src/rooms/internal/migrations ./internal/migrations

USER appuser

EXPOSE 8080 9090

ENTRYPOINT ["./room-service"]
