package schemas

import (
	"errors"
	"net/http"
	"strconv"
)

type CreateRateRequest struct {
	BaseCurrencyID  int64 `json:"base_currency_id"`
	QuoteCurrencyID int64 `json:"quote_currency_id"`
}

type CreateRateResponse struct {
	ID string `json:"id"`
}

func ValidateCreateRateRequest(req CreateRateRequest) error {
	if req.BaseCurrencyID <= 0 {
		return errors.New("base_currency_id is required")
	}
	if req.QuoteCurrencyID <= 0 {
		return errors.New("quote_currency_id is required")
	}
	if req.BaseCurrencyID == req.QuoteCurrencyID {
		return errors.New("base_currency_id and quote_currency_id must differ")
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
