package models

import "time"

type Rate struct {
	ID              string     `json:"id"`
	BaseCurrencyID  int64      `json:"base_currency_id"`
	QuoteCurrencyID int64      `json:"quote_currency_id"`
	Price           *float64   `json:"price,omitempty"`
	Status          RateStatus `json:"status"`
	PriceTimestamp  *time.Time `json:"price_timestamp,omitempty"`
	UpdatedAt       *time.Time `json:"updated_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}
