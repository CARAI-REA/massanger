package config

import (
	"fmt"
	"os"

	"github.com/CARAI-REA/messanger/callService/platform/pkg/prodguard"
	"github.com/joho/godotenv"

	"sfu/internal/config/env"
)

var appConfig *config

type config struct {
	AppEnv   AppEnvConfig
	Logger   LoggerConfig
	HTTP     HTTPConfig
	Redis    RedisConfig
	JWT      JWTConfig
	RoomGRPC RoomGRPCConfig
	Kafka    KafkaConfig
	Metrics  MetricsConfig
	Instance InstanceConfig
	WebRTC   WebRTCConfig
}

func Load(path ...string) error {
	err := godotenv.Load(path...)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	appEnvCfg, err := env.NewAppConfig()
	if err != nil {
		return err
	}
	loggerCfg, err := env.NewLoggerConfig()
	if err != nil {
		return err
	}
	httpCfg, err := env.NewHTTPConfig()
	if err != nil {
		return err
	}
	redisCfg, err := env.NewRedisConfig()
	if err != nil {
		return err
	}
	jwtCfg, err := env.NewJWTConfig()
	if err != nil {
		return err
	}
	roomGRPCCfg, err := env.NewRoomGRPCConfig()
	if err != nil {
		return err
	}
	kafkaCfg, err := env.NewKafkaConfig()
	if err != nil {
		return err
	}
	metricsCfg, err := env.NewMetricsConfig()
	if err != nil {
		return err
	}
	instanceCfg, err := env.NewInstanceConfig()
	if err != nil {
		return err
	}
	webrtcCfg, err := env.NewWebRTCConfig()
	if err != nil {
		return err
	}

	appConfig = &config{
		AppEnv:   appEnvCfg,
		Logger:   loggerCfg,
		HTTP:     httpCfg,
		Redis:    redisCfg,
		JWT:      jwtCfg,
		RoomGRPC: roomGRPCCfg,
		Kafka:    kafkaCfg,
		Metrics:  metricsCfg,
		Instance: instanceCfg,
		WebRTC:   webrtcCfg,
	}
	return appConfig.ValidateProduction()
}

func (c *config) ValidateProduction() error {
	if !prodguard.IsProduction(c.AppEnv.Env()) {
		return nil
	}
	if err := prodguard.ForbidWeakSecret("JOIN_TOKEN_SECRET_KEY", c.JWT.JoinTokenSecretKey()); err != nil {
		return err
	}
	if err := prodguard.ForbidWeakSecret("SERVICE_JWT_SECRET", c.JWT.ServiceTokenSecretKey()); err != nil {
		return err
	}
	if err := prodguard.RequireNonEmpty("REDIS_PASSWORD", c.Redis.Password()); err != nil {
		return err
	}
	if len(c.HTTP.AllowedOrigins()) == 0 {
		return fmt.Errorf("production: WS_ALLOWED_ORIGINS is required")
	}
	if err := prodguard.ForbidLocalhostURL("PUBLIC_WS_URL", c.Instance.PublicWSURL()); err != nil {
		return err
	}
	if len(c.WebRTC.NAT1To1IPs()) == 0 {
		return fmt.Errorf("production: WEBRTC_NAT_1TO1_IPS is required")
	}
	if err := prodguard.ForbidWeakSecret("WEBRTC_TURN_CREDENTIAL", c.WebRTC.TURNCredential()); err != nil {
		return err
	}
	return nil
}

func AppConfig() *config { return appConfig }
