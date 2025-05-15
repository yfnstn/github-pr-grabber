package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

// CaptureOptions defines the options for capturing PR pages
type CaptureOptions struct {
	Format    string // "pdf" or "png"
	OutputDir string
	WaitTime  int    // seconds to wait for page load
	FullPage  bool   // whether to capture full page
	AuthToken string // GitHub Personal Access Token
}

// capturePRPage captures a PR page as PDF or PNG
func capturePRPage(url string, options CaptureOptions) error {
	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("could not start playwright: %v", err)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch()
	if err != nil {
		return fmt.Errorf("could not launch browser: %v", err)
	}
	defer browser.Close()

	// Create a new context
	context, err := browser.NewContext()
	if err != nil {
		return fmt.Errorf("could not create context: %v", err)
	}
	defer context.Close()

	// If auth token is provided, set it in the Authorization header
	if options.AuthToken != "" {
		// Set up request interception to add the auth header
		if err := context.Route("**/*", func(route playwright.Route) {
			headers := route.Request().Headers()
			headers["Authorization"] = fmt.Sprintf("Bearer %s", options.AuthToken)
			route.Continue(playwright.RouteContinueOptions{
				Headers: headers,
			})
		}); err != nil {
			return fmt.Errorf("could not set up request interception: %v", err)
		}
	}

	page, err := context.NewPage()
	if err != nil {
		return fmt.Errorf("could not create page: %v", err)
	}

	// Navigate to the PR page
	if _, err := page.Goto(url); err != nil {
		return fmt.Errorf("could not goto: %v", err)
	}

	// Wait for the page to be fully loaded
	time.Sleep(time.Duration(options.WaitTime) * time.Second)

	// Extract PR number and repo name from URL for filename
	// URL format: https://github.com/owner/repo/pull/123
	parts := strings.Split(url, "/")
	if len(parts) < 7 {
		return fmt.Errorf("invalid PR URL format: %s", url)
	}
	repo := parts[4]
	prNumber := parts[6]
	filename := fmt.Sprintf("%s_pr_%s", repo, prNumber)

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(options.OutputDir, 0755); err != nil {
		return fmt.Errorf("could not create output directory: %v", err)
	}

	// Generate the output file path
	outputPath := filepath.Join(options.OutputDir, filename)
	if options.Format == "pdf" {
		outputPath += ".pdf"
		_, err := page.PDF(playwright.PagePdfOptions{
			Path:            playwright.String(outputPath),
			Format:          playwright.String("Letter"),
			PrintBackground: playwright.Bool(true),
		})
		if err != nil {
			return fmt.Errorf("could not save PDF: %v", err)
		}
	} else {
		outputPath += ".png"
		_, err := page.Screenshot(playwright.PageScreenshotOptions{
			Path:     playwright.String(outputPath),
			FullPage: playwright.Bool(options.FullPage),
		})
		if err != nil {
			return fmt.Errorf("could not save screenshot: %v", err)
		}
	}

	return nil
}
