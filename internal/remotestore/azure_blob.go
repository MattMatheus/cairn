package remotestore

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"strings"

	"cairn/internal/syncstate"
)

type TokenProvider interface {
	Token(ctx context.Context) (string, error)
}

type AzureCLITokenProvider struct{}

func (AzureCLITokenProvider) Token(ctx context.Context) (string, error) {
	command := exec.CommandContext(ctx, "az", "account", "get-access-token", "--resource", "https://storage.azure.com/", "--query", "accessToken", "--output", "tsv")
	output, err := command.Output()
	if err != nil {
		return "", err
	}
	token := strings.TrimSpace(string(output))
	if token == "" {
		return "", errors.New("azure cli returned an empty storage token")
	}
	return token, nil
}

type AzureBlobConfig struct {
	Account   string
	Endpoint  string
	Container string
	Prefix    string
}

type AzureBlobStore struct {
	Config        AzureBlobConfig
	TokenProvider TokenProvider
	Client        *http.Client
}

func NewAzureBlobStore(config AzureBlobConfig, provider TokenProvider) (*AzureBlobStore, error) {
	if config.Account == "" && config.Endpoint == "" {
		return nil, errors.New("azure blob account or endpoint is required")
	}
	if config.Container == "" {
		return nil, errors.New("azure blob container is required")
	}
	if provider == nil {
		provider = AzureCLITokenProvider{}
	}
	return &AzureBlobStore{Config: config, TokenProvider: provider, Client: http.DefaultClient}, nil
}

func (s *AzureBlobStore) ReadManifest(ctx context.Context) (syncstate.Manifest, bool, error) {
	content, ok, err := s.ReadObject(ctx, RemoteManifestPath)
	if err != nil || !ok {
		return syncstate.Manifest{}, ok, err
	}
	var manifest syncstate.Manifest
	if err := json.Unmarshal(content, &manifest); err != nil {
		return syncstate.Manifest{}, false, err
	}
	return manifest, true, nil
}

func (s *AzureBlobStore) WriteManifest(ctx context.Context, manifest syncstate.Manifest) error {
	content, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return s.WriteObject(ctx, RemoteManifestPath, append(content, '\n'))
}

func (s *AzureBlobStore) ReadObject(ctx context.Context, path string) ([]byte, bool, error) {
	req, err := s.request(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, false, err
	}
	resp, err := s.client().Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, false, nil
	}
	if err := requireSuccess(resp); err != nil {
		return nil, false, err
	}
	content, err := io.ReadAll(resp.Body)
	return content, true, err
}

func (s *AzureBlobStore) WriteObject(ctx context.Context, path string, content []byte) error {
	req, err := s.request(ctx, http.MethodPut, path, bytes.NewReader(content))
	if err != nil {
		return err
	}
	req.Header.Set("x-ms-blob-type", "BlockBlob")
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(content)))
	resp, err := s.client().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return requireSuccess(resp)
}

func (s *AzureBlobStore) ListObjects(ctx context.Context, prefix string) ([]ObjectInfo, error) {
	token, err := s.TokenProvider.Token(ctx)
	if err != nil {
		return nil, err
	}
	listURL, err := url.Parse(s.containerURL())
	if err != nil {
		return nil, err
	}
	query := listURL.Query()
	query.Set("restype", "container")
	query.Set("comp", "list")
	if joinedPrefix := JoinPrefix(s.Config.Prefix, prefix); joinedPrefix != "" {
		query.Set("prefix", joinedPrefix)
	}
	listURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, listURL.String(), nil)
	if err != nil {
		return nil, err
	}
	authorize(req, token)
	resp, err := s.client().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := requireSuccess(resp); err != nil {
		return nil, err
	}
	var parsed blobList
	if err := xml.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	objects := make([]ObjectInfo, 0, len(parsed.Blobs))
	for _, blob := range parsed.Blobs {
		objects = append(objects, ObjectInfo{
			Path: StripPrefix(s.Config.Prefix, blob.Name),
			Size: blob.Properties.ContentLength,
		})
	}
	return objects, nil
}

func (s *AzureBlobStore) request(ctx context.Context, method string, workspacePath string, body io.Reader) (*http.Request, error) {
	token, err := s.TokenProvider.Token(ctx)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, s.ObjectURL(workspacePath), body)
	if err != nil {
		return nil, err
	}
	authorize(req, token)
	return req, nil
}

func (s *AzureBlobStore) ObjectName(workspacePath string) string {
	return JoinPrefix(s.Config.Prefix, workspacePath)
}

func (s *AzureBlobStore) ObjectURL(workspacePath string) string {
	return s.containerURL() + "/" + escapePath(s.ObjectName(workspacePath))
}

func (s *AzureBlobStore) containerURL() string {
	endpoint := strings.TrimRight(s.Config.Endpoint, "/")
	if endpoint == "" {
		endpoint = fmt.Sprintf("https://%s.blob.core.windows.net", s.Config.Account)
	}
	return endpoint + "/" + pathEscape(s.Config.Container)
}

func (s *AzureBlobStore) client() *http.Client {
	if s.Client != nil {
		return s.Client
	}
	return http.DefaultClient
}

func authorize(req *http.Request, token string) {
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("x-ms-version", "2023-11-03")
}

func requireSuccess(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	content, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	return fmt.Errorf("azure blob request failed: %s: %s", resp.Status, strings.TrimSpace(string(content)))
}

func escapePath(value string) string {
	parts := strings.Split(CleanPath(value), "/")
	for i, part := range parts {
		parts[i] = pathEscape(part)
	}
	return strings.Join(parts, "/")
}

func pathEscape(value string) string {
	return strings.ReplaceAll(url.PathEscape(value), "+", "%20")
}

type blobList struct {
	Blobs []blobListItem `xml:"Blobs>Blob"`
}

type blobListItem struct {
	Name       string             `xml:"Name"`
	Properties blobListProperties `xml:"Properties"`
}

type blobListProperties struct {
	ContentLength int64 `xml:"Content-Length"`
}
