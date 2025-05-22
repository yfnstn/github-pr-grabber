# GitHub PR Grabber

A Go script to support audit evidence collection, the script can: 
- Track merged pull requests in a GitHub repository 
- Open a provided list of PRs in your browser

## Prerequisites

- Go 1.x
- GitHub CLI (`gh`) installed and authenticated (for working with private repos)

## Installation

1. Make sure you have the GitHub CLI installed:

   For installation instructions see https://github.com/cli/cli#installation 

2. Authenticate with GitHub:
   ```bash
   gh auth login
   ```

3. Build the Go script:
   ```bash
   go build
   ```

## Usage

The script can be used in two ways: interactive mode or command-line mode.

### Interactive Mode

Run the program without any arguments or with the `-i` flag:
```bash
./github-pr-grabber
# or
./github-pr-grabber -i
```

You'll be presented with an interactive menu:
```
GitHub PR Grabber
=================

Select a mode:
1. List Mode - Fetch PRs and save to CSV
2. Open Mode - Open PRs from CSV in browser
3. Exit
```

### Command-Line Mode

Alternatively, you can use command-line flags for automation or scripting:

#### List Mode
```bash
./github-pr-grabber -mode list -since YYYY-MM-DD -repo owner/repo [-search term]
```

Example:
```bash
./github-pr-grabber -mode list -since 2023-05-01 -repo yfnstn/github-pr-grabber -search "security"
```

#### Open Mode
```bash
./github-pr-grabber -mode open -urls <csv_file>
```

Example:
```bash
./github-pr-grabber -mode open -urls generated/csv/merged_prs_yfnstn_github-pr-grabber_20230501_security.csv
```

### Available Flags

Long form flags:
- `-mode`: Operation mode ('list' or 'open')
- `-since`: Start date in YYYY-MM-DD format (for list mode)
- `-repo`: GitHub repository in owner/repo format (for list mode)
- `-search`: Optional search term (for list mode)
- `-urls`: CSV file containing PR URLs (for open mode)
- `-i`: Run in interactive mode

Shorthand flags:
- `-m`: Shorthand for -mode
- `-s`: Shorthand for -since
- `-r`: Shorthand for -repo
- `-q`: Shorthand for -search (query)
- `-u`: Shorthand for -urls

Examples using shorthand flags:
```bash
# List mode with shorthand flags
./github-pr-grabber -m list -s 2023-05-01 -r yfnstn/github-pr-grabber -q security

# Open mode with shorthand flags
./github-pr-grabber -m open -u generated/csv/merged_prs.csv
```

### Mode Details

#### 1. List Mode
Fetches PRs from GitHub and saves them to a CSV file.

In interactive mode, you'll be prompted for:
- Start date (in YYYY-MM-DD format)
- Repository (in owner/repo format)
- Optional search term

The script will create a CSV file in the `generated/csv` directory containing:
- PR Number
- Title
- Merged At
- URL (direct link to the PR on GitHub)

#### 2. Open Mode
Opens PR URLs from a CSV file in your default browser.

In interactive mode, you'll be prompted for:
- Path to the CSV file containing PR URLs

## Features

- Fetch up to 10,000 PRs in a single query
- Filter PRs by date and search term
- Save PR details to CSV
- Open PRs in your default browser
- Support for private repositories via GitHub token
- Support for GitHub CLI authentication
- Can be run from any directory

## Notes

- The script requires GitHub CLI authentication to access the repository
- For private repositories, a GitHub Personal Access Token is required (set in `.env`)
- The search term is optional and will filter PRs by matching the term in their titles or descriptions
- You can run the script from any directory - it no longer needs to be run from within the target repository
- All generated files are stored in the `generated` directory:
  - CSV files are stored in `generated/csv/`
- The script will create the output directories if they don't exist
- The script will fetch all matching PRs, not just the first 30 results
- Results are fetched in batches of 10,000 to ensure complete data collection
