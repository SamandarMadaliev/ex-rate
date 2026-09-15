package schemas

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/SamandarMadaliev/ex-rate/internal/models"
)

type CreateRateRequest struct {
	BaseCurrency  string `json:"base_currency"`
	QuoteCurrency string `json:"quote_currency"`
}

type CreateRateResponse struct {
	ID string `json:"id"`
}

// RateResponse is the JSON shape returned for a single rate. It mirrors
// models.Rate but exposes the currency pair as slugs (e.g. "USD") instead of
// internal currency IDs.
type RateResponse struct {
	ID             string            `json:"id"`
	BaseCurrency   string            `json:"base_currency"`
	QuoteCurrency  string            `json:"quote_currency"`
	Price          *float64          `json:"price,omitempty"`
	Status         models.RateStatus `json:"status"`
	PriceTimestamp *time.Time        `json:"price_timestamp,omitempty"`
	UpdatedAt      *time.Time        `json:"updated_at,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
}

func NewRateResponse(rate *models.Rate, baseCurrency, quoteCurrency string) RateResponse {
	return RateResponse{
		ID:             rate.ID,
		BaseCurrency:   baseCurrency,
		QuoteCurrency:  quoteCurrency,
		Price:          rate.Price,
		Status:         rate.Status,
		PriceTimestamp: rate.PriceTimestamp,
		UpdatedAt:      rate.UpdatedAt,
		CreatedAt:      rate.CreatedAt,
	}
}

func ValidateCreateRateRequest(req CreateRateRequest) error {
	if strings.TrimSpace(req.BaseCurrency) == "" {
		return errors.New("base_currency is required")
	}
	if strings.TrimSpace(req.QuoteCurrency) == "" {
		return errors.New("quote_currency is required")
	}
	if strings.EqualFold(req.BaseCurrency, req.QuoteCurrency) {
		return errors.New("base_currency and quote_currency must differ")
	}
	return nil
}

type GetLatestRateRequest struct {
	BaseCurrencyID  int64
	QuoteCurrencyID int64
}

func ParseGetLatestRateRequest(r *http.Request) (GetLatestRateRequest, error) {
	baseCurrencyID, err := strconv.ParseInt(r.URL.Query().Get("base"), 10, 64)
	if err != nil || baseCurrencyID <= 0 {
		return GetLatestRateRequest{}, errors.New("base is required")
	}

	quoteCurrencyID, err := strconv.ParseInt(r.URL.Query().Get("quote"), 10, 64)
	if err != nil || quoteCurrencyID <= 0 {
		return GetLatestRateRequest{}, errors.New("quote is required")
	}

	return GetLatestRateRequest{
		BaseCurrencyID:  baseCurrencyID,
		QuoteCurrencyID: quoteCurrencyID,
	}, nil
}
