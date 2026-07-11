package main

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds everything the ingestion service needs, loaded once at
// startup from environment variables. Keeping this in one place makes it
// trivial to see every external dependency a service has — useful both for
// onboarding a new engineer and for reasoning about failure modes.
type Config struct {
	KafkaBrokers      []string
	DataGovAPIKey     string
	DataGovResourceID string
	PollIntervalSecs  int
}

func loadConfig() (Config, error) {
	cfg := Config{
		KafkaBrokers:      []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
		DataGovAPIKey:     os.Getenv("DATA_GOV_API_KEY"),
		DataGovResourceID: os.Getenv("DATA_GOV_RESOURCE_ID"),
	}

	pollStr := getEnv("POLL_INTERVAL_SECONDS", "3600")
	pollSecs, err := strconv.Atoi(pollStr)
	if err != nil {
		return cfg, fmt.Errorf("invalid POLL_INTERVAL_SECONDS %q: %w", pollStr, err)
	}
	cfg.PollIntervalSecs = pollSecs

	if cfg.DataGovAPIKey == "" {
		return cfg, fmt.Errorf("DATA_GOV_API_KEY is required (get a free key at https://data.gov.in/user/register)")
	}
	if cfg.DataGovResourceID == "" {
		return cfg, fmt.Errorf("DATA_GOV_RESOURCE_ID is required (found on the dataset's catalog page)")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
