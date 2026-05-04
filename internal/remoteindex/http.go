package remoteindex

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
)

type HTTPClient struct {
	BaseURL string
	Client  *http.Client
	Token   func(context.Context) (string, error)
}

func AzureCLIToken(audience string) func(context.Context) (string, error) {
	return func(ctx context.Context) (string, error) {
		if audience == "" {
			return "", nil
		}
		command := exec.CommandContext(ctx, "az", "account", "get-access-token", "--resource", audience, "--query", "accessToken", "--output", "tsv")
		output, err := command.Output()
		if err != nil {
			return "", err
		}
		token := strings.TrimSpace(string(output))
		if token == "" {
			return "", errors.New("azure cli returned an empty indexer token")
		}
		return token, nil
	}
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
