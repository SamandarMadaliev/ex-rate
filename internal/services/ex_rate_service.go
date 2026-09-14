package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ExRateService struct {
	client   *http.Client
	apiURL   string
	apiToken string
}

func NewService(client *http.Client, apiURL, apiToken string) *ExRateService {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &ExRateService{
		client:   client,
		apiURL:   strings.TrimRight(apiURL, "/"),
		apiToken: apiToken,
	}
}

// GetRate fetches the base/quote pair price from EX_RATE_API and returns the
// decoded response body as-is.
func (s *ExRateService) GetRate(ctx context.Context, baseSlug, quoteSlug string) (map[string]any, error) {
	q := url.Values{}
	q.Set("base", strings.ToUpper(baseSlug))
	q.Set("symbols", strings.ToUpper(quoteSlug))
	q.Set("access_key", s.apiToken)

	endpoint := s.apiURL + "/latest?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("pricing: failed to build request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("pricing: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pricing: unexpected status %d", resp.StatusCode)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("pricing: failed to decode response: %w", err)
	}

	return body, nil
}
