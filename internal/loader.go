package internal

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
)

// DataLoader interface: for reading data from a specific file/path.
type DataLoader interface {
	LoadData(ctx context.Context, path string) (map[CompanyYearKey]map[string]float64, error)
}

// CSVLoader A simple CSV loader example.
type CSVLoader struct{}

func (CSVLoader) LoadData(ctx context.Context, path string) (map[CompanyYearKey]map[string]float64, error) {
	return loadDatasetCSV(path) // reuse your existing CSV parsing function
}

// JSONLoader Temporary files to test the JSON Import only
type JSONLoader struct{}

func (JSONLoader) LoadData(ctx context.Context, path string) (map[CompanyYearKey]map[string]float64, error) {
	return loadJSONDataset(path)
}

func loadDatasetCSV(filename string) (map[CompanyYearKey]map[string]float64, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)

	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read headers: %v", err)
	}

	idxCompany := indexOf(headers, "company_id")
	idxDate := indexOf(headers, "date")
	if idxCompany == -1 || idxDate == -1 {
		return nil, fmt.Errorf("missing required columns (company_id, date)")
	}

	// Use rowData to store the 'latest' row (by full date) for each (company, year)
	data := make(map[CompanyYearKey]rowData)

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		companyID := row[idxCompany]

		// 1) Parse the date into a time.Time
		parsedTime, err := ParseDateOrYear(row[idxDate])
		if err != nil {
			// Could skip or handle the error
			log.Printf("Skipping row for %s due to date parse error: %v", companyID, err)
			continue
		}

		// 2) The year is just parsedTime.Year()
		yearInt := parsedTime.Year()

		// The key for our "final" data
		key := CompanyYearKey{
			CompanyID: companyID,
			Year:      yearInt,
		}

		// Gather numeric columns from the row
		numericVals := map[string]float64{}
		for i, colName := range headers {
			if i == idxCompany || i == idxDate {
				continue
			}
			valStr := row[i]
			if valStr == "" {
				continue
			}
			if v, err := strconv.ParseFloat(valStr, 64); err == nil {
				numericVals[colName] = v
			}
		}

		// 3) Check if we have an existing entry for (company, year). If not, store it.
		if existing, ok := data[key]; !ok {
			data[key] = rowData{
				date:    parsedTime,
				numeric: numericVals,
			}
		} else {
			// 4) If we do have an existing entry, only overwrite if the new date is "later"
			//    (based on your chosen definition of "latest").
			if parsedTime.After(existing.date) {
				data[key] = rowData{
					date:    parsedTime,
					numeric: numericVals,
				}
			}
		}
	}

	// 5) Flatten rowData => final map for each (CompanyYearKey)
	result := make(map[CompanyYearKey]map[string]float64)
	for key, rd := range data {
		result[key] = rd.numeric
	}

	return result, nil
}

func loadJSONDataset(filename string) (map[CompanyYearKey]map[string]float64, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	bytes, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %w", filename, err)
	}

	// Unmarshal into a slice of rawJSONRow
	var rows []rawJSONRow
	if err := json.Unmarshal(bytes, &rows); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON from %s: %w", filename, err)
	}

	// We'll store rowData with date/time for each (company, year)
	data := make(map[CompanyYearKey]rowData)

	for _, r := range rows {
		// Parse the date -> time.Time
		parsedTime, err := ParseDateOrYear(r.DateStr)
		if err != nil {
			log.Printf("Skipping row for company=%s due to invalid date %q: %v", r.CompanyID, r.DateStr, err)
			continue
		}

		// Convert to year
		yearInt := parsedTime.Year()

		key := CompanyYearKey{
			CompanyID: r.CompanyID,
			Year:      yearInt,
		}

		// Gather numeric columns
		numericVals := make(map[string]float64)
		if r.Dis1 != nil {
			numericVals["dis_1"] = *r.Dis1
		}
		if r.Dis2 != nil {
			numericVals["dis_2"] = *r.Dis2
		}
		if r.Dis3 != nil {
			numericVals["dis_3"] = *r.Dis3
		}
		if r.Dis4 != nil {
			numericVals["dis_4"] = *r.Dis4
		}
		// If you had more columns (dis_5, etc.), add them likewise.

		// Check if we have an existing row for this (company, year).
		if existing, ok := data[key]; !ok {
			// If not, store it
			data[key] = rowData{
				date:    parsedTime,
				numeric: numericVals,
			}
		} else {
			// If we do have an entry, only overwrite if parsedTime is "later"
			if parsedTime.After(existing.date) {
				data[key] = rowData{
					date:    parsedTime,
					numeric: numericVals,
				}
			}
		}
	}

	// Flatten rowData => final map for each (CompanyYearKey)
	result := make(map[CompanyYearKey]map[string]float64, len(data))
	for k, rd := range data {
		result[k] = rd.numeric
	}

	return result, nil
}
