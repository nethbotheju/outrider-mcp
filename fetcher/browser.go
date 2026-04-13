package fetcher

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"

	"github.com/chromedp/chromedp"
)

// chromeAvailable checks if a Chrome/Chromium binary exists on the system.
func chromeAvailable() bool {
	var name string
	switch runtime.GOOS {
	case "darwin":
		name = "Google Chrome"
	case "windows":
		name = "chrome"
	default:
		name = "chromium-browser"
	}
	_, err := exec.LookPath(name)
	return err == nil
}

func (f *Fetcher) fetchViaBrowser(ctx context.Context, rawURL string) (*FetchResult, error) {
	allocCtx, cancel := chromedp.NewContext(ctx, chromedp.WithLogf(log.Printf))
	defer cancel()

	var htmlContent string
	err := chromedp.Run(allocCtx,
		chromedp.Navigate(rawURL),
		chromedp.WaitReady("body"),
		chromedp.OuterHTML("html", &htmlContent),
	)
	if err != nil {
		return nil, fmt.Errorf("browser: %w", err)
	}

	if strings.TrimSpace(htmlContent) == "" {
		return nil, fmt.Errorf("browser: empty response from %s", rawURL)
	}

	return extractContent([]byte(htmlContent), rawURL)
}
