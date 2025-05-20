package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
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

// detectDelimiter tries to determine if the file uses tabs or commas as delimiters
func detectDelimiter(file *os.File) (rune, error) {
	// Read the first line
	reader := bufio.NewReader(file)
	firstLine, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return 0, fmt.Errorf("error reading first line: %v", err)
	}

	// Reset file position for subsequent reads
	if _, err := file.Seek(0, 0); err != nil {
		return 0, fmt.Errorf("error resetting file position: %v", err)
	}

	// Count tabs and commas
	tabCount := strings.Count(firstLine, "\t")
	commaCount := strings.Count(firstLine, ",")

	// If we have more tabs than commas, use tab as delimiter
	if tabCount > commaCount {
		return '\t', nil
	}
	// Otherwise use comma (even if counts are equal, comma is more common)
	return ',', nil
}

// ParsePRURLsFromCSV reads a CSV file and returns a slice of PR URLs
// The function detects the CSV format by analyzing headers and can handle:
// 1. A direct URL column
// 2. Separate owner, repo, and PR number columns
// The file can be either tab or comma delimited
func ParsePRURLsFromCSV(csvFile string) ([]PRURL, error) {
	file, err := os.Open(csvFile)
	if err != nil {
		return nil, fmt.Errorf("error opening CSV file: %v", err)
	}
	defer file.Close()

	// Detect the delimiter
	delimiter, err := detectDelimiter(file)
	if err != nil {
		return nil, fmt.Errorf("error detecting delimiter: %v", err)
	}

	reader := csv.NewReader(file)
	reader.Comma = delimiter

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
	for _, record := range records[1:] {
		var url string
		if format.URLColumn != -1 {
			// Use direct URL if available
			if format.URLColumn >= len(record) {
				continue
			}
			url = strings.TrimSpace(record[format.URLColumn])
		} else {
			// Build URL from components
			if format.OwnerColumn >= len(record) || format.RepoColumn >= len(record) || format.PRNumberColumn >= len(record) {
				continue
			}
			owner := strings.TrimSpace(record[format.OwnerColumn])
			repo := strings.TrimSpace(record[format.RepoColumn])
			prNumber := strings.TrimSpace(record[format.PRNumberColumn])
			url = buildGitHubURL(owner, repo, prNumber)
		}

		if url == "" {
			continue
		}

		prURLs = append(prURLs, PRURL{URL: url})
	}

	return prURLs, nil
}
