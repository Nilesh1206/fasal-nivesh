package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// WeatherRecord is our normalized shape for one district-day of weather.
type WeatherRecord struct {
	District   string  `json:"district"`
	Date       string  `json:"date"` // YYYY-MM-DD
	RainfallMM float64 `json:"rainfall_mm"`
	TempAvgC   float64 `json:"temp_avg_c"`
	HumidityPc float64 `json:"humidity_pct"`
	SolarRad   float64 `json:"solar_rad"`
}

// nasaPowerResponse mirrors the relevant slice of NASA POWER's JSON shape:
// properties.parameter.<PARAM>.<YYYYMMDD> = value
type nasaPowerResponse struct {
	Properties struct {
		Parameter map[string]map[string]float64 `json:"parameter"`
	} `json:"properties"`
}

type NASAPowerClient struct {
	httpClient *http.Client
}

func NewNasaPowerClient() *NASAPowerClient {
	return &NASAPowerClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// FetchDailyWeather pulls rainfall, temperature, humidity, and solar
// radiation for a single point (lat/lon) between start and end dates
// (both YYYYMMDD). No API key needed — this is a genuinely free, public
// NASA endpoint, which is why it's a good fit for a zero-budget project.

func (c *NASAPowerClient) FetchDailyWeather(district string, lat, lon float64, start, end string) ([]WeatherRecord, error) {
	q := url.Values{}
	q.Set("parameters", "T2M,PRECTOTCORR,RH2M,ALLSKY_SFC_SW_DWN")
	q.Set("community", "AG")
	q.Set("longitude", fmt.Sprintf("%f", lon))
	q.Set("latitude", fmt.Sprintf("%f", lat))
	q.Set("start", start)
	q.Set("end", end)
	q.Set("format", "JSON")

	endpoint := "https://power.larc.nasa.gov/api/temporal/daily/point?" + q.Encode()

	resp, err := c.httpClient.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("fetching NASA POWER data: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("NASA POWER returned status %d", resp.StatusCode)
	}

	var parsed nasaPowerResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("decoding NASA POWER data: %w", err)
	}

	temp := parsed.Properties.Parameter["T2M"]
	rain := parsed.Properties.Parameter["PRECTOTCORR"]
	humidity := parsed.Properties.Parameter["RH2M"]
	solar := parsed.Properties.Parameter["ALLSKY_SFC_SW_DWN"]

	records := make([]WeatherRecord, 0, len(temp))
	for dateKey, tempVal := range temp {
		records = append(records, WeatherRecord{
			District:   district,
			Date:       formatNASADate(dateKey),
			TempAvgC:   tempVal,
			RainfallMM: rain[dateKey],
			HumidityPc: humidity[dateKey],
			SolarRad:   solar[dateKey],
		})
	}
	return records, nil
}

// formatNASADate converts NASA's YYYYMMDD key into YYYY-MM-DD.
func formatNASADate(raw string) string {
	if len(raw) != 8 {
		return raw
	}
	return raw[0:4] + "-" + raw[4:6] + "-" + raw[6:8]
}
