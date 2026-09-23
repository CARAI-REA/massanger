package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "room_service_requests_total",
			Help: "Total number of gRPC requests handled by room service",
		},
		[]string{"method", "code"},
	)

	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "room_service_request_duration_seconds",
			Help:    "gRPC request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	ActiveRooms = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "room_service_active_rooms",
			Help: "Approximate number of active rooms tracked by this instance",
		},
	)

	ActiveParticipants = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "room_service_active_participants",
			Help: "Approximate number of active participants tracked by this instance",
		},
	)

	KafkaErrorsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "room_service_kafka_errors_total",
			Help: "Total number of Kafka publish errors",
		},
	)
)
