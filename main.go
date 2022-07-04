package main

import (
	"cmpro/must"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/chromedp"
)

const (
	executionTimeout = 60 * time.Second

	// TODO: Remove it
	// file:///home/alex/Downloads/Production.html
	defaultServiceURL = ""

	downloadButtonSelector = "#downloadButton"
)

func main() {
	var serviceURL, downloadsDir string

	flag.StringVar(&serviceURL, "service-url", defaultServiceURL, "CM Pro Service URL")
	flag.StringVar(&downloadsDir, "downloads-dir", filepath.Join(must.String(os.Getwd()), "Downloads"), "Downloads dir")
	flag.Parse()

	if serviceURL == "" {
		log.Fatal("service URL is not set")
	}

	// Ensure downloads dir exists
	if err := os.MkdirAll(downloadsDir, os.ModePerm); err != nil {
		log.Fatal(fmt.Errorf("can't create a downloads dir: %w", err))
	}
	ctx, cancel := chromedp.NewContext(
		context.Background(),
		// chromedp.WithLogf(log.Printf),
		// chromedp.WithDebugf(log.Printf),
	)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, executionTimeout)
	defer cancel()

	done := make(chan struct{})

	var downloadMu sync.Mutex
	var downloadGUID string

	chromedp.ListenTarget(ctx, func(v interface{}) {
		downloadMu.Lock()
		defer downloadMu.Unlock()

		switch ev := v.(type) {
		case *browser.EventDownloadWillBegin:
			downloadGUID = ev.GUID
		case *browser.EventDownloadProgress:
			if ev.GUID == downloadGUID && ev.State == "completed" {
				done <- struct{}{}
			}
		}
	})

	if err := chromedp.Run(ctx,
		chromedp.Navigate(serviceURL),
		chromedp.WaitVisible(downloadButtonSelector),
		browser.SetDownloadBehavior(browser.SetDownloadBehaviorBehaviorAllow).
			WithDownloadPath(downloadsDir).WithEventsEnabled(true),
	); err != nil {
		log.Fatal(fmt.Errorf("can't load service page: %w", err))
	}

	if err := chromedp.Run(ctx, chromedp.Click(downloadButtonSelector)); err != nil {
		log.Fatal(fmt.Errorf("can't click download button: %w", err))
	}

	select {
	case <-ctx.Done():
		log.Fatal(fmt.Errorf("failure: %w", ctx.Err()))
	case <-done:
		fmt.Println("Finished")
	}
}
