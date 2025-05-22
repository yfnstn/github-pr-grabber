package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func promptUser(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Error reading input: %v", err)
	}
	return strings.TrimSpace(input)
}

func promptDate() (time.Time, error) {
	for {
		dateStr := promptUser("Enter start date (YYYY-MM-DD): ")
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			fmt.Println("Invalid date format. Please use YYYY-MM-DD")
			continue
		}
		if date.After(time.Now()) {
			fmt.Println("Error: The date cannot be in the future")
			continue
		}
		return date, nil
	}
}

func promptRepo() string {
	for {
		repo := promptUser("Enter repository (owner/repo): ")
		if !strings.Contains(repo, "/") {
			fmt.Println("Invalid repository format. Please use owner/repo")
			continue
		}
		return repo
	}
}

func promptSearchTerm() string {
	searchTerm := promptUser("Enter search term (optional, press Enter to skip): ")
	return strings.TrimSpace(searchTerm)
}

func promptCSVFile() string {
	for {
		file := promptUser("Enter path to CSV file: ")
		if _, err := os.Stat(file); os.IsNotExist(err) {
			fmt.Println("File does not exist. Please enter a valid file path.")
			continue
		}
		return file
	}
}

func handleListMode() {
	fmt.Println("\n=== List Mode ===")
	fmt.Println("This mode will fetch PRs and save them to a CSV file.")

	sinceDate, err := promptDate()
	if err != nil {
		log.Fatalf("Error with date input: %v", err)
	}

	repo := promptRepo()
	searchTerm := promptSearchTerm()

	fmt.Printf("\nFetching PRs merged since %s for %s...\n", sinceDate.Format("2006-01-02"), repo)
	if searchTerm != "" {
		fmt.Printf("Filtering for search term: %s\n", searchTerm)
	}

	prs, err := getMergedPRs(sinceDate, repo, searchTerm)
	if err != nil {
		log.Fatalf("Error getting PRs: %v", err)
	}

	if len(prs) == 0 {
		fmt.Println("No PRs found for the specified criteria.")
		return
	}

	// Create generated/csv directory if it doesn't exist
	if err := os.MkdirAll("generated/csv", 0755); err != nil {
		log.Fatalf("Error creating output directory: %v", err)
	}

	csvFile := filepath.Join("generated/csv", fmt.Sprintf("merged_prs_%s_%s.csv",
		strings.Replace(repo, "/", "_", -1),
		sinceDate.Format("20060102")))
	if searchTerm != "" {
		csvFile = filepath.Join("generated/csv", fmt.Sprintf("%s_%s.csv",
			strings.TrimSuffix(filepath.Base(csvFile), ".csv"),
			strings.Replace(searchTerm, " ", "_", -1)))
	}

	if err := saveToCSV(prs, csvFile); err != nil {
		log.Fatalf("Error saving to CSV: %v", err)
	}
	fmt.Printf("Results saved to %s\n", csvFile)
}

func handleOpenMode() {
	fmt.Println("\n=== Open Mode ===")
	fmt.Println("This mode will open PR URLs from a CSV file in your browser.")

	csvFile := promptCSVFile()

	if err := openPRsFromCSV(csvFile); err != nil {
		log.Fatalf("Error opening PRs: %v", err)
	}
}

func main() {
	// Define flags with both long and short versions
	mode := flag.String("mode", "", "Operation mode: 'list' to get PR list, 'open' to open URLs from CSV")
	modeShort := flag.String("m", "", "Shorthand for -mode")

	sinceDateStr := flag.String("since", "", "Start date in YYYY-MM-DD format (for list mode)")
	sinceDateStrShort := flag.String("s", "", "Shorthand for -since")

	repo := flag.String("repo", "", "GitHub repository in owner/repo format (for list mode)")
	repoShort := flag.String("r", "", "Shorthand for -repo")

	searchTerm := flag.String("search", "", "Optional search term (for list mode)")
	searchTermShort := flag.String("q", "", "Shorthand for -search (query)")

	urlsFile := flag.String("urls", "", "CSV file containing PR URLs (for open mode)")
	urlsFileShort := flag.String("u", "", "Shorthand for -urls")

	interactive := flag.Bool("i", false, "Run in interactive mode")

	flag.Parse()

	// Use shorthand values if provided
	if *modeShort != "" {
		*mode = *modeShort
	}
	if *sinceDateStrShort != "" {
		*sinceDateStr = *sinceDateStrShort
	}
	if *repoShort != "" {
		*repo = *repoShort
	}
	if *searchTermShort != "" {
		*searchTerm = *searchTermShort
	}
	if *urlsFileShort != "" {
		*urlsFile = *urlsFileShort
	}

	// If no flags are provided or interactive mode is requested, run interactively
	if *interactive || (flag.NFlag() == 0 && !flag.Parsed()) {
		runInteractiveMode()
		return
	}

	// Handle command-line mode
	switch *mode {
	case "list":
		if *sinceDateStr == "" || *repo == "" {
			fmt.Println("Usage for list mode:")
			fmt.Println("  ./github-pr-grabber -mode list -since YYYY-MM-DD -repo owner/repo [-search term]")
			fmt.Println("  or using shorthand flags:")
			fmt.Println("  ./github-pr-grabber -m list -s YYYY-MM-DD -r owner/repo [-q term]")
			fmt.Println("  or")
			fmt.Println("  ./github-pr-grabber -i")
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
			fmt.Println("  or using shorthand flags:")
			fmt.Println("  ./github-pr-grabber -m open -u <csv_file>")
			fmt.Println("  or")
			fmt.Println("  ./github-pr-grabber -i")
			flag.PrintDefaults()
			os.Exit(1)
		}

		if err := openPRsFromCSV(*urlsFile); err != nil {
			log.Fatalf("Error opening PRs: %v", err)
		}

	default:
		fmt.Println("Please specify a mode: 'list' or 'open'")
		fmt.Println("\nList mode usage:")
		fmt.Println("  ./github-pr-grabber -mode list -since YYYY-MM-DD -repo owner/repo [-search term]")
		fmt.Println("  or using shorthand flags:")
		fmt.Println("  ./github-pr-grabber -m list -s YYYY-MM-DD -r owner/repo [-q term]")
		fmt.Println("\nOpen mode usage:")
		fmt.Println("  ./github-pr-grabber -mode open -urls <csv_file>")
		fmt.Println("  or using shorthand flags:")
		fmt.Println("  ./github-pr-grabber -m open -u <csv_file>")
		fmt.Println("\nOr run in interactive mode:")
		fmt.Println("  ./github-pr-grabber -i")
		os.Exit(1)
	}
}

func runInteractiveMode() {
	fmt.Println("GitHub PR Grabber")
	fmt.Println("=================")

	for {
		fmt.Println("\nSelect a mode:")
		fmt.Println("1. List Mode - Fetch PRs and save to CSV")
		fmt.Println("2. Open Mode - Open PRs from CSV in browser")
		fmt.Println("3. Exit")

		choice := promptUser("Enter your choice (1-3): ")

		switch choice {
		case "1":
			handleListMode()
		case "2":
			handleOpenMode()
		case "3":
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Println("Invalid choice. Please enter 1, 2, or 3.")
		}
	}
}
