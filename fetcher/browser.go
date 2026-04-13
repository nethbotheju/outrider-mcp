package fetcher

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/chromedp/chromedp"
)

// chromeAvailable checks if a Chrome/Chromium binary exists on the system.
func chromeAvailable() bool {
	switch runtime.GOOS {
	case "darwin":
		if _, err := os.Stat("/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"); err == nil {
			return true
		}
		_, err := exec.LookPath("google-chrome")
		return err == nil
	case "windows":
		_, err := exec.LookPath("chrome")
		return err == nil
	default:
		_, err := exec.LookPath("chromium-browser")
		if err != nil {
			_, err = exec.LookPath("google-chrome")
		}
		return err == nil
	}
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
