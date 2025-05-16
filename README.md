# GitHub PR Grabber

A Go script to support audit evidence collection, the script can: 
- Track merged pull requests in a GitHub repository 
- Open a provided list of PRs in your browser 
- Generate PDFs/screenshots of PR pages from a provided list of urls

## Prerequisites

- Go 1.x
- GitHub CLI (`gh`) installed and authenticated (for working with private repos)
- Playwright (for PDF/screenshot generation)
- Node.js (for installing Playwright)

## Installation

1. Make sure you have the GitHub CLI installed:

For installation instructiosn see https://github.com/cli/cli#installation 

2. Authenticate with GitHub:
   ```bash
   gh auth login
   ```

3. Install Playwright and its dependencies:
   ```bash
   # Install Node.js if you haven't already
   brew install node

   # Install Playwright
   npm install -g playwright
   playwright install chromium
   ```

4. Build the Go script:
   ```bash
   go build
   ```

5. (Optional) Create a `.env` file for GitHub authentication:
   ```bash
   # Create .env file
   echo "GITHUB_TOKEN=your_github_token_here" > .env
   ```
   To get a GitHub token:
   - Go to GitHub.com → Settings → Developer Settings → Personal Access Tokens → Fine-grained-tokens
   - Generate a new token with the `repo` scope (for private repos)
   - Copy the token and paste it in your `.env` file

## Usage

The script has three modes of operation:

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

### 3. Capture Mode
Generates PDFs or screenshots of PR pages from URLs in a CSV file.

```bash
./github-pr-grabber -mode capture -urls <csv_file> [-format pdf|png] [-output dir] [-wait seconds] [-fullpage]
```

Options:
- `-format`: Output format, either "pdf" (default) or "png"
- `-output`: Output directory (default: "generated/captures")
- `-wait`: Seconds to wait for page load (default: 5)
- `-fullpage`: Whether to capture the full page (default: true)

Example:
```bash
# Generate PDFs
./github-pr-grabber -mode capture -urls merged_prs.csv -output ./generated/captures -wait 10

# Generate screenshots
./github-pr-grabber -mode capture -urls merged_prs.csv -format png -output ./generated/captures
```

For private repositories, make sure you have set up your GitHub token in the `.env` file as described in the installation section.

## Features

- Fetch up to 10,000 PRs in a single query
- Filter PRs by date and search term
- Save PR details to CSV
- Open PRs in your default browser
- Generate PDFs or screenshots of PR pages
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
  - Captures (PDFs/PNGs) are stored in `generated/captures/` by default
- The script will create the output directories if they don't exist
- For best results with PDF generation, use a wait time of at least 5-10 seconds 
- The script will fetch all matching PRs, not just the first 30 results
- Results are fetched in batches of 10000 to ensure complete data collection 
