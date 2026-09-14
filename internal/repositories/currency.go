package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/SamandarMadaliev/ex-rate/internal/models"
)

var ErrCurrencyNotFound = errors.New("currency not found")

type CurrencyRepository struct {
	db *sql.DB
}

func NewCurrencyRepository(db *sql.DB) *CurrencyRepository {
	return &CurrencyRepository{db: db}
}

func (r *CurrencyRepository) List(ctx context.Context) ([]*models.Currency, error) {
	query := `SELECT id, slug, is_active FROM currencies ORDER BY id`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list currencies: %w", err)
	}
	defer rows.Close()

	var currencies []*models.Currency
	for rows.Next() {
		c := &models.Currency{}
		if err := rows.Scan(&c.ID, &c.Slug, &c.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan currency: %w", err)
		}
		currencies = append(currencies, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to list currencies: %w", err)
	}

	return currencies, nil
}

func (r *CurrencyRepository) ValidateCurrencyPairByIDs(ctx context.Context, baseCurrencyID, quoteCurrencyID int64) (baseValid bool, quoteValid bool, err error) {
	query := `SELECT id FROM currencies WHERE id IN ($1, $2) AND is_active = true`

	rows, err := r.db.QueryContext(ctx, query, baseCurrencyID, quoteCurrencyID)
	if err != nil {
		return false, false, fmt.Errorf("failed to validate currency pair: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return false, false, fmt.Errorf("failed to scan currency id: %w", err)
		}
		if id == baseCurrencyID {
			baseValid = true
		}
		if id == quoteCurrencyID {
			quoteValid = true
		}
	}

	if err := rows.Err(); err != nil {
		return false, false, fmt.Errorf("failed to validate currency pair: %w", err)
	}

	return baseValid, quoteValid, nil
}

// GetSlugsByIDs resolves the currency slugs (e.g. "USD") for the given pair
// of currency IDs. It returns ErrCurrencyNotFound if either ID has no
// matching row.
func (r *CurrencyRepository) GetSlugsByIDs(ctx context.Context, baseCurrencyID, quoteCurrencyID int64) (baseSlug string, quoteSlug string, err error) {
	query := `SELECT id, slug FROM currencies WHERE id IN ($1, $2)`

	rows, err := r.db.QueryContext(ctx, query, baseCurrencyID, quoteCurrencyID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get currency slugs: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id   int64
			slug string
		)
		if err := rows.Scan(&id, &slug); err != nil {
			return "", "", fmt.Errorf("failed to scan currency slug: %w", err)
		}
		switch id {
		case baseCurrencyID:
			baseSlug = slug
		case quoteCurrencyID:
			quoteSlug = slug
		}
	}

	if err := rows.Err(); err != nil {
		return "", "", fmt.Errorf("failed to get currency slugs: %w", err)
	}

	if baseSlug == "" || quoteSlug == "" {
		return "", "", ErrCurrencyNotFound
	}

	return baseSlug, quoteSlug, nil
}
