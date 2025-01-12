package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"go-crux-microservice/utils"
)

const (
	cruxAPIURLBase = "https://chromeuxreport.googleapis.com/v1/records:queryRecord"
)

// Load environment variables
func getEnvVar(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

type CruxRequest struct {
	Origin  string   `json:"origin"`
	Metrics []string `json:"metrics,omitempty"`
}

type CruxMetric struct {
	Name       string           `json:"name"`
	Percentile float64          `json:"p75"`
	Histogram  []HistogramEntry `json:"histogram,omitempty"`
}

type HistogramEntry struct {
	Start   float64 `json:"start"`
	End     float64 `json:"end,omitempty"`
	Density float64 `json:"density"`
}

func FetchCruxMetrics(url string, apiKey string, desiredMetrics []string) ([]CruxMetric, error) {
	requestBody, err := json.Marshal(CruxRequest{Origin: url, Metrics: desiredMetrics})
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error marshalling request body: %v", err))
	}

	// Debug Step: Print the URL as part of the origin in the request
	fmt.Println("Request body:", string(requestBody))

	urlWithKey := fmt.Sprintf("%s?key=%s", cruxAPIURLBase, apiKey)
	req, err := http.NewRequest("POST", urlWithKey, bytes.NewReader(requestBody))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error creating request: %v", err))
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error making request: %v", err))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error reading response body: %v", err))
	}

	// Debug Step: Logging full response of the API
	fmt.Println("Response status code:", resp.StatusCode)
	fmt.Println("Response body:", string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-200 status code: %d\nResponse body: %s\n", resp.StatusCode, string(body))
	}

	var cruxResponse map[string]interface{}
	err = json.Unmarshal(body, &cruxResponse)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling response: %v\nRaw response body: %s\n", err, string(body))
	}

	metrics := extractMetricsFromResponse(cruxResponse, desiredMetrics)
	return metrics, nil
}

func extractMetricsFromResponse(response map[string]interface{}, desiredMetrics []string) []CruxMetric {
	record, _ := response["record"].(map[string]interface{})
	metricsData, _ := record["metrics"].(map[string]interface{})
	var metrics []CruxMetric
	for metricName, metricData := range metricsData {
		// Skip metrics not in desired list
		if !utils.StringInSlice(desiredMetrics, metricName) {
			continue
		}
		metricMap, _ := metricData.(map[string]interface{})
		percentile, _ := metricMap["percentiles"].(map[string]interface{})

		// Safely extract p75 and handle potential issues if p75 is not found
		p75Interface, ok := percentile["p75"].(interface{})
		if !ok {
			fmt.Printf("Warning: p75 for metric '%s' is not found.\n", metricName)
			continue
		}

		var p75 float64

		//p75 for CLS for a reason is String, adding a switch to always convert it to float
		switch v := p75Interface.(type) {
		case float64:
			p75 = v
		case string:
			p75, _ = strconv.ParseFloat(v, 64)
		default:
			fmt.Printf("Warning: Unexpected p75 type for metric '%s': %T\n", metricName, v)
			continue
		}

		//Debug Step: Print type of p75
		fmt.Printf("Type of p75: %T\n", p75)

		var histogram []HistogramEntry
		histogramData, ok := metricMap["histogram"].([]interface{})
		if ok {
			for _, bucket := range histogramData {
				bucketMap, _ := bucket.(map[string]interface{})
				start, _ := bucketMap["start"].(float64)
				end, _ := bucketMap["end"].(float64)
				if !ok {
					end = 0
				}
				density, _ := bucketMap["density"].(float64)
				histogram = append(histogram, HistogramEntry{Start: start, End: end, Density: density})
			}
		}
		// Store metrics, p75 should be a float64 type now
		metrics = append(metrics, CruxMetric{Name: metricName, Percentile: p75, Histogram: histogram})
	}
	return metrics
}

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	cruxAPIKey := getEnvVar("CRUX_API_KEY", "YOUR_DEFAULT_API_KEY")
	dbHost := getEnvVar("DB_HOST", "localhost")
	dbPort := getEnvVar("DB_PORT", "5432")
	dbUser := getEnvVar("DB_USER", "your_user")
	dbPassword := getEnvVar("DB_PASSWORD", "your_password")
	dbName := getEnvVar("DB_NAME", "crux")

	dbURL := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName)
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return
	}
	defer db.Close()

	desiredMetrics := []string{"largest_contentful_paint", "cumulative_layout_shift", "interaction_to_next_paint"}
	urls, _ := utils.ReadURLsFromFile("urls.json")

	for _, url := range urls {
		metrics, _ := FetchCruxMetrics(url, cruxAPIKey, desiredMetrics)

		for _, metric := range metrics {
			// Extract densities from histogram (3 buckets)
			var good, needsImprovement, poor float64
			if len(metric.Histogram) >= 3 {
				good = metric.Histogram[0].Density
				needsImprovement = metric.Histogram[1].Density
				poor = metric.Histogram[2].Density
			}

			// Insert data into the database
			_, err := db.Exec(
				"INSERT INTO crux_metrics (url, metric, score, good, needs_improvement, poor) VALUES ($1, $2, $3, $4, $5, $6)",
				url, metric.Name, metric.Percentile, good, needsImprovement, poor)
			if err != nil {
				fmt.Println("Error inserting data:", err)
			}
		}
	}
}
