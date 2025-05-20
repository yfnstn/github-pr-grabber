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

4. (Optional) Create a `.env` file for GitHub authentication:
   ```bash
   # Create .env file
   echo "GITHUB_TOKEN=your_github_token_here" > .env
   ```
   To get a GitHub token:
   - Go to GitHub.com → Settings → Developer Settings → Personal Access Tokens
   - You can use either:
     - Classic tokens (starts with `ghp_`)
     - Fine-grained tokens (starts with `github_pat_`)
   - Generate a new token with the appropriate permissions:
     - For classic tokens: use the `repo` scope
     - For fine-grained tokens: select the specific repository and grant "Read" access
   - Copy the token and paste it in your `.env` file

## Usage

The script has two modes of operation:

### 1. List Mode
Fetches PRs from GitHub and saves them to a CSV file.

```bash
./github-pr-grabber -mode list -since YYYY-MM-DD -repo owner/repo [-search term]
```

Example:
```bash
./github-pr-grabber -mode list -since 2023-05-01 -repo yfnstn/github-pr-grabber -search "term"
```

This will create a CSV file in the `generated/csv` directory named `merged_prs_yfnstn_github-pr-grabber_20230501_term.csv` containing:
- PR Number
- Title
- Merged At
- URL (direct link to the PR on GitHub)

### 2. Open Mode
Opens PR URLs from a CSV file in your default browser.

```bash
./github-pr-grabber -mode open -urls <csv_file>
```

Example:
```bash
./github-pr-grabber -mode open -urls merged_prs_yfnstn_github-pr-grabber_20230501_term.csv
```

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
- Results are fetched in batches of 10000 to ensure complete data collection
