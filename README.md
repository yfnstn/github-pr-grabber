# GitHub PR Tracker

A Go script to track merged pull requests in a GitHub repository and open PRs in your browser.

## Prerequisites

- Go 1.x
- GitHub CLI (`gh`) installed and authenticated

## Installation

1. Make sure you have the GitHub CLI installed:

For installation instructiosn see https://github.com/cli/cli#installation 

2. Authenticate with GitHub:
   ```bash
   gh auth login
   ```

3. Build the Go script:
   ```bash
   go build
   ```

## Usage

The script has two modes of operation:

### 1. List Mode
Fetches PRs from GitHub and saves them to a CSV file.

```bash
./github-pr-tracker -mode list -since YYYY-MM-DD -repo owner/repo [-search term]
```

Example:
```bash
./github-pr-tracker -mode list -since 2023-05-01 -repo yfnstn/github-pr-tracker -search "initial"
```

This will create a CSV file named `merged_prs_yfnstn_github-pr-tracker_20230501_initial.csv` containing:
- PR Number
- Title
- Merged At
- URL (direct link to the PR on GitHub)

### 2. Open Mode
Opens PR URLs from a CSV file in your default browser.

```bash
./github-pr-tracker -mode open -urls <csv_file>
```

Example:
```bash
./github-pr-tracker -mode open -urls merged_prs_yfnstn_github-pr-tracker_20230501_initial.csv
```

## Features

- Fetch up to 1000 PRs in a single query
- Filter PRs by date and search term
- Save PR details to CSV
- Open PRs in your default browser
- Support for GitHub CLI authentication
- Can be run from any directory

## Notes

- The script requires GitHub CLI authentication to access the repository
- The search term is optional and will filter PRs by matching the term in their titles or descriptions
- You can run the script from any directory - it no longer needs to be run from within the target repository
- The script will fetch all matching PRs, not just the first 30 results
- Results are fetched in batches of 10000 to ensure complete data collection 
