package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	WSConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "signaling_ws_connections",
		Help: "Current WebSocket connections",
	})

	HandshakeFailuresTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "signaling_handshake_failures_total",
		Help: "WebSocket handshake failures by reason",
	}, []string{"reason"})

	MessagesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "signaling_messages_total",
		Help: "Signaling messages by type and direction",
	}, []string{"type", "direction"})

	MessageErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "signaling_message_errors_total",
		Help: "Signaling error messages by code",
	}, []string{"code"})

	RedisErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "signaling_redis_errors_total",
		Help: "Redis operation errors",
	}, []string{"op"})

	KafkaEventsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "signaling_kafka_events_total",
		Help: "Kafka events processed by type",
	}, []string{"event_type"})

	RouteLocalTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "signaling_route_local_total",
		Help: "Messages delivered via local hub",
	})

	RouteRemoteTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "signaling_route_remote_total",
		Help: "Messages published for remote delivery",
	})
)
