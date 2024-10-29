package app

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"rates_service/config"
	"rates_service/infrastructure/prommetrics"
	"rates_service/infrastructure/tracer"
	"rates_service/internal/modules/rates/repository"
	"rates_service/internal/modules/rates/service"
	"rates_service/internal/modules/rates/service/provider"
	"rates_service/internal/router"
	"rates_service/pkg/proto/gen/ratesservicepb"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	grpcZap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type App struct {
	cfg        config.AppCfg
	log        *zap.Logger
	db         *sql.DB
	grpcServer *grpc.Server
	httpServer *http.Server
}

func NewApp(cfg config.AppCfg, logger *zap.Logger) (*App, error) {
	app := &App{cfg: cfg, log: logger}

	Provider := provider.NewGarantexProvider(http.DefaultClient)

	Repository, db, err := repository.NewPostgresRepo(cfg)
	if err != nil {
		return nil, err
	}
	app.db = db

	if cfg.GatherMetrics {
		prommetrics.InitPromMetrics(app.db, logger)
	}

	ratesService := service.NewRateService(logger, cfg.Market, Provider, Repository)

	// tracing
	handler := tracer.InitTracer(cfg)

	// grpc server
	app.grpcServer = grpc.NewServer(
		grpc.ChainUnaryInterceptor(grpcZap.UnaryServerInterceptor(logger)),
		grpc.StatsHandler(handler), // tracing
	)
	ratesservicepb.RegisterRatesServiceServer(app.grpcServer, ratesService)
	grpc_health_v1.RegisterHealthServer(app.grpcServer, health.NewServer()) // healthcheck

	// http server
	app.httpServer = &http.Server{
		Addr:    app.cfg.HTTP.Host + ":" + app.cfg.HTTP.Port,
		Handler: router.InitChi(),
	}

	return app, nil
}

func (app *App) Run(sig chan os.Signal) error {
	// migration
	if app.cfg.DB.MigrationPath != "" {
		absMigr, err := filepath.Abs(app.cfg.DB.MigrationPath)
		if err != nil {
			return fmt.Errorf("%s: %w", "migrate instance init failed", err)
		}
		migr, err := migrate.New(fmt.Sprintf("file://%s", absMigr),
			fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
				app.cfg.DB.User, app.cfg.DB.Password, app.cfg.DB.Host, app.cfg.DB.Port, app.cfg.DB.Name))
		if err != nil {
			return fmt.Errorf("%s: %w", "migrate instance init failed", err)
		}
		if err := migr.Up(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("%s: %w", "db migration failed", err)
		}
	}
	// tracing
	if err := tracer.Start(context.Background()); err != nil {
		return fmt.Errorf("failed start tracer exporter: %w", err)
	}
	// grpc
	listener, err := net.Listen("tcp", app.cfg.GRPC.Host+":"+app.cfg.GRPC.Port)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	go func() {
		if err := app.grpcServer.Serve(listener); err != nil {
			app.log.Error("grpc serve failed", zap.Error(err))
		}
	}()
	// http
	go func() {
		if err := app.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.log.Error("http serve failed", zap.Error(err))
		}
	}()
	<-sig
	app.stop()
	return nil
}

func (app *App) stop() {
	if err := app.httpServer.Shutdown(context.Background()); err != nil {
		app.log.Error("http shutdown", zap.Error(err))
	}

	app.grpcServer.GracefulStop()

	if err := tracer.Shutdown(context.Background()); err != nil {
		app.log.Error("tracer provider shutdown call failed", zap.Error(err))
	}

	if err := app.db.Close(); err != nil {
		app.log.Error("http shutdown", zap.Error(err))
	}
}
