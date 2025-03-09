package main

import (
	"encoding/json"
	"fmt"
	"log"
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

type UserDetailsResponse struct {
	Data struct {
		ID            int    `json:"id"`
		Username      string `json:"username"`
		TrainDistance int    `json:"trainDistance"`
		TrainDuration int    `json:"trainDuration"`
		Points        int    `json:"points"`
	} `json:"data"`
}

var (
	currentConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "traewelling_current_connections",
			Help: "Anzahl der aktuellen Verbindungen pro Benutzer",
		},
		[]string{"username"},
	)
	totalTrainDistance = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "traewelling_total_train_distance_km",
			Help: "Gesamte Zugstrecke eines Benutzers in Kilometern",
		},
		[]string{"username"},
	)
	totalTrainDuration = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "traewelling_total_train_duration_minutes",
			Help: "Gesamte Zugdauer eines Benutzers in Minuten",
		},
		[]string{"username"},
	)
	totalPoints = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "traewelling_total_points",
			Help: "Gesamtpunkte eines Benutzers",
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

func fetchUserDetails(username string) (*UserDetailsResponse, error) {
	url := fmt.Sprintf("https://traewelling.de/api/v1/user/%s", username)
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Fehler beim Erstellen der Benutzerdetail Anfrage: %v", err)
	}

	token := os.Getenv("TRAEWELLING_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("TRAEWELLING_TOKEN ist nicht gesetzt")
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Fehler beim Senden der Benutzerdetail Anfrage für Benutzer '%s': %v", username, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Fehlerhafte Benutzerdetail Antwort für Benutzer '%s': %d", username, resp.StatusCode)
	}

	var apiResponse UserDetailsResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResponse)
	if err != nil {
		return nil, fmt.Errorf("Fehler beim Parsen der JSON-Daten der Benutzerdetails für Benutzer '%s': %v", username, err)
	}

	return &apiResponse, nil
}

func updateMetricsForUser(username string) {
	statusData, err := fetchStatuses(username)
	if err != nil {
		log.Printf("Fehler beim Abrufen der Statusdaten für Benutzer '%s': %v\n", username, err)
		return
	}

	userDetails, err := fetchUserDetails(username)
	if err != nil {
		log.Printf("Fehler beim Abrufen der Benutzerdetails für Benutzer '%s': %v\n", username, err)
		return
	}

	currentConnections.WithLabelValues(username).Set(float64(len(statusData.Data)))
	totalTrainDistance.WithLabelValues(username).Set(float64(userDetails.Data.TrainDistance) / 1000)
	totalTrainDuration.WithLabelValues(username).Set(float64(userDetails.Data.TrainDuration))
	totalPoints.WithLabelValues(username).Set(float64(userDetails.Data.Points))

	log.Printf("Aktualisierte Metriken für Benutzer '%s'\n", username)

}

func updateMetrics() {
	usernames := os.Getenv("TRAEWELLING_USERNAMES")
	if usernames == "" {
		log.Println("TRAEWELLING_USERNAMES ist nicht gesetzt")
		return
	}

	for _, username := range strings.Split(usernames, ",") {
		username = strings.TrimSpace(username)
		updateMetricsForUser(username)
	}
}

func main() {
	go func() {
		for {
			updateMetrics()
			time.Sleep(5 * time.Minute)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	fmt.Println("Server läuft auf Port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
