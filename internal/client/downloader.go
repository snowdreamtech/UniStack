// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Downloader handles HTTP downloads with retry logic and exponential backoff.
type Downloader struct {
	HTTPClient *http.Client
	MaxRetries int
}

// NewDownloader creates a new Downloader with default settings.
func NewDownloader() *Downloader {
	return &Downloader{
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		MaxRetries: 3,
	}
}

// Download stream-downloads a file from the URL. It executes the provided callback
// with the HTTP response body on success. The caller is responsible for reading and
// not closing the body (the Download method closes it).
func (d *Downloader) Download(ctx context.Context, url string, handleBody func(io.Reader) error) error {
	var lastErr error
	backoff := 1 * time.Second

	for attempt := 0; attempt <= d.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retrying, respecting context cancellation
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return ctx.Err()
			}
			backoff *= 2
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := d.HTTPClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("HTTP GET failed: %w", err)
			continue
		}

		// Handle non-2xx status codes
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			resp.Body.Close()
			if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
				lastErr = fmt.Errorf("server error or rate limited (status %d)", resp.StatusCode)
				continue
			}
			// For 4xx errors (except 429), do not retry
			return fmt.Errorf("HTTP error: %s", resp.Status)
		}

		// Success! Pass the body to the handler
		err = handleBody(resp.Body)
		resp.Body.Close()
		return err
	}

	return fmt.Errorf("download failed after %d retries, last error: %w", d.MaxRetries, lastErr)
}
