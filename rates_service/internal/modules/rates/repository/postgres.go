package repository

import (
	"context"
	"database/sql"
	"fmt"
	"rates_service/config"
	"rates_service/internal/models"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type PostgresRepo struct {
	db *sqlx.DB
}

func NewPostgresRepo(cfg config.AppCfg) (PostgresRepo, *sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Name, cfg.DB.Password)
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return PostgresRepo{}, nil, err
	}
	return PostgresRepo{
			db: db,
		},
		db.DB,
		nil
}

func (pr PostgresRepo) Create(ctx context.Context, rates models.RatesDTO) error {
	res, err := pr.db.ExecContext(ctx, `
		INSERT INTO rates
		(stamp,ask_price,bid_price)
		VALUES
		($1,$2,$3);
	`, time.Unix(rates.Timestamp, 0), rates.AskPrice, rates.BidPrice)
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
