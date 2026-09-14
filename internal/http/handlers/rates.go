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
			helpers.WriteError(w, http.StatusInternalServerError, "failed to validate currency pair")
			return
		}
		if !baseValid {
			helpers.WriteError(w, http.StatusUnprocessableEntity, "base currency not found or inactive")
			return
		}
		if !quoteValid {
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
			helpers.WriteError(w, http.StatusInternalServerError, "failed to create rate")
			return
		}

		if err := workers.Submit(jobs.FetchRateJob(rate.ID, rateRepo, currencyRepo, priceService)); err != nil {
			log.Printf("failed to enqueue rate fetch job for %s: %v", rate.ID, err)
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(schemas.CreateRateResponse{ID: rate.ID})
	}
}

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
			helpers.WriteError(w, http.StatusInternalServerError, "failed to get rate")
			return
		}

		json.NewEncoder(w).Encode(rate)
	}
}

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
			helpers.WriteError(w, http.StatusInternalServerError, "failed to get latest rate")
			return
		}

		json.NewEncoder(w).Encode(rate)
	}
}
