package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

type PR struct {
	Number   string
	Title    string
	MergedAt string
	URL      string
}

func runGHCommand(args ...string) (string, error) {
	cmd := exec.Command("gh", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error running GitHub CLI command: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

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

func main() {
	// Define mode flag
	mode := flag.String("mode", "", "Operation mode: 'list' to get PR list, 'open' to open URLs from CSV")

	// PR list mode flags
	sinceDateStr := flag.String("since", "", "Start date in YYYY-MM-DD format (for list mode)")
	repo := flag.String("repo", "", "GitHub repository in owner/repo format (for list mode)")
	searchTerm := flag.String("search", "", "Optional search term (for list mode)")

	// Open mode flags
	urlsFile := flag.String("urls", "", "CSV file containing PR URLs (for open mode)")

	flag.Parse()

	switch *mode {
	case "list":
		if *sinceDateStr == "" || *repo == "" {
			fmt.Println("Usage for list mode:")
			fmt.Println("  ./github-pr-tracker -mode list -since YYYY-MM-DD -repo owner/repo [-search term]")
			flag.PrintDefaults()
			os.Exit(1)
		}

		sinceDate, err := time.Parse("2006-01-02", *sinceDateStr)
		if err != nil {
			log.Fatalf("Invalid date format: %v", err)
		}

		if sinceDate.After(time.Now()) {
			log.Fatalf("Error: The date %s is in the future", sinceDate.Format("2006-01-02"))
		}

		fmt.Printf("Fetching PRs merged since %s for %s...\n", sinceDate.Format("2006-01-02"), *repo)
		if *searchTerm != "" {
			fmt.Printf("Filtering for search term: %s\n", *searchTerm)
		}

		prs, err := getMergedPRs(sinceDate, *repo, *searchTerm)
		if err != nil {
			log.Fatalf("Error getting PRs: %v", err)
		}

		if len(prs) == 0 {
			fmt.Println("No PRs found for the specified criteria.")
			os.Exit(0)
		}

		csvFile := fmt.Sprintf("merged_prs_%s_%s.csv",
			strings.Replace(*repo, "/", "_", -1),
			sinceDate.Format("20060102"))
		if *searchTerm != "" {
			csvFile = fmt.Sprintf("%s_%s.csv",
				strings.TrimSuffix(csvFile, ".csv"),
				strings.Replace(*searchTerm, " ", "_", -1))
		}

		if err := saveToCSV(prs, csvFile); err != nil {
			log.Fatalf("Error saving to CSV: %v", err)
		}
		fmt.Printf("Results saved to %s\n", csvFile)

	case "open":
		if *urlsFile == "" {
			fmt.Println("Usage for open mode:")
			fmt.Println("  ./github-pr-tracker -mode open -urls <csv_file>")
			flag.PrintDefaults()
			os.Exit(1)
		}

		// Read URLs from CSV
		file, err := os.Open(*urlsFile)
		if err != nil {
			log.Fatalf("Error opening CSV file: %v", err)
		}
		defer file.Close()

		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		if err != nil {
			log.Fatalf("Error reading CSV file: %v", err)
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

	default:
		fmt.Println("Please specify a mode: 'list' or 'open'")
		fmt.Println("\nList mode usage:")
		fmt.Println("  ./github-pr-tracker -mode list -since YYYY-MM-DD -repo owner/repo [-search term]")
		fmt.Println("\nOpen mode usage:")
		fmt.Println("  ./github-pr-tracker -mode open -urls <csv_file>")
		os.Exit(1)
	}
}
