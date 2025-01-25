package internal

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel"

	c "esgbook-software-engineer-technical-test-2024/config"
)

type CompanyYearKey struct {
	CompanyID string
	Year      int
}

type rowData struct {
	date    time.Time
	numeric map[string]float64
}

// just to test the json file
type rawJSONRow struct {
	CompanyID string   `json:"company_id"`
	DateStr   string   `json:"date"`
	Dis1      *float64 `json:"dis_1,omitempty"` // use pointer so we can detect "null" vs. 0
	Dis2      *float64 `json:"dis_2,omitempty"`
	Dis3      *float64 `json:"dis_3,omitempty"`
	Dis4      *float64 `json:"dis_4,omitempty"`
	// Add more fields if needed
}

type DisclosureData map[CompanyYearKey]map[string]float64

func indexOf(slice []string, target string) int {
	for i, s := range slice {
		if s == target {
			return i
		}
	}
	return -1
}

func getAllDataCompanyKeys(datasets map[string]map[CompanyYearKey]map[string]float64) []CompanyYearKey {
	unique := make(map[CompanyYearKey]bool)
	for _, ds := range datasets {
		for key := range ds {
			unique[key] = true
		}
	}
	out := make([]CompanyYearKey, 0, len(unique))
	for k := range unique {
		out = append(out, k)
	}
	return out
}

func evaluateMetric(
	ctx context.Context,
	metric c.Metric,
	key CompanyYearKey,
//results map[CompanyYearKey]map[string]float64,
	results map[string]float64,
	datasets map[string]map[CompanyYearKey]map[string]float64,
) (float64, bool) {
	// if we've already computed metric, return it
	if val, ok := results[metric.Name]; ok {
		return val, false
	}

	opFn, ok := operations[metric.Operation.Type]
	if !ok {
		log.Printf("Unknown operation: %s", metric.Operation.Type)
		return 0, true
	}

	val, isNull, err := opFn(ctx, metric.Operation.Parameters, key, results, datasets)
	if err != nil {
		// You might decide an error means “null,” or handle differently
		log.Printf("Error in operation %s: %v", metric.Operation.Type, err)
		return 0, true
	}

	if !isNull {
		results[metric.Name] = val
	}

	return val, isNull
}

func getValue(
	source string,
	key CompanyYearKey,
//results map[CompanyYearKey]map[string]float64,
	results map[string]float64,
	datasets map[string]map[CompanyYearKey]map[string]float64,
) (float64, bool) {
	// Check for self-reference
	if strings.HasPrefix(source, "self.") {
		metricName := strings.TrimPrefix(source, "self.")
		val, ok := results[metricName]
		if !ok {
			// We haven't computed it yet -> we need to do so dynamically if possible
			// Or consider your approach to ordering the metrics in the config (topological).
			// A simple approach is to do a second pass or a recursion here.
			// For now, let's treat it as "null" if not precomputed:
			return 0, true
		}
		return val, false
	}

	// Otherwise "datasetName.metricName"
	parts := strings.Split(source, ".")
	if len(parts) != 2 {
		return 0, true // invalid format => null
	}
	datasetName := parts[0]
	metricKey := parts[1]

	ds, ok := datasets[datasetName]
	if !ok {
		// unknown dataset => null
		return 0, true
	}
	row, ok := ds[key]
	if !ok {
		// no row => null
		return 0, true
	}
	val, ok := row[metricKey]
	if !ok {
		return 0, true
	}
	return val, false
}

