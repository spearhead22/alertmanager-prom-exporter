// main.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Alert struct {
	Status      AlertStatus       `json:"status"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	StartsAt    string            `json:"startsAt"`
}

type AlertStatus struct {
	State       string   `json:"state"`
	SilencedBy  []string `json:"silencedBy"`
	InhibitedBy []string `json:"inhibitedBy"`
}

var (
	alertsTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "alertmanager_alerts_total",
			Help: "Total number of alerts by status and name",
		},
		[]string{"alertname", "state", "instance"},
	)

	scrapeErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "alertmanager_scrape_errors_total",
			Help: "Total number of scrape errors",
		},
	)
)

func init() {
	prometheus.MustRegister(alertsTotal)
	prometheus.MustRegister(scrapeErrors)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func validateURL(url string) error {
	_, err := http.NewRequest("GET", url, nil)
	return err
}

func fetchAlerts(alertmanagerURL string) ([]Alert, error) {
	resp, err := http.Get(alertmanagerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch alerts: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var alerts []Alert
	if err := json.NewDecoder(resp.Body).Decode(&alerts); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	return alerts, nil
}

func updateMetrics(alertmanagerURL string) {
	alerts, err := fetchAlerts(alertmanagerURL)
	if err != nil {
		log.Printf("Error fetching alerts: %v\n", err)
		scrapeErrors.Inc()
		return
	}

	alertsTotal.Reset()

	for _, alert := range alerts {
		alertname, ok := alert.Labels["alertname"]
		if !ok {
			alertname = "unknown"
		}

		instance, ok := alert.Labels["instance"]
		if !ok {
			instance = "unknown"
		}

		alertsTotal.WithLabelValues(
			alertname,
			alert.Status.State,
			instance,
		).Inc()
	}
}

func main() {
	alertmanagerURL := getEnvOrDefault(
		"ALERTMANAGER_URL",
		"http://localhost:9093/api/v1/alerts",
	)

	exporterPort := getEnvOrDefault("EXPORTER_PORT", "8080")

	updateInterval := getEnvOrDefault("UPDATE_INTERVAL", "15")

	interval, err := time.ParseDuration(updateInterval + "s")
	if err != nil {
		log.Printf("Invalid update interval, using default 15s: %v", err)
		interval = 15 * time.Second
	}

	if err := validateURL(alertmanagerURL); err != nil {
		log.Fatalf("Invalid Alertmanager URL: %v", err)
	}

	log.Printf("Starting with configuration:")
	log.Printf("Alertmanager URL: %s", alertmanagerURL)
	log.Printf("Exporter Port: %s", exporterPort)
	log.Printf("Update Interval: %s", interval)

	go func() {
		for {
			updateMetrics(alertmanagerURL)
			time.Sleep(interval)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())

	log.Printf("Starting exporter on :%s/metrics", exporterPort)
	if err := http.ListenAndServe(":"+exporterPort, nil); err != nil {
		log.Fatalf("Error starting server: %v\n", err)
	}
}
