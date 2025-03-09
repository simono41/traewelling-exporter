package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type StatusResponse struct {
	Data []struct {
		ID    int `json:"id"`
		User  struct {
			DisplayName string `json:"displayName"`
			Username    string `json:"username"`
		} `json:"userDetails"`
		Train struct {
			Trip      int     `json:"trip"`
			Category  string  `json:"category"`
			LineName  string  `json:"lineName"`
			JourneyNumber int `json:"journeyNumber"`
			Distance  float64 `json:"distance"`
			Duration  int     `json:"duration"`
			Points    int     `json:"points"`
			Origin    struct {
				Name             string `json:"name"`
				DeparturePlanned string `json:"departurePlanned"`
				DepartureReal    string `json:"departureReal"`
			} `json:"origin"`
			Destination struct {
				Name           string `json:"name"`
				ArrivalPlanned string `json:"arrivalPlanned"`
				ArrivalReal    string `json:"arrivalReal"`
			} `json:"destination"`
		} `json:"train"`
	} `json:"data"`
}

var (
	totalTripDuration = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "traewelling_total_trip_duration_minutes",
			Help: "Gesamtfahrzeit eines Benutzers in Minuten",
		},
		[]string{"username"},
	)
	totalTripPoints = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "traewelling_total_trip_points",
			Help: "Gesamtpunkte eines Benutzers",
		},
		[]string{"username"},
	)
	totalTripDistance = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "traewelling_total_trip_distance_km",
			Help: "Gesamtentfernung eines Benutzers in Kilometern",
		},
		[]string{"username"},
	)
)

func fetchStatuses(username string) (*StatusResponse, error) {
	url := fmt.Sprintf("https://traewelling.de/api/v1/user/%s/statuses", username)
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
        return nil, fmt.Errorf("Fehler beim Senden der Anfrage für Benutzer '%s': %v", username, err)
    }
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("Fehlerhafte Antwort für Benutzer '%s': %d", username, resp.StatusCode)
    }

	var apiResponse StatusResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResponse)
	if err != nil {
        return nil, fmt.Errorf("Fehler beim Parsen der JSON-Daten für Benutzer '%s': %v", username, err)
    }

    return &apiResponse, nil
}

func processAndPrintTripDetails(username string, data *StatusResponse) {
	fmt.Printf("Fahrtdetails für Benutzer '%s':\n", username)
	fmt.Println("==========================")

	var totalDuration float64
	var totalPoints int
	var totalDistance float64

	for _, trip := range data.Data {
        fmt.Printf("Fahrt-ID: %d\n", trip.Train.Trip)
        fmt.Printf("Kategorie: %s\n", trip.Train.Category)
        fmt.Printf("Linie: %s (Nummer: %d)\n", trip.Train.LineName, trip.Train.JourneyNumber)
        fmt.Printf("Startbahnhof: %s\n", trip.Train.Origin.Name)
        fmt.Printf("\tGeplante Abfahrt: %s\n", trip.Train.Origin.DeparturePlanned)
        fmt.Printf("\tTatsächliche Abfahrt: %s\n", trip.Train.Origin.DepartureReal)
        fmt.Printf("Zielbahnhof: %s\n", trip.Train.Destination.Name)
        fmt.Printf("\tGeplante Ankunft: %s\n", trip.Train.Destination.ArrivalPlanned)
        fmt.Printf("\tTatsächliche Ankunft: %s\n", trip.Train.Destination.ArrivalReal)
        fmt.Printf("Dauer: %d Minuten\n", trip.Train.Duration)
        fmt.Printf("Entfernung: %.2f km\n", trip.Train.Distance/1000)
        fmt.Printf("Punkte: %d\n", trip.Train.Points)
        fmt.Println("--------------------------")

        totalDuration += float64(trip.Train.Duration)
        totalPoints += trip.Train.Points
        totalDistance += trip.Train.Distance / 1000
    }

	totalTripDuration.WithLabelValues(username).Set(totalDuration)
	totalTripPoints.WithLabelValues(username).Set(float64(totalPoints))
	totalTripDistance.WithLabelValues(username).Set(totalDistance)

	fmt.Printf("\nGesamtfahrzeit für %s: %.2f Minuten\n", username, totalDuration)
	fmt.Printf("Gesamtpunkte für %s: %d Punkte\n", username, totalPoints)
	fmt.Printf("Gesamtentfernung für %s: %.2f km\n", username, totalDistance)
}

func updateMetrics() {
	usernames := os.Getenv("TRAEWELLING_USERNAMES")
	if usernames == "" {
	    fmt.Println("TRAEWELLING_USERNAMES ist nicht gesetzt")
	    return
    }

	for _, username := range strings.Split(usernames, ",") {
	    username = strings.TrimSpace(username) // Entferne Leerzeichen
	    data, err := fetchStatuses(username)
	    if err != nil {
	        fmt.Printf("Fehler beim Abrufen der Daten für Benutzer '%s': %v\n", username, err)
	        continue
	    }

	    processAndPrintTripDetails(username, data)
    }
}

func main() {
	go func() {
	    for {
	        updateMetrics()
	        time.Sleep(5 * time.Minute) // Aktualisierung alle 5 Minuten
	    }
    }()

	http.Handle("/metrics", promhttp.Handler())
	fmt.Println("Server läuft auf Port 8080...")
	http.ListenAndServe(":8080", nil)
}
