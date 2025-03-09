package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	totalTrips = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "traewelling_total_trips",
		Help: "Gesamtanzahl der Fahrten",
	})
	totalDistance = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "traewelling_total_distance_km",
		Help: "Gesamtdistanz aller Fahrten in Kilometern",
	})
	averageDuration = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "traewelling_average_duration_minutes",
		Help: "Durchschnittliche Dauer der Fahrten in Minuten",
	})
)

type DashboardResponse struct {
	Data []struct {
		ID    int `json:"id"`
		Train struct {
			Trip      int     `json:"trip"`
			Distance  float64 `json:"distance"`
			Duration  int     `json:"duration"`
			Category  string  `json:"category"`
			LineName  string  `json:"lineName"`
			JourneyNumber int `json:"journeyNumber"`
			Origin    struct {
				Name string `json:"name"`
			} `json:"origin"`
			Destination struct {
				Name string `json:"name"`
			} `json:"destination"`
		} `json:"train"`
	} `json:"data"`
}

func fetchTraewellingData() (*DashboardResponse, error) {
	url := "https://traewelling.de/api/v1/dashboard"
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Fehler beim Erstellen der Anfrage: %v", err)
	}

	token := os.Getenv("TRAEWELLING_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("TRAEWELLING_TOKEN ist nicht gesetzt")
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Fehler beim Senden der Anfrage: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Fehlerhafte Antwort: %d", resp.StatusCode)
	}

	var apiResponse DashboardResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResponse)
	if err != nil {
		return nil, fmt.Errorf("Fehler beim Parsen der JSON-Daten: %v", err)
	}

	return &apiResponse, nil
}

func updateMetrics() {
	data, err := fetchTraewellingData()
	if err != nil {
		fmt.Printf("Fehler beim Abrufen der Daten: %v\n", err)
		return
	}

	var totalTripsCount int
	var totalDistanceSum float64
	var totalDurationSum int

	for _, trip := range data.Data {
		totalTripsCount++
		totalDistanceSum += trip.Train.Distance / 1000
		totalDurationSum += trip.Train.Duration
	}

	totalTrips.Set(float64(totalTripsCount))
	totalDistance.Set(totalDistanceSum)
	if totalTripsCount > 0 {
		averageDuration.Set(float64(totalDurationSum) / float64(totalTripsCount))
	}
}

func main() {
	go func() {
		for {
			updateMetrics()
			time.Sleep(30 * time.Second)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	fmt.Println("Server läuft auf Port 8080...")
	http.ListenAndServe(":8080", nil)
}
