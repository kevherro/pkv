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

func NewRemote(baseURL string, timeout time.Duration, client *http.Client) *Remote {
	if client == nil {
		client = &http.Client{}
	}

	return &Remote{
		client:  client,
		baseURL: baseURL,
		timeout: timeout,
	}
}

func (r *Remote) Delete(endpoint string) error {
	_, err := r.requestWithExpectedStatus(http.MethodDelete, endpoint, nil, []int{http.StatusNoContent, http.StatusNotFound})
	if err != nil {
		return err
	}

	return nil
}

func (r *Remote) Put(endpoint string, contentLength int64, body io.Reader) (http.Header, error) {
	req, err := r.newRequest(http.MethodPut, endpoint, body)
	if err != nil {
		return nil, err
	}
	req.ContentLength = contentLength

	resp, err := r.doRequestWithExpectedStatus(req, []int{http.StatusCreated, http.StatusNotFound})
	if err != nil {
		return nil, err
	}

	return resp.Header, nil
}

func (r *Remote) Get(endpoint string) (string, error) {
	resp, err := r.requestWithExpectedStatus(http.MethodGet, endpoint, nil, []int{http.StatusOK})
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("remote get: %v", err)
	}

	return string(body), nil
}

func (r *Remote) Head(endpoint string) (http.Header, error) {
	resp, err := r.requestWithExpectedStatus(http.MethodHead, endpoint, nil, []int{http.StatusOK, http.StatusNotFound})
	if err != nil {
		return nil, err
	}

	return resp.Header, nil
}

func (r *Remote) requestWithExpectedStatus(method string, endpoint string, body io.Reader, expectedStatusCodes []int) (*http.Response, error) {
	req, err := r.newRequest(method, endpoint, body)
	if err != nil {
		return nil, err
	}

	return r.doRequestWithExpectedStatus(req, expectedStatusCodes)
}

func (r *Remote) newRequest(method string, endpoint string, body io.Reader) (*http.Request, error) {
	url := r.baseURL + endpoint

	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	return http.NewRequestWithContext(ctx, method, url, body)
}

func (r *Remote) doRequestWithExpectedStatus(req *http.Request, expectedStatusCodes []int) (*http.Response, error) {
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s request to %s: %v", req.Method, req.URL, err)
	}

	defer func() {
		if err != nil {
			resp.Body.Close()
		}
	}()

	for _, statusCode := range expectedStatusCodes {
		if resp.StatusCode == statusCode {
			return resp, nil
		}
	}

	return resp, fmt.Errorf("%s request to %s: unexpected status code %d", req.Method, req.URL, resp.StatusCode)
}
