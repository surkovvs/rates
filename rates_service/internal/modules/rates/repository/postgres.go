package repository

import (
	"context"
	"fmt"
	"rates_service/config"
	"rates_service/internal/models"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

type PostgresRepo struct {
	db  *sqlx.DB
	psv *prometheus.SummaryVec
}

func NewPostgresRepo(cfg config.AppCfg) (PostgresRepo, prometheus.Collector, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Name, cfg.DB.Password)
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return PostgresRepo{}, nil, err
	}
	return PostgresRepo{
			db: db,
		},
		collectors.NewDBStatsCollector(db.DB, "rates_db"),
		nil
}

func (pr *PostgresRepo) Create(ctx context.Context, rates models.RatesDTO) error {
	startTime := time.Now()
	defer pr.psv.WithLabelValues("").Observe(float64(time.Since(startTime).Seconds()))

	res, err := pr.db.ExecContext(ctx, `
		INSERT INTO rates
		(stamp,ask_price,bid_price)
		VALUES
		($1,$2,$3);
	`, rates.Timestamp, rates.AskPrice, rates.BidPrice)
	if err != nil {
		return err
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if ra != 1 {
		return fmt.Errorf("unexpected num of rows has been inserted: %d, want 1", ra)
	}
	return nil
}
