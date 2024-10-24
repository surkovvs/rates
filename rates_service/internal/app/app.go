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
)

type App struct {
	cfg        config.AppCfg
	log        *zap.Logger
	db         *sql.DB
	grpcServer *grpc.Server
	httpServer *http.Server
	router     http.Handler
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

	app.grpcServer = grpc.NewServer(grpc.ChainUnaryInterceptor(grpcZap.UnaryServerInterceptor(logger)))
	ratesservicepb.RegisterRatesServiceServer(app.grpcServer, ratesService)

	app.router = router.InitChi()

	return app, nil
}

func (a *App) Run(sig chan os.Signal) error {
	if a.cfg.DB.MigrationPath != "" {
		absMigr, err := filepath.Abs(a.cfg.DB.MigrationPath)
		if err != nil {
			return fmt.Errorf("%s: %w", "migrate instance init failed", err)
		}
		migr, err := migrate.New(fmt.Sprintf("file://%s", absMigr),
			fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
				a.cfg.DB.User, a.cfg.DB.Password, a.cfg.DB.Host, a.cfg.DB.Port, a.cfg.DB.Name))
		if err != nil {
			return fmt.Errorf("%s: %w", "migrate instance init failed", err)
		}
		if err := migr.Up(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("%s: %w", "db migration failed", err)
		}
	}
	listener, err := net.Listen("tcp", a.cfg.GRPC.Host+":"+a.cfg.GRPC.Port)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	go func() {
		if err := a.grpcServer.Serve(listener); err != nil {
			a.log.Error("grpc serve failed", zap.Error(err))
		}
	}()
	a.httpServer = &http.Server{
		Addr:    a.cfg.HTTP.Host + ":" + a.cfg.HTTP.Port,
		Handler: a.router,
	}
	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.log.Error("http serve failed", zap.Error(err))
		}
	}()
	<-sig
	a.stop()
	return nil
}

func (a *App) stop() {
	if err := a.httpServer.Shutdown(context.Background()); err != nil {
		a.log.Error("http shutdown", zap.Error(err))
	}
	a.grpcServer.GracefulStop()
	if err := a.db.Close(); err != nil {
		a.log.Error("http shutdown", zap.Error(err))
	}
}
