package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"time"
)

// PR represents a pull request with its key information
type PR struct {
	Number   string
	Title    string
	MergedAt string
	URL      string
}

// getMergedPRs fetches merged PRs from GitHub for the specified repository and date range
func getMergedPRs(sinceDate time.Time, repo string, searchTerm string) ([]PR, error) {
	dateStr := sinceDate.Format("2006-01-02")

	searchQuery := fmt.Sprintf("merged:>=%s", dateStr)
	if searchTerm != "" {
		searchQuery += " " + searchTerm
	}

	// Get merged PRs for the specified repository
	output, err := runGHCommand(
		"pr", "list",
		"--repo", repo,
		"--search", searchQuery,
		"--json", "number,title,mergedAt,url",
		"--jq", ".[] | [.number, .title, .mergedAt, .url] | @tsv",
		"--limit", "10000",
	)
	if err != nil {
		return nil, err
	}

	var prs []PR
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 4 {
			continue
		}
		prs = append(prs, PR{
			Number:   fields[0],
			Title:    fields[1],
			MergedAt: fields[2],
			URL:      fields[3],
		})
	}

	return prs, nil
}

// saveToCSV saves the PR list to a CSV file
func saveToCSV(prs []PR, outputFile string) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"PR Number", "Title", "Merged At", "URL"}); err != nil {
		return err
	}

	// Write PR data
	for _, pr := range prs {
		if err := writer.Write([]string{pr.Number, pr.Title, pr.MergedAt, pr.URL}); err != nil {
			return err
		}
	}

	return nil
}
