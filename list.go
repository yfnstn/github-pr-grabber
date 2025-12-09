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

// fetchPRsForDateRange fetches PRs for a specific date range and returns them along with the count
func fetchPRsForDateRange(startDate, endDate time.Time, repo, searchTerm string) ([]PR, int, error) {
	startStr := startDate.Format("2006-01-02")
	endStr := endDate.Format("2006-01-02")

	// Build search query for this date range
	searchQuery := fmt.Sprintf("merged:%s..%s", startStr, endStr)
	if searchTerm != "" {
		searchQuery += " " + searchTerm
	}

	// Get merged PRs for this date range
	output, err := runGHCommand(
		"pr", "list",
		"--repo", repo,
		"--search", searchQuery,
		"--json", "number,title,mergedAt,url",
		"--jq", ".[] | [.number, .title, .mergedAt, .url] | @tsv",
		"--limit", "1000",
	)
	if err != nil {
		return nil, 0, err
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

	return prs, len(prs), nil
}

// fetchPRsRecursive fetches PRs for a date range, recursively splitting if we hit the 1000 limit
func fetchPRsRecursive(startDate, endDate time.Time, repo, searchTerm string, seenPRs map[string]bool, allPRs *[]PR, depth int) error {
	// Prevent infinite recursion
	if depth > 10 {
		return fmt.Errorf("maximum recursion depth reached for date range %s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	}

	startStr := startDate.Format("2006-01-02")
	endStr := endDate.Format("2006-01-02")

	prs, count, err := fetchPRsForDateRange(startDate, endDate, repo, searchTerm)
	if err != nil {
		return fmt.Errorf("error fetching PRs for %s to %s: %v", startStr, endStr, err)
	}

	// If we got exactly 1000 results, we might have hit the limit - split the range
	if count >= 1000 {
		// Calculate midpoint
		duration := endDate.Sub(startDate)
		if duration < 24*time.Hour {
			// Can't split further (less than a day), warn and continue
			fmt.Printf("  Warning: Hit 1000 PR limit for %s to %s (less than 1 day, cannot split further)\n", startStr, endStr)
		} else {
			// Split in half and fetch both halves
			midpoint := startDate.Add(duration / 2)
			fmt.Printf("  Hit 1000 PR limit for %s to %s, splitting into smaller chunks...\n", startStr, endStr)

			// Fetch first half
			if err := fetchPRsRecursive(startDate, midpoint, repo, searchTerm, seenPRs, allPRs, depth+1); err != nil {
				return err
			}

			// Fetch second half (add 1 second to avoid overlap)
			if err := fetchPRsRecursive(midpoint.Add(time.Second), endDate, repo, searchTerm, seenPRs, allPRs, depth+1); err != nil {
				return err
			}

			return nil
		}
	}

	// Add PRs that we haven't seen before
	newCount := 0
	for _, pr := range prs {
		if !seenPRs[pr.URL] {
			*allPRs = append(*allPRs, pr)
			seenPRs[pr.URL] = true
			newCount++
		}
	}

	if depth == 0 {
		fmt.Printf("  Found %d PRs in this chunk (total so far: %d)\n", newCount, len(*allPRs))
	}

	return nil
}

// getMergedPRs fetches merged PRs from GitHub for the specified repository and date range
// To work around GitHub's 1000 result limit, this function splits the date range into
// monthly chunks and fetches PRs for each chunk separately. If a chunk hits the limit,
// it recursively splits that chunk into smaller pieces.
func getMergedPRs(sinceDate time.Time, repo string, searchTerm string) ([]PR, error) {
	now := time.Now()
	var allPRs []PR

	// Use a map to track seen PRs by URL to avoid duplicates
	seenPRs := make(map[string]bool)

	// Split the date range into monthly chunks to avoid hitting the 1000 result limit
	currentStart := sinceDate
	chunkCount := 0

	for currentStart.Before(now) {
		chunkCount++
		// Calculate end date for this chunk (one month later, or now if that's earlier)
		currentEnd := currentStart.AddDate(0, 1, 0)
		if currentEnd.After(now) {
			currentEnd = now
		}

		startStr := currentStart.Format("2006-01-02")
		endStr := currentEnd.Format("2006-01-02")

		fmt.Printf("Fetching PRs for chunk %d: %s to %s...\n", chunkCount, startStr, endStr)

		// Fetch PRs for this chunk (with recursive splitting if needed)
		if err := fetchPRsRecursive(currentStart, currentEnd, repo, searchTerm, seenPRs, &allPRs, 0); err != nil {
			fmt.Printf("Warning: Error fetching PRs for %s to %s: %v\n", startStr, endStr, err)
		}

		// Move to next chunk
		currentStart = currentEnd
	}

	fmt.Printf("\nTotal PRs fetched: %d\n", len(allPRs))
	return allPRs, nil
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
