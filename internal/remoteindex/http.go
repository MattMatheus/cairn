package remoteindex

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type HTTPClient struct {
	BaseURL string
	Client  *http.Client
	Token   func(context.Context) (string, error)
}

func (c HTTPClient) Status(ctx context.Context, req StatusRequest) (StatusResponse, error) {
	var out StatusResponse
	err := c.do(ctx, http.MethodPost, "/index/status", req, &out)
	return out, err
}

func (c HTTPClient) Refresh(ctx context.Context, req RefreshRequest) (RefreshResponse, error) {
	var out RefreshResponse
	err := c.do(ctx, http.MethodPost, "/index/refresh", req, &out)
	return out, err
}

func (c HTTPClient) Search(ctx context.Context, req SearchRequest) (SearchResponse, error) {
	var out SearchResponse
	err := c.do(ctx, http.MethodPost, "/search", req, &out)
	return out, err
}

func (c HTTPClient) do(ctx context.Context, method string, path string, in any, out any) error {
	body, err := json.Marshal(in)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(c.BaseURL, "/")+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	if c.Token != nil {
		token, err := c.Token(ctx)
		if err != nil {
			return err
		}
		if token != "" {
			request.Header.Set("Authorization", "Bearer "+token)
		}
	}
	response, err := c.client().Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		content, _ := io.ReadAll(io.LimitReader(response.Body, 1024))
		return fmt.Errorf("remote indexer request failed: %s: %s", response.Status, strings.TrimSpace(string(content)))
	}
	return json.NewDecoder(response.Body).Decode(out)
}

func (c HTTPClient) client() *http.Client {
	if c.Client != nil {
		return c.Client
	}
	return http.DefaultClient
}
