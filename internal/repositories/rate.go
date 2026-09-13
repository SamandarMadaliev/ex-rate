package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/SamandarMadaliev/ex-rate/internal/models"
)

var ErrRateNotFound = errors.New("rate not found")

type RateRepository struct {
	db *sql.DB
}

func NewRateRepository(db *sql.DB) *RateRepository {
	return &RateRepository{db: db}
}

func (r *RateRepository) Create(ctx context.Context, rate *models.Rate) error {
	query := `
		INSERT INTO rates (base_currency_id, quote_currency_id, price, status, price_timestamp)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	price := nullFloat64(rate.Price)
	priceTimestamp := nullTime(rate.PriceTimestamp)

	if err := r.db.QueryRowContext(ctx, query, rate.BaseCurrencyID, rate.QuoteCurrencyID, price, string(rate.Status), priceTimestamp).
		Scan(&rate.ID, &rate.CreatedAt); err != nil {
		return fmt.Errorf("failed to create rate: %w", err)
	}

	return nil
}

func (r *RateRepository) GetByID(ctx context.Context, id string) (*models.Rate, error) {
	query := `
		SELECT id, base_currency_id, quote_currency_id, price, status, price_timestamp, updated_at, created_at
		FROM rates
		WHERE id = $1
	`

	rate, err := scanRate(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRateNotFound
		}
		return nil, fmt.Errorf("failed to get rate by id: %w", err)
	}

	return rate, nil
}

func (r *RateRepository) GetLatest(ctx context.Context, baseCurrencyID, quoteCurrencyID int64) (*models.Rate, error) {
	query := `
		SELECT id, base_currency_id, quote_currency_id, price, status, price_timestamp, updated_at, created_at
		FROM rates
		WHERE base_currency_id = $1 AND quote_currency_id = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	rate, err := scanRate(r.db.QueryRowContext(ctx, query, baseCurrencyID, quoteCurrencyID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRateNotFound
		}
		return nil, fmt.Errorf("failed to get latest rate: %w", err)
	}

	return rate, nil
}

func scanRate(row *sql.Row) (*models.Rate, error) {
	var (
		rate           models.Rate
		price          sql.NullFloat64
		priceTimestamp sql.NullTime
		updatedAt      sql.NullTime
	)

	if err := row.Scan(
		&rate.ID,
		&rate.BaseCurrencyID,
		&rate.QuoteCurrencyID,
		&price,
		&rate.Status,
		&priceTimestamp,
		&updatedAt,
		&rate.CreatedAt,
	); err != nil {
		return nil, err
	}

	if price.Valid {
		rate.Price = &price.Float64
	}
	if priceTimestamp.Valid {
		rate.PriceTimestamp = &priceTimestamp.Time
	}
	if updatedAt.Valid {
		rate.UpdatedAt = &updatedAt.Time
	}

	return &rate, nil
}

func nullFloat64(f *float64) sql.NullFloat64 {
	if f == nil {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: *f, Valid: true}
}

func nullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}
