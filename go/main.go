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
		ID   int `json:"id"`
		User struct {
			DisplayName string `json:"displayName"`
			Username    string `json:"username"`
		} `json:"userDetails"`
		Train struct {
			Trip          int    `json:"trip"`
			Category      string `json:"category"`
			LineName      string `json:"lineName"`
			JourneyNumber int    `json:"journeyNumber"`
			Origin        struct {
				Name             string `json:"name"`
				DeparturePlanned string `json:"departurePlanned"`
				DepartureReal    string `json:"departureReal"`
			} `json:"origin"`
			Destination struct {
				Name           string `json:"name"`
				ArrivalPlanned string `json:"arrivalPlanned"`
				ArrivalReal    string `json:"arrivalReal"`
			} `json:"destination"`
			TripType int `json:"tripType"` // 0: personal, 1: business, 2: commute
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
	currentTrainStatuses = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "traewelling_current_train_statuses",
			Help: "Zeigt an, ob ein Zug aktiv ist (1 = aktiv, 0 = inaktiv)",
		},
		[]string{"username", "line_name", "origin", "destination", "train_type", "trip_type"},
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
	existingTrips = map[string]bool{}
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

func isTrainActive(departureTimeStr, arrivalTimeStr string) bool {
	if departureTimeStr == "" || arrivalTimeStr == "" {
		log.Printf("Unbekannter Fehler beim Parsen der Abfahrts oder Ankunftszeit!")
		return false
	}

	departureTime, err := time.Parse(time.RFC3339, departureTimeStr)
	if err != nil {
		log.Printf("Fehler beim Parsen der Abfahrtszeit: %v\n", err)
		return false
	}

	arrivalTime, err := time.Parse(time.RFC3339, arrivalTimeStr)
	if err != nil {
		log.Printf("Fehler beim Parsen der Ankunftszeit: %v\n", err)
		return false
	}

	now := time.Now()
	return now.After(departureTime) && now.Before(arrivalTime)
}

func getTripType(tripType int) string {
	switch tripType {
	case 0:
		return "personal"
	case 1:
		return "business"
	case 2:
		return "commute"
	default:
		return "unknown"
	}
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

	// Bereinige die existingTrips-Map
	for key := range existingTrips {
		found := false
		for _, trip := range statusData.Data {
			tripType := getTripType(trip.Train.TripType)
			tripKey := fmt.Sprintf("%s_%s_%s_%s_%s_%s", username, trip.Train.LineName, trip.Train.Origin.Name, trip.Train.Destination.Name, trip.Train.Category, tripType)
			if key == tripKey {
				found = true
				break
			}
		}
		if !found {
			delete(existingTrips, key)
		}
	}

	for _, trip := range statusData.Data {
		active := isTrainActive(trip.Train.Origin.DeparturePlanned, trip.Train.Destination.ArrivalPlanned)

		tripType := getTripType(trip.Train.TripType)

		key := fmt.Sprintf("%s_%s_%s_%s_%s_%s", username, trip.Train.LineName, trip.Train.Origin.Name, trip.Train.Destination.Name, trip.Train.Category, tripType)

		if existingTrips[key] {
			log.Printf("Fahrt %s für Benutzer '%s' bereits vorhanden, überspringe...\n", trip.Train.LineName, username)
			continue
		}

		currentTrainStatuses.WithLabelValues(
			username,
			trip.Train.LineName,
			trip.Train.Origin.Name,
			trip.Train.Destination.Name,
			trip.Train.Category,
			tripType,
		).Set(func() float64 {
			if active {
				return 1
			} else {
				return 0
			}
		}())

		existingTrips[key] = true

		if active {
			log.Printf("Aktive Verbindung für Benutzer '%s':\n", username)
			log.Printf("  Linie: %s\n", trip.Train.LineName)
			log.Printf("  Startbahnhof: %s\n", trip.Train.Origin.Name)
			log.Printf("    Geplante Abfahrt: %s\n", trip.Train.Origin.DeparturePlanned)
			log.Printf("    Tatsächliche Abfahrt: %s\n", trip.Train.Origin.DepartureReal)
			log.Printf("  Zielbahnhof: %s\n", trip.Train.Destination.Name)
			log.Printf("    Geplante Ankunft: %s\n", trip.Train.Destination.ArrivalPlanned)
			log.Printf("    Tatsächliche Ankunft: %s\n", trip.Train.Destination.ArrivalReal)
			log.Printf("  Zugtyp: %s\n", trip.Train.Category)
			log.Printf("  Fahrtart: %s\n", tripType)
		}
	}

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
			time.Sleep(5 * time.Minute) // Aktualisierung alle 5 Minuten
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Standardport
	}
	fmt.Printf("Server läuft auf Port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
