package internal

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"go.opentelemetry.io/otel"
)

const wastePath = "data/waste_data.csv"
const emissionPath = "data/emissions_data.csv"
const disclosurePath = "data/disclosure_data.csv"

func CalculateScoreHandler(ctx context.Context, configFileName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received request to calculate scores")

		// Start a span for tracing, using the request context
		tracer := otel.Tracer("score-app")
		childCtx, span := tracer.Start(r.Context(), "computeScores")
		defer span.End()

		lr := NewLoaderRegistry()
		dataService := NewDataLoaderService(lr)

		// 1) Calculate the score using your business logic function
		scoreConfig, scoredResults, err := CalculateScore(childCtx, configFileName, dataService)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return
		}

		// 2) Prepare to send results as CSV
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", `attachment; filename="scores.csv"`)
		csvWriter := csv.NewWriter(w)
		defer csvWriter.Flush()

		// 3) Write Header Row: "company", "year", plus each metric
		header := []string{"company", "year"}
		for _, metric := range scoreConfig.Metrics {
			header = append(header, metric.Name)
		}
		err = csvWriter.Write(header)
		if err != nil {
			log.Printf("Failed to write CSV header: %v", err)
			http.Error(w, "Failed to write CSV header", http.StatusInternalServerError)
			return
		}

		// 4) Write Data Rows
		for cy, metricsMap := range scoredResults {
			row := []string{
				cy.CompanyID,
				strconv.Itoa(cy.Year),
			}
			for _, metric := range scoreConfig.Metrics {
				if val, ok := metricsMap[metric.Name]; ok {
					row = append(row, fmt.Sprintf("%.2f", val))
				} else {
					row = append(row, "") // or "NULL"
				}
			}
			if err := csvWriter.Write(row); err != nil {
				log.Printf("Failed to write CSV row: %v", err)
				http.Error(w, "Failed to write CSV row", http.StatusInternalServerError)
				return
			}
		}
	}
}

func HealthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	if err := isServiceHealthy(); err != nil {
		log.Printf("Health check failed: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Print("Status ok!")
	w.WriteHeader(http.StatusOK)
}

func isServiceHealthy() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	select {
	case <-time.After(2 * time.Second):
		return nil
	case <-ctx.Done():
		return fmt.Errorf("timeout while checking service health")
	}
}
