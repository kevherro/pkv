package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Remote struct {
	client  *http.Client
	baseURL string
	timeout time.Duration
}

func (r *Remote) Get(endpoint string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	resp, err := r.request(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("remote get: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("remote get: received non-OK status: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("remote get: %v", err)
	}

	return string(body), nil
}

func (r *Remote) Head(endpoint string) (http.Header, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	resp, err := r.request(ctx, http.MethodHead, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("remote head: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("remote head: resource not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("remote head: received non-OK status: %d %s", resp.StatusCode, resp.Status)
	}

	return resp.Header, nil
}

func (r *Remote) request(ctx context.Context, method string, endpoint string, body io.Reader) (*http.Response, error) {
	url := r.baseURL + endpoint

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("request: %v", err)
	}

	return r.client.Do(req)
}
