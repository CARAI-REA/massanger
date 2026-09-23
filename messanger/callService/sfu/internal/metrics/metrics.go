package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	WSConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "sfu_ws_connections",
		Help: "Current WebSocket connections",
	})
	PeerConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "sfu_peer_connections",
		Help: "Current WebRTC peer connections",
	})
	HandshakeFailuresTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sfu_handshake_failures_total",
		Help: "WS handshake failures",
	}, []string{"reason"})
	MessagesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sfu_messages_total",
		Help: "Control messages",
	}, []string{"type", "direction"})
	TracksPublished = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "sfu_tracks_published",
		Help: "Published remote tracks",
	})
	TracksSubscribed = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "sfu_tracks_subscribed",
		Help: "Local tracks subscribed by peers",
	})
	RTPPacketsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sfu_rtp_packets_total",
		Help: "RTP packets forwarded",
	}, []string{"direction"})
	KafkaEventsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sfu_kafka_events_total",
		Help: "Kafka events",
	}, []string{"event_type"})
	AffinityRedirectsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "sfu_affinity_redirects_total",
		Help: "Affinity redirects",
	})
	RedisErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sfu_redis_errors_total",
		Help: "Redis errors",
	}, []string{"op"})
	MessageErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sfu_message_errors_total",
		Help: "Error messages by code",
	}, []string{"code"})
)
