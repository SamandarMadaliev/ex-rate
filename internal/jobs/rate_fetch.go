package jobs

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/SamandarMadaliev/ex-rate/internal/models"
	"github.com/SamandarMadaliev/ex-rate/internal/repositories"
	"github.com/SamandarMadaliev/ex-rate/internal/services"
	"github.com/SamandarMadaliev/ex-rate/internal/worker"
)

const jobTimeout = 10 * time.Second

const saveTimeout = 5 * time.Second

func FetchRateJob(
	rateID string,
	rateRepo *repositories.RateRepository,
	currencyRepo *repositories.CurrencyRepository,
	priceService *services.ExRateService,
) worker.Job {
	return func(ctx context.Context) {
		ctx, cancel := context.WithTimeout(ctx, jobTimeout)
		defer cancel()

		rate, err := rateRepo.GetByID(ctx, rateID)
		if err != nil {
			log.Printf("rate job %s: failed to load rate: %v", rateID, err)
			return
		}

		baseSlug, quoteSlug, err := currencyRepo.GetSlugsByIDs(ctx, rate.BaseCurrencyID, rate.QuoteCurrencyID)
		if err != nil {
			log.Printf("rate job %s: failed to resolve currency slugs: %v", rateID, err)
			saveResult(rateID, rateRepo, nil, models.RateStatusFailed, nil)
			return
		}

		body, err := priceService.GetRate(ctx, baseSlug, quoteSlug)
		if err != nil {
			log.Printf("rate job %s: failed to fetch price: %v", rateID, err)
			saveResult(rateID, rateRepo, nil, models.RateStatusFailed, nil)
			return
		}

		price, err := extractPrice(body, quoteSlug)
		if err != nil {
			log.Printf("rate job %s: failed to parse price: %v", rateID, err)
			saveResult(rateID, rateRepo, nil, models.RateStatusFailed, nil)
			return
		}

		now := time.Now()
		saveResult(rateID, rateRepo, &price, models.RateStatusCompleted, &now)
	}
}

func saveResult(rateID string, rateRepo *repositories.RateRepository, price *float64, status models.RateStatus, priceTimestamp *time.Time) {
	ctx, cancel := context.WithTimeout(context.Background(), saveTimeout)
	defer cancel()

	if err := rateRepo.UpdateResult(ctx, rateID, price, status, priceTimestamp); err != nil {
		log.Printf("rate job %s: failed to save result: %v", rateID, err)
	}
}

func extractPrice(body map[string]any, quoteSlug string) (float64, error) {
	rates, ok := body["rates"].(map[string]any)
	if !ok {
		return 0, fmt.Errorf("response missing \"rates\" object")
	}

	price, ok := rates[quoteSlug].(float64)
	if !ok {
		return 0, fmt.Errorf("no rate for %s in response", quoteSlug)
	}

	return price, nil
}
