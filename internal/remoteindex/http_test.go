package remoteindex

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestHTTPClientContracts(t *testing.T) {
	var requests []string
	client := HTTPClient{
		BaseURL: "https://indexer.example",
		Token:   func(context.Context) (string, error) { return "token", nil },
		Client: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			body, _ := io.ReadAll(req.Body)
			requests = append(requests, req.Method+" "+req.URL.Path+" "+string(body))
			if req.Header.Get("Authorization") != "Bearer token" {
				t.Fatalf("missing bearer token")
			}
			switch req.URL.Path {
			case "/index/status":
				return response(http.StatusOK, `{"available":true,"fresh":true,"indexed_count":3}`), nil
			case "/index/refresh":
				return response(http.StatusAccepted, `{"accepted":true,"job_id":"job-1"}`), nil
			case "/search":
				return response(http.StatusOK, `{"results":[{"path":"working/a.md","title":"A","score":0.8}],"attempted_modes":["semantic"]}`), nil
			default:
				t.Fatalf("unexpected path %s", req.URL.Path)
				return response(http.StatusNotFound, ""), nil
			}
		})},
	}
	ctx := context.Background()
	if status, err := client.Status(ctx, StatusRequest{WorkspaceID: "pod-1"}); err != nil || !status.Available {
		t.Fatalf("Status() = %#v err=%v", status, err)
	}
	if refresh, err := client.Refresh(ctx, RefreshRequest{WorkspaceID: "pod-1"}); err != nil || refresh.JobID != "job-1" {
		t.Fatalf("Refresh() = %#v err=%v", refresh, err)
	}
	if search, err := client.Search(ctx, SearchRequest{WorkspaceID: "pod-1", Query: "alpha"}); err != nil || len(search.Results) != 1 {
		t.Fatalf("Search() = %#v err=%v", search, err)
	}
	if len(requests) != 3 || !strings.Contains(requests[2], `"query":"alpha"`) {
		t.Fatalf("unexpected requests %#v", requests)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func response(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{},
	}
}
