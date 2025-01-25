package internal

const dir = "data"

// ComputeScores runs the full scoring process for all (company, year) keys
//func ComputeScores(
//	ctx context.Context,ComputeScores,
//	cfg *c.Config,
//	datasets map[string]map[CompanyYearKey]map[string]float64,
//) map[CompanyYearKey]map[string]float64 {
//	tracer := otel.Tracer("score-app")
//	ctx, span := tracer.Start(ctx, "computeScores")
//	defer span.End()
//
//	allKeys := getAllDataCompanyKeys(datasets)
//
//	// Instead of a serial loop, call parallelComputeScores
//	finalResults := parallelComputeScores(ctx, allKeys, cfg, datasets, 4) // 4 workers, for example
//
//	return finalResults
//}

//func CalculateScore(ctx context.Context, configFileName string) (*c.Config, map[CompanyYearKey]map[string]float64, error) {
//	// 1. Load the scoring config
//	scoreConfig, err := c.InitScoreConfig(configFileName)
//	if err != nil {
//		return nil, nil, fmt.Errorf("error initializing score config: %w", err)
//	}
//	log.Printf("Loaded config: %s\n", scoreConfig.Name)
//
//	// 2. Load all CSVs
//	disclosureData, err := loadDatasetCSV(disclosurePath)
//	if err != nil {
//		return nil, nil, fmt.Errorf("error loading disclosure: %w", err)
//	}
//	wasteData, err := loadDatasetCSV(wastePath)
//	if err != nil {
//		return nil, nil, fmt.Errorf("error loading waste: %w", err)
//	}
//	emissionsData, err := loadDatasetCSV(emissionPath)
//	if err != nil {
//		return nil, nil, fmt.Errorf("error loading emissions: %w", err)
//	}
//
//	// 3. Combine into a single map
//	datasets := map[string]map[CompanyYearKey]map[string]float64{
//		"disclosure": disclosureData,
//		"waste":      wasteData,
//		"emissions":  emissionsData,
//	}
//
//	// 4. Compute the scores in a serial fashion
//	scoredResults := ComputeScores(ctx, scoreConfig, datasets)
//
//	return scoreConfig, scoredResults, nil
//}
//
//func ComputeScores(
//	ctx context.Context,
//	cfg *c.Config,
//	datasets map[string]map[CompanyYearKey]map[string]float64,
//) map[CompanyYearKey]map[string]float64 {
//	allKeys := getAllDataCompanyKeys(datasets)
//
//	finalResults := make(map[CompanyYearKey]map[string]float64, len(allKeys))
//	for _, cy := range allKeys {
//		metricValues := computeScoresForKey(ctx, cy, cfg, datasets)
//		finalResults[cy] = metricValues
//	}
//	return finalResults
//}
