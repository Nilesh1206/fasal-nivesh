package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// PriceRecord mirrors one row of the data.gov.in "Current daily price of
// various commodities" dataset. Field names on the source API are messy
// (mixed casing, occasional typos across dataset versions) so this struct
// is the single place that translates their shape into ours.
type PriceRecord struct {
	State       string  `json:"state"`
	District    string  `json:"district"`
	Market      string  `json:"market"`
	Commodity   string  `json:"commodity"`
	Variety     string  `json:"variety"`
	ArrivalDate string  `json:"arrival_date"`
	MinPrice    float64 `json:"min_price"`
	MaxPrice    float64 `json:"max_price"`
	ModalPrice  float64 `json:"modal_price"`
}

type dataGovResponse struct {
	Records []rawRecord `json:"records"`
}

// rawRecord captures fields as strings first — the data.gov.in API is
// inconsistent about whether numeric fields are quoted, so we parse
// defensively rather than trusting json.Unmarshal to do it for us.
type rawRecord struct {
	State       string `json:"state"`
	District    string `json:"district"`
	Market      string `json:"market"`
	Commodity   string `json:"commodity"`
	Variety     string `json:"variety"`
	ArrivalDate string `json:"arrival_date"`
	MinPrice    string `json:"min_price"`
	MaxPrice    string `json:"max_price"`
	ModalPrice  string `json:"modal_price"`
}

type DataGovClient struct {
	apiKey     string
	resourceID string
	httpClient *http.Client
}

func NewDataGovClient(apiKey, resourceID string) *DataGovClient {
	return &DataGovClient{
		apiKey:     apiKey,
		resourceID: resourceID,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// FetchLatestPrices pulls the most recent page of mandi price records.
// limit/offset support pagination — start with a modest limit while you're
// building, then raise it once you're confident the pipeline behaves.
func (c *DataGovClient) FetchLatestPrices(limit, offset int) ([]PriceRecord, error) {
	endpoint := fmt.Sprintf("https://api.data.gov.in/resource/%s", c.resourceID)
	q := url.Values{}
	q.Set("api-key", c.apiKey)
	q.Set("format", "json")
	q.Set("limit", fmt.Sprintf("%d", limit))
	q.Set("offset", fmt.Sprintf("%d", offset))

	resp, err := c.httpClient.Get(endpoint + "?" + q.Encode())
	if err != nil {
		return nil, fmt.Errorf("fetching mandi prices: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("data.gov.in returned status %d", resp.StatusCode)
	}

	var parsed dataGovResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("decoding data.gov.in response: %w", err)
	}

	records := make([]PriceRecord, 0, len(parsed.Records))
	for _, r := range parsed.Records {
		records = append(records, PriceRecord{
			State:       r.State,
			District:    r.District,
			Market:      r.Market,
			Commodity:   r.Commodity,
			Variety:     r.Variety,
			ArrivalDate: r.ArrivalDate,
			MinPrice:    parseFloatSafe(r.MinPrice),
			MaxPrice:    parseFloatSafe(r.MaxPrice),
			ModalPrice:  parseFloatSafe(r.ModalPrice),
		})
	}
	return records, nil
}

// parseFloatSafe converts a possibly messy numeric string into a float64.
// It trims whitespace, drops thousands separators (commas), and treats
// empty or non-numeric values as 0.
func parseFloatSafe(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	// Some datasets use commas as thousand separators; remove them.
	s = strings.ReplaceAll(s, ",", "")
	// Reject common non-numeric tokens.
	lower := strings.ToLower(s)
	if lower == "na" || lower == "n/a" || lower == "-" || lower == "null" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}
