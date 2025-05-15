package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

// PRURL represents a PR URL with its metadata
type PRURL struct {
	URL string
}

// CSVFormat represents the detected format of the CSV file
type CSVFormat struct {
	URLColumn      int // -1 if not found
	OwnerColumn    int // -1 if not found
	RepoColumn     int // -1 if not found
	PRNumberColumn int // -1 if not found
}

// detectCSVFormat analyzes the CSV headers to determine which columns contain relevant information
func detectCSVFormat(headers []string) CSVFormat {
	format := CSVFormat{
		URLColumn:      -1,
		OwnerColumn:    -1,
		RepoColumn:     -1,
		PRNumberColumn: -1,
	}

	for i, header := range headers {
		header = strings.ToLower(strings.TrimSpace(header))
		switch header {
		case "url", "pr url", "pull request url":
			format.URLColumn = i
		case "owner", "repository owner", "repo owner":
			format.OwnerColumn = i
		case "repo", "repository", "repo name":
			format.RepoColumn = i
		case "pr", "pr number", "pull request", "pull request number":
			format.PRNumberColumn = i
		}
	}

	return format
}

// buildGitHubURL constructs a GitHub PR URL from owner, repo, and PR number
func buildGitHubURL(owner, repo, prNumber string) string {
	return fmt.Sprintf("https://github.com/%s/%s/pull/%s", owner, repo, prNumber)
}

// ParsePRURLsFromCSV reads a CSV file and returns a slice of PR URLs
// The function detects the CSV format by analyzing headers and can handle:
// 1. A direct URL column
// 2. Separate owner, repo, and PR number columns
func ParsePRURLsFromCSV(csvFile string) ([]PRURL, error) {
	file, err := os.Open(csvFile)
	if err != nil {
		return nil, fmt.Errorf("error opening CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV file: %v", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file must have at least a header row and one data row")
	}

	// Detect CSV format from headers
	format := detectCSVFormat(records[0])

	// Validate that we have either a URL column or the necessary columns to build a URL
	if format.URLColumn == -1 && (format.OwnerColumn == -1 || format.RepoColumn == -1 || format.PRNumberColumn == -1) {
		return nil, fmt.Errorf("CSV must have either a URL column or owner, repo, and PR number columns")
	}

	var prURLs []PRURL
	// Process data rows (skip header)
	for i, record := range records[1:] {
		var url string
		if format.URLColumn != -1 {
			// Use direct URL if available
			if format.URLColumn >= len(record) {
				fmt.Printf("Warning: Row %d is missing URL column, skipping\n", i+2)
				continue
			}
			url = strings.TrimSpace(record[format.URLColumn])
		} else {
			// Build URL from components
			if format.OwnerColumn >= len(record) || format.RepoColumn >= len(record) || format.PRNumberColumn >= len(record) {
				fmt.Printf("Warning: Row %d is missing required columns, skipping\n", i+2)
				continue
			}
			owner := strings.TrimSpace(record[format.OwnerColumn])
			repo := strings.TrimSpace(record[format.RepoColumn])
			prNumber := strings.TrimSpace(record[format.PRNumberColumn])
			url = buildGitHubURL(owner, repo, prNumber)
		}

		if url == "" {
			fmt.Printf("Warning: Row %d has empty URL, skipping\n", i+2)
			continue
		}

		prURLs = append(prURLs, PRURL{URL: url})
	}

	return prURLs, nil
}
