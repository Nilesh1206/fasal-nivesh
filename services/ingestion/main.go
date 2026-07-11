package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// district is a minimal seed list of (name, lat, lon) pairs to pull weather
// for. Replace/extend this with the actual districts your mandi data
// covers — in a real deployment this would live in Cassandra or a config
// file, not hardcoded, but it's a fine starting point.
type district struct {
	name string
	lat  float64
	lon  float64
}

var seedDistricts = []district{
	{"Pune", 18.5204, 73.8567},
	{"Nashik", 19.9975, 73.7898},
	{"Ahmednagar", 19.0952, 74.7496},
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	dataGov := NewDataGovClient(cfg.DataGovAPIKey, cfg.DataGovResourceID)
	nasaPower := NewNasaPowerClient()
	publisher := NewPublisher(cfg.KafkaBrokers)
	defer publisher.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown: stop the poll loop cleanly on SIGINT/SIGTERM so
	// in-flight Kafka writes aren't abandoned mid-batch.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("shutdown signal received, stopping poll loop...")
		cancel()
	}()

	ticker := time.NewTicker(time.Duration(cfg.PollIntervalSecs) * time.Second)
	defer ticker.Stop()

	// Run once immediately on startup, then on every tick.
	runOnce(ctx, dataGov, nasaPower, publisher)
	for {
		select {
		case <-ctx.Done():
			log.Println("ingestion service stopped")
			return
		case <-ticker.C:
			runOnce(ctx, dataGov, nasaPower, publisher)
		}
	}
}

func runOnce(ctx context.Context, dataGov *DataGovClient, nasaPower *NASAPowerClient, publisher *Publisher) {
	log.Println("fetching mandi prices...")
	prices, err := dataGov.FetchLatestPrices(100, 0)
	if err != nil {
		// Log and continue rather than crash — one bad poll cycle
		// shouldn't take down a service that'll try again next tick.
		log.Printf("ERROR fetching prices: %v", err)
	} else {
		if err := publisher.PublishPrices(ctx, prices); err != nil {
			log.Printf("ERROR publishing prices: %v", err)
		} else {
			log.Printf("published %d price records", len(prices))
		}
	}

	log.Println("fetching weather data...")
	end := time.Now()
	start := end.AddDate(0, 0, -7) // last 7 days, keeps the payload small
	startStr := formatYYYYMMDD(start)
	endStr := formatYYYYMMDD(end)

	for _, d := range seedDistricts {
		weather, err := nasaPower.FetchDailyWeather(d.name, d.lat, d.lon, startStr, endStr)
		if err != nil {
			log.Printf("ERROR fetching weather for %s: %v", d.name, err)
			continue
		}
		if err := publisher.PublishWeather(ctx, weather); err != nil {
			log.Printf("ERROR publishing weather for %s: %v", d.name, err)
			continue
		}
		log.Printf("published %d weather records for %s", len(weather), d.name)
	}
}

func formatYYYYMMDD(t time.Time) string {
	return t.Format("20060102")
}
