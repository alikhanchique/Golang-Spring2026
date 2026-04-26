package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"time"
)

type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

type PaymentClient struct {
	cfg        RetryConfig
	httpClient *http.Client
	serverURL  string
}

func NewPaymentClient(serverURL string, cfg RetryConfig) *PaymentClient {
	return &PaymentClient{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 5 * time.Second},
		serverURL:  serverURL,
	}
}

// IsRetryable returns true for transient errors: network timeouts and HTTP 429/500/502/503/504.
// Returns false for permanent errors like 401 or 404.
func IsRetryable(resp *http.Response, err error) bool {
	if err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) && urlErr.Timeout() {
			return true
		}
		return true
	}
	if resp != nil {
		switch resp.StatusCode {
		case 429, 500, 502, 503, 504:
			return true
		}
	}
	return false
}

// CalculateBackoff computes exponential backoff with Full Jitter for the given attempt (0-indexed).
func CalculateBackoff(attempt int, base, max time.Duration) time.Duration {
	backoff := base * time.Duration(math.Pow(2, float64(attempt)))
	if backoff > max {
		backoff = max
	}
	return time.Duration(rand.Int63n(int64(backoff) + 1))
}

// ExecutePayment sends a POST to the payment endpoint, retrying on transient failures.
// Respects context cancellation immediately — does not sleep after context is done.
func (c *PaymentClient) ExecutePayment(ctx context.Context) (*http.Response, error) {
	var (
		lastResp *http.Response
		lastErr  error
	)

	for attempt := 0; attempt < c.cfg.MaxRetries; attempt++ {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.serverURL+"/pay", nil)
		if err != nil {
			return nil, fmt.Errorf("building request: %w", err)
		}

		resp, err := c.httpClient.Do(req)
		lastErr = err
		lastResp = resp

		if err == nil && !IsRetryable(resp, nil) {
			fmt.Printf("Attempt %d: Success! Status %d\n", attempt+1, resp.StatusCode)
			return resp, nil
		}

		retryable := IsRetryable(resp, err)
		if !retryable {
			if err != nil {
				return nil, fmt.Errorf("permanent error: %w", err)
			}
			return resp, nil
		}

		if attempt == c.cfg.MaxRetries-1 {
			break
		}

		wait := CalculateBackoff(attempt, c.cfg.BaseDelay, c.cfg.MaxDelay)
		statusCode := 0
		if resp != nil {
			statusCode = resp.StatusCode
		}
		if err != nil {
			fmt.Printf("Attempt %d failed (error: %v): waiting %v...\n", attempt+1, err, wait.Round(time.Millisecond))
		} else {
			fmt.Printf("Attempt %d failed (status %d): waiting %v...\n", attempt+1, statusCode, wait.Round(time.Millisecond))
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(wait):
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("all %d attempts failed, last error: %w", c.cfg.MaxRetries, lastErr)
	}
	return lastResp, fmt.Errorf("all %d attempts failed, last status: %d", c.cfg.MaxRetries, lastResp.StatusCode)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	var requestCount int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt64(&requestCount, 1)
		if n <= 3 {
			fmt.Printf("[Server] Request #%d -> 503 Service Unavailable\n", n)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		fmt.Printf("[Server] Request #%d -> 200 OK\n", n)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "success"}`))
	}))
	defer server.Close()

	client := NewPaymentClient(server.URL, RetryConfig{
		MaxRetries: 5,
		BaseDelay:  500 * time.Millisecond,
		MaxDelay:   10 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("=== Starting payment with retry ===")
	resp, err := client.ExecutePayment(ctx)
	if err != nil {
		fmt.Printf("Payment failed: %v\n", err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("Payment succeeded with status %d\n", resp.StatusCode)
}
