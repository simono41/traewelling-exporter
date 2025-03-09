# Träwelling Prometheus Exporter

## Beschreibung

Dieser Prometheus Exporter sammelt Daten von der Träwelling API und stellt sie als Prometheus-Metriken bereit. Er bietet detaillierte Informationen über aktuelle Züge, Gesamtdistanz, Gesamtdauer und Gesamtpunkte für jeden Benutzer.

## Funktionen

- Abfrage der Träwelling API für aktuelle Zugdaten und Benutzerdetails
- Bereitstellung von Prometheus-Metriken für:
  - Aktuelle Zugstatus (aktiv/inaktiv)
  - Gesamte Zugstrecke pro Benutzer
  - Gesamte Zugdauer pro Benutzer
  - Gesamtpunkte pro Benutzer
- Unterstützung für mehrere Benutzernamen
- Regelmäßige Aktualisierung der Daten (alle 5 Minuten)

## Voraussetzungen

- Go 1.20 oder höher
- Docker und Docker Compose (für containerisierte Ausführung)
- Gültiger Träwelling API-Token
- Benutzernamen der zu überwachenden Nutzer

## Installation

1. Klone das Repository:
   ```
   git clone https://github.com/yourusername/traewelling-prometheus-exporter.git
   cd traewelling-prometheus-exporter
   ```

2. Installiere die erforderlichen Go-Pakete:
   ```
   go get github.com/prometheus/client_golang/prometheus
   go get github.com/prometheus/client_golang/prometheus/promauto
   go get github.com/prometheus/client_golang/prometheus/promhttp
   ```

3. Erstelle eine `.env` Datei im Projektverzeichnis und füge deine Umgebungsvariablen hinzu:
   ```
   TRAEWELLING_TOKEN=your_token
   TRAEWELLING_USERNAMES=
   ```

## Ausführung

### Lokale Ausführung

1. Baue die Anwendung:
   ```
   go build -o traewelling-exporter
   ```

2. Führe die Anwendung aus:
   ```
   ./traewelling-exporter
   ```

### Mit Docker

1. Baue das Docker-Image:
   ```
   docker-compose build
   ```

2. Starte den Container:
   ```
   docker-compose up -d
   ```

## Prometheus Konfiguration

Füge folgende Job-Konfiguration zu deiner `prometheus.yml` hinzu:
```
scrape_configs:
  - job_name: 'traewelling'
    static_configs:
      - targets: ['localhost:8080']
```

## Metriken

- `traewelling_current_train_statuses`: Aktueller Status eines Zuges (1 = aktiv, 0 = inaktiv)
- `traewelling_total_train_distance_km`: Gesamte Zugstrecke eines Benutzers in Kilometern
- `traewelling_total_train_duration_minutes`: Gesamte Zugdauer eines Benutzers in Minuten
- `traewelling_total_points`: Gesamtpunkte eines Benutzers

## Beitragen

Beiträge sind willkommen! Bitte erstelle ein Issue oder einen Pull Request für Verbesserungsvorschläge oder Fehlerbehebungen.

## Lizenz

[MIT License](LICENSE)

## Kontakt

Bei Fragen oder Problemen erstelle bitte ein GitHub Issue oder kontaktiere [admin@brothertec.eu](mailto:admin@brothertec.eu).
