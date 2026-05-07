package remotestore

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"cairn/internal/syncstate"
)

type staticToken string

func (s staticToken) Token(context.Context) (string, error) {
	return string(s), nil
}

func TestAzureBlobPathMapping(t *testing.T) {
	store, err := NewAzureBlobStore(AzureBlobConfig{
		Account:   "acct",
		Container: "cairn",
		Prefix:    "pod-a",
	}, staticToken("token"))
	if err != nil {
		t.Fatalf("NewAzureBlobStore() error = %v", err)
	}
	if got := store.ObjectName("working/a file.md"); got != "pod-a/working/a file.md" {
		t.Fatalf("ObjectName() = %q", got)
	}
	if got := store.ObjectURL("working/a file.md"); got != "https://acct.blob.core.windows.net/cairn/pod-a/working/a%20file.md" {
		t.Fatalf("ObjectURL() = %q", got)
	}
}

func TestAzureBlobStoreUsesBearerTokenAndMapsObjects(t *testing.T) {
	var requests []string
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		requests = append(requests, r.Method+" "+r.URL.String())
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatalf("missing bearer token: %#v", r.Header)
		}
		switch {
		case r.Method == http.MethodPut && r.URL.Path == "/cairn/prefix/working/a.md":
			content, _ := io.ReadAll(r.Body)
			if string(content) != "hello" {
				t.Fatalf("unexpected body %q", string(content))
			}
			return response(http.StatusCreated, ""), nil
		case r.Method == http.MethodGet && r.URL.Path == "/cairn/prefix/working/a.md":
			return response(http.StatusOK, "hello"), nil
		case r.Method == http.MethodGet && r.URL.Path == "/cairn" && r.URL.Query().Get("comp") == "list":
			if r.URL.Query().Get("prefix") != "prefix/working" {
				t.Fatalf("unexpected list prefix %q", r.URL.Query().Get("prefix"))
			}
			return response(http.StatusOK, `<?xml version="1.0" encoding="utf-8"?>
<EnumerationResults>
  <Blobs>
    <Blob>
      <Name>prefix/working/a.md</Name>
      <Properties><Content-Length>5</Content-Length></Properties>
    </Blob>
  </Blobs>
</EnumerationResults>`), nil
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.String())
			return response(http.StatusInternalServerError, ""), nil
		}
	})

	store, err := NewAzureBlobStore(AzureBlobConfig{
		Endpoint:  "https://example.test",
		Container: "cairn",
		Prefix:    "prefix",
	}, staticToken("test-token"))
	if err != nil {
		t.Fatalf("NewAzureBlobStore() error = %v", err)
	}
	store.Client = &http.Client{Transport: transport}

	ctx := context.Background()
	if err := store.WriteObject(ctx, "working/a.md", []byte("hello")); err != nil {
		t.Fatalf("WriteObject() error = %v", err)
	}
	content, ok, err := store.ReadObject(ctx, "working/a.md")
	if err != nil || !ok || string(content) != "hello" {
		t.Fatalf("ReadObject() = %q ok=%t err=%v", string(content), ok, err)
	}
	objects, err := store.ListObjects(ctx, "working")
	if err != nil {
		t.Fatalf("ListObjects() error = %v", err)
	}
	if len(objects) != 1 || objects[0].Path != "working/a.md" || objects[0].Size != 5 {
		t.Fatalf("unexpected objects %#v", objects)
	}
	if strings.Join(requests, "\n") != "PUT https://example.test/cairn/prefix/working/a.md\nGET https://example.test/cairn/prefix/working/a.md\nGET https://example.test/cairn?comp=list&prefix=prefix%2Fworking&restype=container" {
		t.Fatalf("unexpected requests:\n%s", strings.Join(requests, "\n"))
	}
}

func TestAzureBlobStoreReadMissingObject(t *testing.T) {
	store, err := NewAzureBlobStore(AzureBlobConfig{Endpoint: "https://example.test", Container: "cairn"}, staticToken("token"))
	if err != nil {
		t.Fatalf("NewAzureBlobStore() error = %v", err)
	}
	store.Client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return response(http.StatusNotFound, ""), nil
	})}
	_, ok, err := store.ReadObject(context.Background(), "missing.md")
	if err != nil {
		t.Fatalf("ReadObject() error = %v", err)
	}
	if ok {
		t.Fatalf("expected missing object")
	}
}

func TestAzureBlobStoreManifestRoundTripUsesRemoteManifestPath(t *testing.T) {
	var stored []byte
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/cairn/prefix/.cairn/remote-manifest.json" {
			t.Fatalf("unexpected manifest path %s", r.URL.Path)
		}
		switch r.Method {
		case http.MethodPut:
			var err error
			stored, err = io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("ReadAll() error = %v", err)
			}
			return response(http.StatusCreated, ""), nil
		case http.MethodGet:
			return response(http.StatusOK, string(stored)), nil
		default:
			t.Fatalf("unexpected method %s", r.Method)
			return response(http.StatusInternalServerError, ""), nil
		}
	})
	store, err := NewAzureBlobStore(AzureBlobConfig{Endpoint: "https://example.test", Container: "cairn", Prefix: "prefix"}, staticToken("token"))
	if err != nil {
		t.Fatalf("NewAzureBlobStore() error = %v", err)
	}
	store.Client = &http.Client{Transport: transport}
	manifest := syncstate.Manifest{ManifestVersion: syncstate.ManifestVersion, WorkspaceID: "pod-1"}
	if err := store.WriteManifest(context.Background(), manifest); err != nil {
		t.Fatalf("WriteManifest() error = %v", err)
	}
	if !json.Valid(stored) {
		t.Fatalf("stored manifest is not JSON: %s", string(stored))
	}
	got, ok, err := store.ReadManifest(context.Background())
	if err != nil {
		t.Fatalf("ReadManifest() error = %v", err)
	}
	if !ok || got.WorkspaceID != "pod-1" {
		t.Fatalf("unexpected manifest ok=%t got=%#v", ok, got)
	}
}

func TestAzureBlobStoreAzuriteAuthUsesSharedKey(t *testing.T) {
	store, err := NewAzureBlobStore(AzureBlobConfig{
		Endpoint:  "http://localhost:10000/devstoreaccount1",
		Container: "cairn",
		AuthMode:  "azurite",
	}, nil)
	if err != nil {
		t.Fatalf("NewAzureBlobStore() error = %v", err)
	}
	store.Client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if !strings.HasPrefix(r.Header.Get("Authorization"), "SharedKey devstoreaccount1:") {
			t.Fatalf("missing azurite shared key auth: %#v", r.Header)
		}
		if r.Header.Get("x-ms-date") == "" || r.Header.Get("x-ms-version") == "" {
			t.Fatalf("missing shared key headers: %#v", r.Header)
		}
		return response(http.StatusCreated, ""), nil
	})}
	if err := store.WriteObject(context.Background(), "working/a.md", []byte("hello")); err != nil {
		t.Fatalf("WriteObject() error = %v", err)
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
