package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/CARAI-REA/messanger/callService/platform/pkg/closer"
	"github.com/CARAI-REA/messanger/callService/platform/pkg/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"sfu/internal/config"
)

type App struct {
	diContainer   *diContainer
	httpServer    *http.Server
	metricsServer *http.Server
}

func New(ctx context.Context) (*App, error) {
	a := &App{}
	if err := a.initDeps(ctx); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 2)

	if consumer := a.diContainer.KafkaConsumer(ctx); consumer != nil {
		consumer.Start(ctx)
	}

	go func() {
		logger.Info(ctx, fmt.Sprintf("metrics server listening on %s", config.AppConfig().Metrics.Address()))
		if err := a.metricsServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	go func() {
		logger.Info(ctx, fmt.Sprintf("sfu HTTP/WS listening on %s path=%s",
			config.AppConfig().HTTP.Address(), config.AppConfig().HTTP.WSPath()))
		if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
}

func (a *App) initDeps(ctx context.Context) error {
	inits := []func(context.Context) error{
		a.initDi,
		a.initLogger,
		a.initCloser,
		a.initHTTPServer,
		a.initMetricsServer,
	}
	for _, f := range inits {
		if err := f(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) initDi(_ context.Context) error {
	a.diContainer = NewDiContainer()
	return nil
}

func (a *App) initLogger(_ context.Context) error {
	return logger.Init(config.AppConfig().Logger.Level(), config.AppConfig().Logger.AsJson())
}

func (a *App) initCloser(_ context.Context) error {
	closer.SetLogger(logger.Logger())
	return nil
}

func (a *App) initHTTPServer(ctx context.Context) error {
	a.httpServer = &http.Server{
		Addr:              config.AppConfig().HTTP.Address(),
		Handler:           a.diContainer.HTTPMux(ctx),
		ReadHeaderTimeout: 5 * time.Second,
	}
	closer.AddNamed("HTTP server", func(ctx context.Context) error {
		return a.httpServer.Shutdown(ctx)
	})
	return nil
}

func (a *App) initMetricsServer(_ context.Context) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	a.metricsServer = &http.Server{
		Addr:              config.AppConfig().Metrics.Address(),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	closer.AddNamed("metrics HTTP server", func(ctx context.Context) error {
		return a.metricsServer.Shutdown(ctx)
	})
	return nil
}
