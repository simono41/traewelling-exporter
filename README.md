# Träwelling Prometheus Exporter

## Beschreibung

Dieser Prometheus Exporter sammelt Daten von der Träwelling API und stellt sie als Prometheus-Metriken zur Verfügung. Er bietet detaillierte Informationen über aktuelle Züge, einschließlich Zugtyp und Fahrtart (privat, geschäftlich, Pendelverkehr), sowie Gesamtstatistiken wie Gesamtdistanz, Gesamtdauer und Gesamtpunkte für jeden Benutzer.

## Funktionen

- Abfrage der Träwelling API für aktuelle Zugdaten und Benutzerdetails
- Bereitstellung von Prometheus-Metriken für:
  - Aktuelle Zugstatus (aktiv/inaktiv) mit Zugtyp und Fahrtart
  - Gesamte Zugstrecke eines Benutzers in Kilometern
  - Gesamte Zugdauer eines Benutzers in Minuten
  - Gesamtpunkte eines Benutzers
- Konsolenausgabe detaillierter Zuginformationen
- Regelmäßige Aktualisierung der Daten (alle 5 Minuten)

## Voraussetzungen

- Go 1.15 oder höher
- Docker und Docker Compose (für containerisierte Ausführung)
- Gültiger Träwelling API-Token
- Liste der Benutzernamen, die abgefragt werden sollen

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

3. Erstelle eine `.env` Datei im Projektverzeichnis und füge deinen Träwelling API-Token sowie die Liste der Benutzernamen hinzu:
   ```
   TRAEWELLING_TOKEN=your_token_here
   TRAEWELLING_USERNAMES=user1,user2,user3
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

- **`traewelling_current_train_statuses`**:
  - Labels:
    - `username`: Der Benutzername.
    - `line_name`: Die Linie des Zuges (z. B. S3).
    - `origin`: Startbahnhof.
    - `destination`: Zielbahnhof.
    - `train_type`: Typ des Zuges (z. B. "national", "regional", etc.).
    - `trip_type`: Art der Fahrt (z. B. "personal", "business", "commute").
  - Wert:
    - **`1`**: Der Zug ist aktiv.
    - **`0`**: Der Zug ist inaktiv.

- **`traewelling_total_train_distance_km`**: Gesamte Zugstrecke eines Benutzers in Kilometern.
- **`traewelling_total_train_duration_minutes`**: Gesamte Zugdauer eines Benutzers in Minuten.
- **`traewelling_total_points`**: Gesamtpunkte eines Benutzers.

## Beitragen

Beiträge sind willkommen! Bitte erstelle ein Issue oder einen Pull Request für Verbesserungsvorschläge oder Fehlerbehebungen.

## Lizenz

[MIT License](LICENSE)

## Kontakt

Bei Fragen oder Problemen erstelle bitte ein GitHub Issue oder kontaktiere [admin@brothertec.eu](mailto:admin@brothertec.eu).
