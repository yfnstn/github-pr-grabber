package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// Only log if the error is not "file not found"
		if !strings.Contains(err.Error(), "no such file") {
			log.Printf("Warning: Error loading .env file: %v", err)
		}
	}

	// Define mode flag
	mode := flag.String("mode", "", "Operation mode: 'list' to get PR list, 'open' to open URLs from CSV, 'capture' to generate PDFs/screenshots")

	// PR list mode flags
	sinceDateStr := flag.String("since", "", "Start date in YYYY-MM-DD format (for list mode)")
	repo := flag.String("repo", "", "GitHub repository in owner/repo format (for list mode)")
	searchTerm := flag.String("search", "", "Optional search term (for list mode)")

	// Open mode flags
	urlsFile := flag.String("urls", "", "CSV file containing PR URLs (for open mode)")

	// Capture mode flags
	captureFormat := flag.String("format", "pdf", "Capture format: 'pdf' or 'png' (for capture mode)")
	captureOutputDir := flag.String("output", "generated/captures", "Output directory for captures (for capture mode)")
	captureWaitTime := flag.Int("wait", 5, "Seconds to wait for page load (for capture mode)")
	captureFullPage := flag.Bool("fullpage", true, "Capture full page (for capture mode)")

	flag.Parse()

	switch *mode {
	case "list":
		if *sinceDateStr == "" || *repo == "" {
			fmt.Println("Usage for list mode:")
			fmt.Println("  ./github-pr-grabber -mode list -since YYYY-MM-DD -repo owner/repo [-search term]")
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

		// Create generated/csv directory if it doesn't exist
		if err := os.MkdirAll("generated/csv", 0755); err != nil {
			log.Fatalf("Error creating output directory: %v", err)
		}

		csvFile := filepath.Join("generated/csv", fmt.Sprintf("merged_prs_%s_%s.csv",
			strings.Replace(*repo, "/", "_", -1),
			sinceDate.Format("20060102")))
		if *searchTerm != "" {
			csvFile = filepath.Join("generated/csv", fmt.Sprintf("%s_%s.csv",
				strings.TrimSuffix(filepath.Base(csvFile), ".csv"),
				strings.Replace(*searchTerm, " ", "_", -1)))
		}

		if err := saveToCSV(prs, csvFile); err != nil {
			log.Fatalf("Error saving to CSV: %v", err)
		}
		fmt.Printf("Results saved to %s\n", csvFile)

	case "open":
		if *urlsFile == "" {
			fmt.Println("Usage for open mode:")
			fmt.Println("  ./github-pr-grabber -mode open -urls <csv_file>")
			flag.PrintDefaults()
			os.Exit(1)
		}

		if err := openPRsFromCSV(*urlsFile); err != nil {
			log.Fatalf("Error opening PRs: %v", err)
		}

	case "capture":
		if *urlsFile == "" {
			fmt.Println("Usage for capture mode:")
			fmt.Println("  ./github-pr-grabber -mode capture -urls <csv_file> [-format pdf|png] [-output dir] [-wait seconds] [-fullpage]")
			fmt.Println("\nNote: For private repos, set the GITHUB_TOKEN environment variable with your GitHub Personal Access Token")
			flag.PrintDefaults()
			os.Exit(1)
		}

		if *captureFormat != "pdf" && *captureFormat != "png" {
			log.Fatalf("Invalid format: %s. Must be 'pdf' or 'png'", *captureFormat)
		}

		// Get token from environment variable
		authToken := os.Getenv("GITHUB_TOKEN")
		if authToken == "" {
			fmt.Println("Warning: GITHUB_TOKEN environment variable not set. Private repos may not be accessible.")
		}

		// Read URLs from CSV
		prURLs, err := ParsePRURLsFromCSV(*urlsFile)
		if err != nil {
			log.Fatalf("Error reading CSV file: %v", err)
		}

		options := CaptureOptions{
			Format:    *captureFormat,
			OutputDir: *captureOutputDir,
			WaitTime:  *captureWaitTime,
			FullPage:  *captureFullPage,
			AuthToken: authToken,
		}

		for i, pr := range prURLs {
			fmt.Printf("\nCapturing PR %d/%d: %s\n", i+1, len(prURLs), pr.URL)
			if err := capturePRPage(pr.URL, options); err != nil {
				fmt.Printf("Error capturing PR: %v\n", err)
				continue
			}
		}

	default:
		fmt.Println("Please specify a mode: 'list', 'open', or 'capture'")
		fmt.Println("\nList mode usage:")
		fmt.Println("  ./github-pr-grabber -mode list -since YYYY-MM-DD -repo owner/repo [-search term]")
		fmt.Println("\nOpen mode usage:")
		fmt.Println("  ./github-pr-grabber -mode open -urls <csv_file>")
		fmt.Println("\nCapture mode usage:")
		fmt.Println("  ./github-pr-grabber -mode capture -urls <csv_file> [-format pdf|png] [-output dir] [-wait seconds] [-fullpage]")
		os.Exit(1)
	}
}
