package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/SamandarMadaliev/ex-rate/internal/http/helpers"
	"github.com/SamandarMadaliev/ex-rate/internal/http/schemas"
	"github.com/SamandarMadaliev/ex-rate/internal/jobs"
	"github.com/SamandarMadaliev/ex-rate/internal/models"
	"github.com/SamandarMadaliev/ex-rate/internal/repositories"
	"github.com/SamandarMadaliev/ex-rate/internal/services"
	"github.com/SamandarMadaliev/ex-rate/internal/worker"
	"github.com/go-chi/chi/v5"
)

// CreateRateHandler godoc
// @Summary      Create a rate
// @Description  Creates a pending rate row for a currency pair and enqueues a background job to fetch its price
// @Tags         rates
// @Accept       json
// @Produce      json
// @Param        request  body      schemas.CreateRateRequest  true  "Currency pair"
// @Success      201      {object}  schemas.CreateRateResponse
// @Failure      400      {object}  schemas.ErrorResponse
// @Failure      422      {object}  schemas.ErrorResponse
// @Failure      500      {object}  schemas.ErrorResponse
// @Router       /rates [post]
func CreateRateHandler(
	rateRepo *repositories.RateRepository,
	currencyRepo *repositories.CurrencyRepository,
	workers *worker.Pool,
	priceService *services.ExRateService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req schemas.CreateRateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			helpers.WriteError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if err := schemas.ValidateCreateRateRequest(req); err != nil {
			helpers.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		baseValid, quoteValid, err := currencyRepo.ValidateCurrencyPairByIDs(r.Context(), req.BaseCurrencyID, req.QuoteCurrencyID)
		if err != nil {
			log.Printf("create rate: failed to validate pair %d/%d: %v", req.BaseCurrencyID, req.QuoteCurrencyID, err)
			helpers.WriteError(w, http.StatusInternalServerError, "failed to validate currency pair")
			return
		}
		if !baseValid {
			log.Printf("create rate: base currency %d not found or inactive", req.BaseCurrencyID)
			helpers.WriteError(w, http.StatusUnprocessableEntity, "base currency not found or inactive")
			return
		}
		if !quoteValid {
			log.Printf("create rate: quote currency %d not found or inactive", req.QuoteCurrencyID)
			helpers.WriteError(w, http.StatusUnprocessableEntity, "quote currency not found or inactive")
			return
		}

		rate := &models.Rate{
			BaseCurrencyID:  req.BaseCurrencyID,
			QuoteCurrencyID: req.QuoteCurrencyID,
			Price:           nil,
			Status:          models.RateStatusPending,
		}

		if err := rateRepo.Create(r.Context(), rate); err != nil {
			log.Printf("create rate: failed to insert for %d/%d: %v", req.BaseCurrencyID, req.QuoteCurrencyID, err)
			helpers.WriteError(w, http.StatusInternalServerError, "failed to create rate")
			return
		}

		if err := workers.Submit(jobs.FetchRateJob(rate.ID, rateRepo, currencyRepo, priceService)); err != nil {
			log.Printf("create rate: failed to enqueue fetch job for %s: %v", rate.ID, err)
		} else {
			log.Printf("create rate: created %s for %d/%d, fetch job enqueued", rate.ID, rate.BaseCurrencyID, rate.QuoteCurrencyID)
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(schemas.CreateRateResponse{ID: rate.ID})
	}
}

// GetRateHandler godoc
// @Summary      Get a rate by ID
// @Description  Returns the rate row for the given ID
// @Tags         rates
// @Produce      json
// @Param        id   path      string  true  "Rate ID (UUID)"
// @Success      200  {object}  models.Rate
// @Failure      404  {object}  schemas.ErrorResponse
// @Failure      500  {object}  schemas.ErrorResponse
// @Router       /rates/{id} [get]
func GetRateHandler(rateRepo *repositories.RateRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if !helpers.IsValidUUID(id) {
			helpers.WriteError(w, http.StatusNotFound, "rate not found")
			return
		}

		rate, err := rateRepo.GetByID(r.Context(), id)
		if err != nil {
			if errors.Is(err, repositories.ErrRateNotFound) {
				helpers.WriteError(w, http.StatusNotFound, "rate not found")
				return
			}
			log.Printf("get rate: failed to load %s: %v", id, err)
			helpers.WriteError(w, http.StatusInternalServerError, "failed to get rate")
			return
		}

		json.NewEncoder(w).Encode(rate)
	}
}

// LatestRateHandler godoc
// @Summary      Get the latest rate for a pair
// @Description  Returns the most recently created rate for the given base/quote currency IDs
// @Tags         rates
// @Produce      json
// @Param        base   query     int  true  "Base currency ID"
// @Param        quote  query     int  true  "Quote currency ID"
// @Success      200    {object}  models.Rate
// @Failure      400    {object}  schemas.ErrorResponse
// @Failure      404    {object}  schemas.ErrorResponse
// @Failure      500    {object}  schemas.ErrorResponse
// @Router       /rates/latest [get]
func LatestRateHandler(rateRepo *repositories.RateRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := schemas.ParseGetLatestRateRequest(r)
		if err != nil {
			helpers.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		rate, err := rateRepo.GetLatest(r.Context(), req.BaseCurrencyID, req.QuoteCurrencyID)
		if err != nil {
			if errors.Is(err, repositories.ErrRateNotFound) {
				helpers.WriteError(w, http.StatusNotFound, "rate not found")
				return
			}
			log.Printf("latest rate: failed to load %d/%d: %v", req.BaseCurrencyID, req.QuoteCurrencyID, err)
			helpers.WriteError(w, http.StatusInternalServerError, "failed to get latest rate")
			return
		}

		json.NewEncoder(w).Encode(rate)
	}
}
