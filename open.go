package main

import (
	"fmt"
	"os/exec"
	"time"
)

// openPRsFromCSV opens PR URLs from a CSV file in the default browser
func openPRsFromCSV(csvFile string) error {
	prURLs, err := ParsePRURLsFromCSV(csvFile)
	if err != nil {
		return err
	}

	for i, pr := range prURLs {
		fmt.Printf("\nOpening PR %d/%d: %s\n", i+1, len(prURLs), pr.URL)
		if err := exec.Command("open", pr.URL).Start(); err != nil {
			fmt.Printf("Error opening URL: %v\n", err)
			continue
		}
		time.Sleep(time.Second)
	}

	return nil
}
