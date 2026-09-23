# OpenTelemetry (example)

All services accept:

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317
OTEL_SERVICE_NAME=rooms|signaling|sfu
OTEL_TRACES_EXPORTER=otlp
```

Instrument:

- rooms: gRPC unary spans (method, user_id)
- signaling/sfu: WS connect, AssertCanJoin, offer/answer/negotiate

Prometheus `/metrics` remains the primary RED metrics source; traces complement latency debugging.
