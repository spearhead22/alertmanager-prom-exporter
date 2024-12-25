# Alertmanager Exporter

Simple Prometheus exporter for Alertmanager that exposes metrics about current alerts.

## Metrics

| Metric | Description |
|--------|-------------|
| `alertmanager_alerts_total` | Total number of alerts by status and name |
| `alertmanager_scrape_errors_total` | Total number of scrape errors |

## Prerequisites

- Go 1.19 or higher
- Access to an Alertmanager instance

## Installation

1. Clone the repository:
```bash
git clone <your-repo-url>
cd alertmanager-exporter
```

2. Install dependencies:
```bash
go mod init alertmanager-exporter
go mod tidy
```

## Configuration

The exporter can be configured using environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `ALERTMANAGER_URL` | URL of Alertmanager API | `http://localhost:9093/api/v1/alerts` |
| `EXPORTER_PORT` | Port for the exporter | `8080` |
| `UPDATE_INTERVAL` | Metrics update interval in seconds | `15` |

## Running

### Local Development

1. Run with default settings:
```bash
go run main.go
```

2. Run with custom settings:
```bash
export ALERTMANAGER_URL="http://your-alertmanager:9093/api/v1/alerts"
export EXPORTER_PORT="9090"
export UPDATE_INTERVAL="30"
go run main.go
```

### Docker

1. Build the image:
```bash
docker build -t alertmanager-exporter .
```

2. Run the container:
```bash
docker run -d \
  -e ALERTMANAGER_URL="http://alertmanager:9093/api/v1/alerts" \
  -e EXPORTER_PORT="9090" \
  -e UPDATE_INTERVAL="30" \
  -p 9090:9090 \
  alertmanager-exporter
```

## Testing

### Setting up local Alertmanager

1. Download and run Alertmanager:
```bash
wget https://github.com/prometheus/alertmanager/releases/download/v0.26.0/alertmanager-0.26.0.linux-amd64.tar.gz
tar xvfz alertmanager-0.26.0.linux-amd64.tar.gz
cd alertmanager-0.26.0.linux-amd64/
```

2. Create config file `alertmanager.yml`:
```yaml
global:
  resolve_timeout: 5m

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'email-notifications'

receivers:
- name: 'email-notifications'
  email_configs:
  - to: 'your-email@example.com'
    from: 'alertmanager@example.com'
    smarthost: 'localhost:25'
    require_tls: false
```

3. Start Alertmanager:
```bash
./alertmanager --config.file=alertmanager.yml --web.listen-address=:9093
```

### Sending test alerts

Create file `alerts.json`:
```json
[
  {
    "labels": {
      "alertname": "HighCPUUsage",
      "instance": "test-host-1",
      "severity": "warning"
    },
    "annotations": {
      "description": "CPU usage is above 80%",
      "summary": "High CPU usage detected"
    },
    "generatorURL": "http://prometheus.example.com",
    "startsAt": "2024-12-25T10:00:00Z"
  }
]
```

Send alert:
```bash
curl -XPOST -H "Content-Type: application/json" localhost:9093/api/v1/alerts -d @alerts.json
```

## Checking the results

1. Check exporter metrics:
```bash
curl localhost:8080/metrics
```

2. Access Alertmanager UI:
```
http://localhost:9093
```

## Build

To build the binary:
```bash
go build -o alertmanager-exporter main.go
```

## Docker build

Create a Dockerfile:
```dockerfile
FROM golang:1.19-alpine

WORKDIR /app
COPY . .

RUN go mod download
RUN go build -o alertmanager-exporter

FROM alpine:latest
WORKDIR /app
COPY --from=0 /app/alertmanager-exporter .

EXPOSE 8080
CMD ["./alertmanager-exporter"]
```

Build and run:
```bash
docker build -t alertmanager-exporter .
docker run -p 8080:8080 alertmanager-exporter
```