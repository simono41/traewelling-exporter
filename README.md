# Träwelling Prometheus Exporter

## Beschreibung

Dieser Prometheus Exporter sammelt Daten von der Träwelling API und stellt sie als Prometheus-Metriken zur Verfügung. Er bietet detaillierte Informationen über Fahrten, einschließlich Gesamtfahrzeit, Gesamtpunkte und Gesamtentfernung für jeden Benutzer.

## Funktionen

- Abfrage der Träwelling API für aktuelle Fahrtdaten
- Bereitstellung von Prometheus-Metriken für:
  - Gesamtfahrzeit pro Benutzer
  - Gesamtpunkte pro Benutzer
  - Gesamtentfernung pro Benutzer
- Konsolenausgabe detaillierter Fahrtinformationen
- Regelmäßige Aktualisierung der Daten (alle 5 Minuten)

## Voraussetzungen

- Go 1.15 oder höher
- Docker und Docker Compose (für containerisierte Ausführung)
- Gültiger Träwelling API-Token

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

3. Erstelle eine `.env` Datei im Projektverzeichnis und füge deinen Träwelling API-Token hinzu:
   ```
   TRAEWELLING_TOKEN=your_token_here
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

## Verwendung

Nach dem Start ist der Exporter unter `http://localhost:8080/metrics` erreichbar. Prometheus kann so konfiguriert werden, dass es diese Endpunkt abfragt.

## Prometheus Konfiguration

Füge folgende Job-Konfiguration zu deiner `prometheus.yml` hinzu:

```
scrape_configs:
  - job_name: 'traewelling'
    static_configs:
      - targets: ['localhost:8080']
```

## Metriken

- `traewelling_total_trip_duration_minutes`: Gesamtfahrzeit eines Benutzers in Minuten
- `traewelling_total_trip_points`: Gesamtpunkte eines Benutzers
- `traewelling_total_trip_distance_km`: Gesamtentfernung eines Benutzers in Kilometern

## Beitragen

Beiträge sind willkommen! Bitte erstelle ein Issue oder einen Pull Request für Verbesserungsvorschläge oder Fehlerbehebungen.

## Lizenz

[MIT License](LICENSE)

## Kontakt

Bei Fragen oder Problemen erstelle bitte ein GitHub Issue oder kontaktiere [dein-kontakt@example.com](mailto:dein-kontakt@example.com).
