package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// openPRsFromCSV opens PR URLs from a CSV file in the default browser
func openPRsFromCSV(csvFile string) error {
	// Read URLs from CSV
	file, err := os.Open(csvFile)
	if err != nil {
		return fmt.Errorf("error opening CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("error reading CSV file: %v", err)
	}

	// Skip header row
	for i, record := range records[1:] {
		if len(record) < 4 {
			continue
		}
		url := record[3]

		fmt.Printf("\nOpening PR %d/%d: %s\n", i+1, len(records)-1, url)
		if err := exec.Command("open", url).Start(); err != nil {
			fmt.Printf("Error opening URL: %v\n", err)
			continue
		}
		time.Sleep(time.Second)
	}

	return nil
}
