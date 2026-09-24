package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/CARAI-REA/messanger/callService/platform/pkg/closer"
	"github.com/CARAI-REA/messanger/callService/platform/pkg/grpc/health"
	"github.com/CARAI-REA/messanger/callService/platform/pkg/logger"
	"github.com/CARAI-REA/messanger/callService/platform/pkg/prodguard"
	roomsV1 "github.com/CARAI-REA/messanger/callService/shared/pkg/proto/rooms/v1"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"rooms/internal/config"
	"rooms/internal/interceptor"
)

type App struct {
	diContainer  *diContainer
	grpcServer   *grpc.Server
	listener     net.Listener
	metricsServer *http.Server
}

func New(ctx context.Context) (*App, error) {
	a := &App{}

	err := a.initDeps(ctx)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		logger.Info(ctx, fmt.Sprintf("metrics server listening on %s", config.AppConfig().Metrics.Address()))
		if err := a.metricsServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	go func() {
		logger.Info(ctx, fmt.Sprintf("gRPC RoomService server listening on %s", config.AppConfig().RoomsGRPC.Address()))
		if err := a.grpcServer.Serve(a.listener); err != nil {
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
		a.initListener,
		a.initGRPCServer,
		a.initMetricsServer,
	}

	for _, f := range inits {
		err := f(ctx)
		if err != nil {
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
	return logger.Init(
		config.AppConfig().Logger.Level(),
		config.AppConfig().Logger.AsJson(),
	)
}

func (a *App) initCloser(_ context.Context) error {
	closer.SetLogger(logger.Logger())
	return nil
}

func (a *App) initListener(_ context.Context) error {
	listener, err := net.Listen("tcp", config.AppConfig().RoomsGRPC.Address())
	if err != nil {
		return err
	}

	closer.AddNamed("TCP listener", func(ctx context.Context) error {
		lerr := listener.Close()
		if lerr != nil && !errors.Is(lerr, net.ErrClosed) {
			return lerr
		}
		return nil
	})

	a.listener = listener
	return nil
}

func (a *App) initGRPCServer(ctx context.Context) error {
	a.grpcServer = grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			interceptor.AuthInterceptor(
				a.diContainer.AccessTokenVerifier(),
				a.diContainer.ServiceTokenVerifier(),
			),
			interceptor.RateLimitInterceptor(config.AppConfig().App.GRPCMaxRPS()),
			interceptor.MetricsInterceptor(),
			interceptor.LoggerInterceptor(),
		),
	)

	closer.AddNamed("gRPC server", func(ctx context.Context) error {
		a.grpcServer.GracefulStop()
		return nil
	})

	reflection.Register(a.grpcServer)
	health.RegisterService(a.grpcServer)

	roomsV1.RegisterRoomServiceServer(a.grpcServer, a.diContainer.RoomsV1API(ctx))

	return nil
}

func (a *App) initMetricsServer(_ context.Context) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	addr := config.AppConfig().Metrics.Address()
	if prodguard.IsProduction(config.AppConfig().App.Env()) {
		_, port, err := net.SplitHostPort(addr)
		if err == nil {
			addr = net.JoinHostPort("127.0.0.1", port)
		}
	}
	a.metricsServer = &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	closer.AddNamed("metrics HTTP server", func(ctx context.Context) error {
		return a.metricsServer.Shutdown(ctx)
	})

	return nil
}
