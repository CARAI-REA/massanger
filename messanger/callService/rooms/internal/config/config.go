package config

import (
	"fmt"
	"os"

	"github.com/CARAI-REA/messanger/callService/platform/pkg/prodguard"
	"github.com/joho/godotenv"

	"rooms/internal/config/env"
)

var appConfig *config

type config struct {
	App       AppEnvConfig
	Logger    LoggerConfig
	RoomsGRPC RoomsGRPCConfig
	Postgres  PostgresConfig
	Redis     RedisConfig
	Kafka     KafkaConfig
	JWT       JWTConfig
	TURN      TURNConfig
	Metrics   MetricsConfig
}

func Load(path ...string) error {
	err := godotenv.Load(path...)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	appCfg, err := env.NewAppConfig()
	if err != nil {
		return err
	}

	loggerCfg, err := env.NewLoggerConfig()
	if err != nil {
		return err
	}

	roomsGRPCCfg, err := env.NewRoomsGRPCConfig()
	if err != nil {
		return err
	}

	postgresCfg, err := env.NewPostgresConfig()
	if err != nil {
		return err
	}

	redisCfg, err := env.NewRedisConfig()
	if err != nil {
		return err
	}

	kafkaCfg, err := env.NewKafkaConfig()
	if err != nil {
		return err
	}

	jwtCfg, err := env.NewJWTConfig()
	if err != nil {
		return err
	}

	turnCfg, err := env.NewTURNConfig()
	if err != nil {
		return err
	}

	metricsCfg, err := env.NewMetricsConfig()
	if err != nil {
		return err
	}

	appConfig = &config{
		App:       appCfg,
		Logger:    loggerCfg,
		RoomsGRPC: roomsGRPCCfg,
		Postgres:  postgresCfg,
		Redis:     redisCfg,
		Kafka:     kafkaCfg,
		JWT:       jwtCfg,
		TURN:      turnCfg,
		Metrics:   metricsCfg,
	}

	return appConfig.ValidateProduction()
}

func AppConfig() *config {
	return appConfig
}

// ValidateProduction fails fast when APP_ENV=production and secrets/SSL/Redis are unsafe.
func (c *config) ValidateProduction() error {
	if c == nil || c.App == nil {
		return fmt.Errorf("config is not loaded")
	}
	if !prodguard.IsProduction(c.App.Env()) {
		return nil
	}

	checks := []error{
		prodguard.ForbidWeakSecret("JWT_SECRET", c.JWT.AuthTokenSecretKey()),
		prodguard.ForbidWeakSecret("JOIN_TOKEN_SECRET_KEY", c.JWT.JoinTokenSecretKey()),
		prodguard.ForbidWeakSecret("SERVICE_JWT_SECRET", c.JWT.ServiceTokenSecretKey()),
		prodguard.ForbidWeakSecret("TURN_SHARED_SECRET", c.TURN.SharedSecret()),
		prodguard.RequireNonEmpty("REDIS_PASSWORD", c.Redis.Password()),
		prodguard.ForbidDisabledSSL("POSTGRES_SSL_MODE", c.Postgres.SSLMode()),
	}
	for _, err := range checks {
		if err != nil {
			return err
		}
	}
	return nil
}
