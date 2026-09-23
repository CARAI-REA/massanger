package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/CARAI-REA/messanger/callService/platform/pkg/closer"
	"github.com/CARAI-REA/messanger/callService/platform/pkg/logger"
	"github.com/CARAI-REA/messanger/callService/platform/pkg/otelx"
	"go.uber.org/zap"

	"sfu/internal/app"
	"sfu/internal/config"
)

const configPath = "./../deploy/compose/sfu/.env"

func main() {
	if err := config.Load(configPath); err != nil {
		panic(fmt.Errorf("failed to load config: %w", err))
	}

	appCtx, appCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer appCancel()
	defer gracefulShutdown()

	closer.Configure(syscall.SIGINT, syscall.SIGTERM)

	otelShutdown, err := otelx.InitTracer(appCtx, "sfu")
	if err != nil {
		logger.Error(appCtx, "otel init failed", zap.Error(err))
	} else {
		closer.AddNamed("otel tracer", otelShutdown)
	}

	a, err := app.New(appCtx)
	if err != nil {
		logger.Error(appCtx, "failed to create application", zap.Error(err))
		return
	}
	if err := a.Run(appCtx); err != nil {
		logger.Error(appCtx, "application run error", zap.Error(err))
	}
}

func gracefulShutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := closer.CloseAll(ctx); err != nil {
		logger.Error(ctx, "graceful shutdown error", zap.Error(err))
	}
}