// We'll call this function in computeScores or a separate step.
func parallelComputeScores(
	ctx context.Context,
	allKeys []CompanyYearKey,
	scoreConfig *c.Config,
	datasets map[string]map[CompanyYearKey]map[string]float64,
	numWorkers int,
) map[CompanyYearKey]map[string]float64 {
	// 1) Create job + result channels
	jobs := make(chan CompanyYearKey, len(allKeys))
	results := make(chan keyResult, len(allKeys))

	// 2) Spawn worker goroutines
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for key := range jobs {
				// Compute the metrics for this (company, year)
				metricsMap := computeScoresForKey(ctx, key, scoreConfig, datasets)
				results <- keyResult{
					Key:    key,
					Result: metricsMap,
				}
			}
		}()
	}

	// 3) Send all jobs to the workers
	for _, key := range allKeys {
		jobs <- key
	}
	close(jobs)

	// 4) Wait for all workers to finish, then close results
	go func() {
		wg.Wait()
		close(results)
	}()

	// 5) Collect results
	finalResults := make(map[CompanyYearKey]map[string]float64, len(allKeys))
	for kr := range results {
		finalResults[kr.Key] = kr.Result
	}

	return finalResults
}

// A simple struct to hold each worker's output
type keyResult struct {
	Key    CompanyYearKey
	Result map[string]float64
}

func computeScoresForKey(
	ctx context.Context,
	key CompanyYearKey,
	scoreConfig *c.Config,
	datasets map[string]map[CompanyYearKey]map[string]float64,
) map[string]float64 {
	metricResults := make(map[string]float64)

	// Evaluate each metric in the config
	for _, metric := range scoreConfig.Metrics {
		val, isNull := evaluateMetric(ctx, metric, key, metricResults, datasets)
		if !isNull {
			// store this metric's final value under its name
			metricResults[metric.Name] = val
		}
	}
	return metricResults
}

func (s *DataLoaderService) LoadAllData(
	ctx context.Context,
	dataDir string,
) (map[string]map[CompanyYearKey]map[string]float64, error) {

	files, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read data directory %s: %w", dataDir, err)
	}

	combined := make(map[string]map[CompanyYearKey]map[string]float64)

	for _, f := range files {
		if f.IsDir() {
			continue // skip subdirectories
		}

		fullPath := filepath.Join(dataDir, f.Name())
		ext := filepath.Ext(f.Name()) // e.g. ".csv"

		loader, ok := s.registry.GetLoader(ext)
		if !ok {
			log.Printf("[WARN] Skipping file with unsupported extension %q: %s", ext, f.Name())
			continue
		}

		ds, err := loader.LoadData(ctx, fullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load data from %s: %w", f.Name(), err)
		}

		// e.g. "waste_data.csv" => datasetName = "waste_data"
		datasetName := strings.TrimSuffix(f.Name(), ext)
		combined[datasetName] = ds
	}

	return combined, nil
}

func CalculateScore(
	ctx context.Context,
	configFileName string,
	dataService *DataLoaderService,
) (*c.Config, map[CompanyYearKey]map[string]float64, error) {

	// Start a tracing span
	tracer := otel.Tracer("score-app")
	ctx, span := tracer.Start(ctx, "CalculateScoreApp")
	defer span.End()

	// 1) Load the scoring config (score_1.yaml)
	scoreConfig, err := c.InitScoreConfig(configFileName)
	if err != nil {
		return nil, nil, fmt.Errorf("error initializing score config: %w", err)
	}
	log.Printf("Loaded config: %s\n", scoreConfig.Name)

	// 2) Load all CSVs (or other files) from "data/" using the injected service
	combined, err := dataService.LoadAllData(ctx, dir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load data from folder: %w", err)
	}

	// 3) Map them to expected dataset names. For example, if you expect "disclosureData",
	//    "wasteData", "emissionsData" as file names:
	datasets := map[string]map[CompanyYearKey]map[string]float64{
		"disclosure": combined["disclosure_data"], // "disclosure_data.csv" => "disclosure_data"
		"waste":      combined["waste_data"],
		"emissions":  combined["emissions_data"],
	}

	// 4) Parallel compute scores
	allKeys := getAllDataCompanyKeys(datasets)

	sort.Slice(allKeys, func(i, j int) bool {
		if allKeys[i].CompanyID == allKeys[j].CompanyID {
			return allKeys[i].Year < allKeys[j].Year
		}
		return allKeys[i].CompanyID < allKeys[j].CompanyID
	})

	scoredResults := parallelComputeScores(ctx, allKeys, scoreConfig, datasets, 4)

	return scoreConfig, scoredResults, nil
}
