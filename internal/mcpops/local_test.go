package mcpops

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"cairn/internal/localindex"
	"cairn/internal/mcpschema"
	"cairn/internal/remoteindex"
	"cairn/internal/remotestore"
)

func TestLocalSearchContextUsesLocalOnlyAutoSearch(t *testing.T) {
	ops := newFixtureOps(t)
	envelope, err := ops.SearchContext(context.Background(), mcpschema.SearchContextRequest{
		Query: "auth",
		Mode:  mcpschema.SearchModeAuto,
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("SearchContext returned error: %v", err)
	}
	if !envelope.OK || len(envelope.Data.Results) == 0 {
		t.Fatalf("expected search results envelope: %#v", envelope)
	}
	if len(envelope.Data.AttemptedModes) != 2 ||
		envelope.Data.AttemptedModes[0] != mcpschema.SearchModeMetadata ||
		envelope.Data.AttemptedModes[1] != mcpschema.SearchModeFullText {
		t.Fatalf("expected metadata and full-text attempts only: %#v", envelope.Data.AttemptedModes)
	}
	if len(envelope.Warnings) != 0 || len(envelope.Unavailable) != 0 || len(envelope.NextSteps) != 0 {
		t.Fatalf("local-only auto search should not report remote degradation: %#v", envelope)
	}
	if envelope.Provenance.Profile != mcpschema.ProfileLocal {
		t.Fatalf("expected local provenance: %#v", envelope.Provenance)
	}
}

func TestLocalSearchContextUsesRemoteIndexWhenConfigured(t *testing.T) {
	ops := newFixtureOps(t)
	remote := &remoteindex.FakeClient{
		SearchResponse: remoteindex.SearchResponse{
			Results: []remoteindex.SearchResult{{Path: "runbooks/semantic.md", Title: "Semantic", Source: "remote_index"}},
		},
	}
	ops.RemoteIndex = remote

	envelope, err := ops.SearchContext(context.Background(), mcpschema.SearchContextRequest{
		Query: "semantic",
		Mode:  mcpschema.SearchModeSemantic,
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("SearchContext returned error: %v", err)
	}
	if len(remote.Calls) != 1 || len(envelope.Data.Results) != 1 {
		t.Fatalf("expected remote search integration, calls=%#v envelope=%#v", remote.Calls, envelope)
	}
	if envelope.Data.Results[0].MatchType != mcpschema.SearchModeSemantic {
		t.Fatalf("expected semantic result shape: %#v", envelope.Data.Results[0])
	}
}

func TestListAndFindDocumentsUseLocalMetadata(t *testing.T) {
	ops := newFixtureOps(t)
	listed, err := ops.ListDocuments(context.Background(), mcpschema.ListDocumentsRequest{
		Type:  "runbook",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("ListDocuments returned error: %v", err)
	}
	if len(listed.Data.Documents) != 1 || listed.Data.Documents[0].Path != "runbooks/auth.md" {
		t.Fatalf("unexpected list response: %#v", listed)
	}

	for name, req := range map[string]mcpschema.FindDocumentRequest{
		"id":     {ID: "cairn:auth"},
		"slug":   {Slug: "auth-timeout-runbook"},
		"title":  {Title: "Auth Timeout"},
		"path":   {Path: "runbooks/auth.md"},
		"type":   {Type: "runbook"},
		"status": {Status: "canonical"},
		"tag":    {Tag: "auth"},
	} {
		found, err := ops.FindDocument(context.Background(), req)
		if err != nil {
			t.Fatalf("FindDocument(%s) returned error: %v", name, err)
		}
		if len(found.Data.Documents) == 0 || found.Data.Documents[0].Path != "runbooks/auth.md" {
			t.Fatalf("FindDocument(%s) unexpected response: %#v", name, found)
		}
	}
}

func TestIndexStatusAndBootstrapAreLocalOnly(t *testing.T) {
	ops := newFixtureOps(t)
	status, err := ops.IndexStatus(context.Background(), mcpschema.IndexStatusRequest{})
	if err != nil {
		t.Fatalf("IndexStatus returned error: %v", err)
	}
	if !status.Data.LocalAvailable || status.Data.RemoteAvailable || !status.Data.Fresh {
		t.Fatalf("unexpected index status: %#v", status)
	}
	if len(status.Warnings) != 0 || len(status.Unavailable) != 0 || len(status.NextSteps) != 1 {
		t.Fatalf("expected clean local-only index status with local next step: %#v", status)
	}

	bootstrap, err := ops.GetBootstrap(context.Background(), mcpschema.EmptyRequest{})
	if err != nil {
		t.Fatalf("GetBootstrap returned error: %v", err)
	}
	if !bootstrap.OK || bootstrap.Data.Summary == "" || len(bootstrap.NextSteps) == 0 {
		t.Fatalf("unexpected bootstrap: %#v", bootstrap)
	}
}

func TestIndexStatusSuggestsRefreshWhenLocalIndexMissing(t *testing.T) {
	root := t.TempDir()
	ops := &Local{Root: root}

	status, err := ops.IndexStatus(context.Background(), mcpschema.IndexStatusRequest{})
	if err != nil {
		t.Fatalf("IndexStatus returned error: %v", err)
	}
	if status.Data.LocalAvailable || status.Data.Fresh {
		t.Fatalf("local index should be missing before refresh: %#v", status)
	}
	if len(status.NextSteps) != 1 || status.NextSteps[0].Action != string(mcpschema.ToolIndexRefresh) {
		t.Fatalf("missing local index should suggest refresh: %#v", status.NextSteps)
	}
}

func TestIndexStatusReportsConfiguredRemoteIndexer(t *testing.T) {
	ops := newFixtureOps(t)
	lastRefresh := time.Date(2026, 5, 6, 13, 30, 0, 0, time.UTC)
	ops.RemoteIndex = &remoteindex.FakeClient{
		StatusResponse: remoteindex.StatusResponse{
			Available:     true,
			Fresh:         true,
			LastRefreshAt: lastRefresh,
			IndexedCount:  4,
		},
	}
	status, err := ops.IndexStatus(context.Background(), mcpschema.IndexStatusRequest{})
	if err != nil {
		t.Fatalf("IndexStatus returned error: %v", err)
	}
	if !status.Data.LocalAvailable || !status.Data.RemoteAvailable || !status.Data.Fresh || status.Data.LastRefreshAt.Before(lastRefresh) {
		t.Fatalf("unexpected remote index status: %#v", status)
	}
	if len(status.Warnings) != 0 || len(status.Unavailable) != 0 {
		t.Fatalf("remote index status should not degrade when available: %#v", status)
	}
}

func TestIndexStatusRemoteFailureSuggestsStatusRetry(t *testing.T) {
	ops := newFixtureOps(t)
	ops.RemoteIndex = failingRemoteIndex{}

	status, err := ops.IndexStatus(context.Background(), mcpschema.IndexStatusRequest{})
	if err != nil {
		t.Fatalf("IndexStatus returned error: %v", err)
	}
	if status.OK != true || len(status.Warnings) != 1 || len(status.Unavailable) != 1 {
		t.Fatalf("expected remote status degradation without local failure: %#v", status)
	}
	if len(status.NextSteps) != 1 || status.NextSteps[0].Action != string(mcpschema.ToolIndexStatus) {
		t.Fatalf("remote status failure should suggest status retry: %#v", status.NextSteps)
	}
}

func TestOpenLocalConfiguresRemoteClientsFromWorkspaceConfig(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".cairn/config.yaml", `schema_version: 1
workspace_id: cairn:workspace:test
managed_folders:
  - working
document_types:
  note: working
remote_sync:
  provider: azure_blob
  account: acct
  container: cairn
  prefix: pod-a
remote_index:
  url: https://indexer.example
  audience: api://cairn-indexer
  tenant_id: tenant
`)
	local, err := OpenLocal(root)
	if err != nil {
		t.Fatalf("OpenLocal() error = %v", err)
	}
	defer local.Close()
	store, ok := local.RemoteStore.(*remotestore.AzureBlobStore)
	if !ok {
		t.Fatalf("expected AzureBlobStore, got %T", local.RemoteStore)
	}
	if store.Config.Account != "acct" || store.Config.Container != "cairn" || store.Config.Prefix != "pod-a" {
		t.Fatalf("unexpected store config: %#v", store.Config)
	}
	client, ok := local.RemoteIndex.(remoteindex.HTTPClient)
	if !ok {
		t.Fatalf("expected remoteindex.HTTPClient, got %T", local.RemoteIndex)
	}
	if client.BaseURL != "https://indexer.example" || client.Token == nil {
		t.Fatalf("unexpected remote index client: %#v", client)
	}
}

func TestOpenLocalConfiguresRemoteClientsFromGeneratedPodRemoteProfile(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".cairn/config.yaml", `schema_version: 1
workspace_id: cairn:workspace:test
managed_folders:
  - working
document_types:
  note: working
profiles:
  local:
    enabled: true
  pod-remote:
    enabled: true
    provider: azure_blob
    account: acct
    container: cairn
    prefix: pod-a
    url: https://indexer.example
    audience: api://cairn-indexer
    tenant_id: tenant
`)
	local, err := OpenLocal(root)
	if err != nil {
		t.Fatalf("OpenLocal() error = %v", err)
	}
	defer local.Close()
	store, ok := local.RemoteStore.(*remotestore.AzureBlobStore)
	if !ok {
		t.Fatalf("expected AzureBlobStore, got %T", local.RemoteStore)
	}
	if store.Config.Account != "acct" || store.Config.Container != "cairn" || store.Config.Prefix != "pod-a" {
		t.Fatalf("unexpected store config: %#v", store.Config)
	}
	client, ok := local.RemoteIndex.(remoteindex.HTTPClient)
	if !ok {
		t.Fatalf("expected remoteindex.HTTPClient, got %T", local.RemoteIndex)
	}
	if client.BaseURL != "https://indexer.example" || client.Token == nil {
		t.Fatalf("unexpected remote index client: %#v", client)
	}
}

func TestOpenLocalConfiguresLocalFSRemoteStore(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".cairn/config.yaml", `schema_version: 1
workspace_id: cairn:workspace:test
managed_folders:
  - working
document_types:
  note: working
remote_sync:
  provider: local_fs
  root: .cairn/local-remote
  prefix: pod-a
`)
	local, err := OpenLocal(root)
	if err != nil {
		t.Fatalf("OpenLocal() error = %v", err)
	}
	defer local.Close()
	store, ok := local.RemoteStore.(*remotestore.LocalFSStore)
	if !ok {
		t.Fatalf("expected LocalFSStore, got %T", local.RemoteStore)
	}
	if store.Root != filepath.Join(root, ".cairn", "local-remote") || store.Prefix != "pod-a" {
		t.Fatalf("unexpected store config: %#v", store)
	}
}

func TestOpenLocalConfiguresAzuriteAuthMode(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".cairn/config.yaml", `schema_version: 1
workspace_id: cairn:workspace:test
managed_folders:
  - working
document_types:
  note: working
remote_sync:
  provider: azure_blob
  endpoint: http://localhost:10000/devstoreaccount1
  container: cairn
  auth_mode: azurite
`)
	local, err := OpenLocal(root)
	if err != nil {
		t.Fatalf("OpenLocal() error = %v", err)
	}
	defer local.Close()
	store, ok := local.RemoteStore.(*remotestore.AzureBlobStore)
	if !ok {
		t.Fatalf("expected AzureBlobStore, got %T", local.RemoteStore)
	}
	if store.Config.Endpoint != "http://localhost:10000/devstoreaccount1" || store.Config.AuthMode != "azurite" {
		t.Fatalf("unexpected store config: %#v", store.Config)
	}
}

func TestOpenLocalWithoutRemoteConfigRemainsLocalOnly(t *testing.T) {
	local, err := OpenLocal(t.TempDir())
	if err != nil {
		t.Fatalf("OpenLocal() error = %v", err)
	}
	defer local.Close()
	if local.RemoteStore != nil || local.RemoteIndex != nil {
		t.Fatalf("expected local-only clients, got store=%T index=%T", local.RemoteStore, local.RemoteIndex)
	}
}

func TestIndexRefreshReportsAcceptedAndRefreshedRemoteResponses(t *testing.T) {
	ops := newFixtureOps(t)
	ops.RemoteIndex = &remoteindex.FakeClient{
		RefreshResponse: remoteindex.RefreshResponse{
			Accepted:  true,
			Refreshed: false,
			JobID:     "job-1",
			Message:   "queued",
		},
	}
	accepted, err := ops.IndexRefresh(context.Background(), mcpschema.IndexRefreshRequest{})
	if err != nil {
		t.Fatalf("IndexRefresh accepted returned error: %v", err)
	}
	if !accepted.OK || !accepted.Data.Accepted || accepted.Data.RemoteRefreshed || accepted.Data.JobID != "job-1" {
		t.Fatalf("unexpected accepted refresh envelope: %#v", accepted)
	}
	if !accepted.Data.LocalRefreshed {
		t.Fatalf("expected local index refresh before optional remote refresh: %#v", accepted)
	}
	if len(accepted.NextSteps) != 1 || accepted.NextSteps[0].Action != string(mcpschema.ToolIndexStatus) {
		t.Fatalf("accepted async refresh should suggest status: %#v", accepted.NextSteps)
	}

	ops.RemoteIndex = &remoteindex.FakeClient{
		RefreshResponse: remoteindex.RefreshResponse{
			Accepted:  true,
			Refreshed: true,
			Message:   "done",
		},
	}
	refreshed, err := ops.IndexRefresh(context.Background(), mcpschema.IndexRefreshRequest{})
	if err != nil {
		t.Fatalf("IndexRefresh refreshed returned error: %v", err)
	}
	if !refreshed.OK || !refreshed.Data.Accepted || !refreshed.Data.RemoteRefreshed {
		t.Fatalf("unexpected refreshed envelope: %#v", refreshed)
	}
	if !refreshed.Data.LocalRefreshed {
		t.Fatalf("expected local index refresh before optional remote refresh: %#v", refreshed)
	}
	if len(refreshed.Data.ChangedPaths) == 0 || len(refreshed.NextSteps) != 1 || refreshed.NextSteps[0].Action != string(mcpschema.ToolSearchContext) {
		t.Fatalf("refreshed response should include changed paths and search next step: %#v", refreshed)
	}
}

func TestIndexRefreshRebuildsLocalIndexWhenRemoteIsNotConfigured(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "runbooks/local.md", managedMarkdown("cairn:local", "Local Runbook", "local-runbook", "runbook", "canonical", []string{"local"}, []string{"codex"}, []string{"matt"}, "capture", "2026-05-03T12:00:00Z", "Local body."))
	index, err := localindex.Open(root)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	defer index.Close()
	ops := &Local{Root: root, Index: index}

	envelope, err := ops.IndexRefresh(context.Background(), mcpschema.IndexRefreshRequest{})
	if err != nil {
		t.Fatalf("IndexRefresh returned error: %v", err)
	}
	if !envelope.OK || !envelope.Data.LocalRefreshed || envelope.Data.RemoteRefreshed {
		t.Fatalf("unexpected local refresh envelope: %#v", envelope)
	}
	if len(envelope.Data.ChangedPaths) != 1 || envelope.Data.ChangedPaths[0].Path != ".cairn/index/cairn.db" {
		t.Fatalf("expected local index changed path: %#v", envelope.Data.ChangedPaths)
	}
	results, err := index.Query(context.Background(), localindex.Query{Text: "Local", Limit: 10})
	if err != nil {
		t.Fatalf("Query returned error: %v", err)
	}
	if len(results) != 1 || results[0].Path != "runbooks/local.md" {
		t.Fatalf("expected refreshed local index result: %#v", results)
	}
}

func TestIndexRefreshUnavailableAndFailureDegrade(t *testing.T) {
	ops := newFixtureOps(t)
	unavailable, err := ops.IndexRefresh(context.Background(), mcpschema.IndexRefreshRequest{})
	if err != nil {
		t.Fatalf("IndexRefresh unavailable returned error: %v", err)
	}
	if !unavailable.OK || !unavailable.Data.LocalRefreshed {
		t.Fatalf("expected local refresh without remote configured: %#v", unavailable)
	}

	ops.RemoteIndex = failingRemoteIndex{}
	failed, err := ops.IndexRefresh(context.Background(), mcpschema.IndexRefreshRequest{})
	if err != nil {
		t.Fatalf("IndexRefresh failure returned error: %v", err)
	}
	if failed.OK || len(failed.Warnings) != 1 || len(failed.Unavailable) != 1 || !failed.Unavailable[0].Retryable {
		t.Fatalf("expected retryable failure degradation: %#v", failed)
	}
	if !failed.Data.LocalRefreshed {
		t.Fatalf("expected local refresh to remain available despite remote failure: %#v", failed)
	}
}

type failingRemoteIndex struct{}

func (failingRemoteIndex) Status(context.Context, remoteindex.StatusRequest) (remoteindex.StatusResponse, error) {
	return remoteindex.StatusResponse{}, context.Canceled
}

func (failingRemoteIndex) Refresh(context.Context, remoteindex.RefreshRequest) (remoteindex.RefreshResponse, error) {
	return remoteindex.RefreshResponse{}, context.Canceled
}

func (failingRemoteIndex) Search(context.Context, remoteindex.SearchRequest) (remoteindex.SearchResponse, error) {
	return remoteindex.SearchResponse{}, context.Canceled
}

func newFixtureOps(t *testing.T) *Local {
	t.Helper()
	root := t.TempDir()
	writeFile(t, root, "runbooks/auth.md", managedMarkdown("cairn:auth", "Auth Timeout Runbook", "auth-timeout-runbook", "runbook", "canonical", []string{"auth"}, []string{"codex"}, []string{"matt"}, "promotion", "2026-05-02T12:00:00Z", "Retry auth token requests with bounded exponential backoff."))
	writeFile(t, root, "notes/token.md", managedMarkdown("cairn:token", "Credential Note", "credential-note", "note", "working", []string{"auth"}, []string{"codex"}, []string{"matt"}, "capture", "2026-05-03T12:00:00Z", "Token rotation note body."))
	index, err := localindex.Open(root)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	t.Cleanup(func() { index.Close() })
	if _, err := index.IndexWorkspace(context.Background(), root); err != nil {
		t.Fatalf("IndexWorkspace returned error: %v", err)
	}
	return &Local{
		Root:  root,
		Index: index,
		Now:   func() time.Time { return time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC) },
	}
}

func writeFile(t *testing.T, root string, relativePath string, content string) {
	t.Helper()
	absolutePath := filepath.Join(root, relativePath)
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(absolutePath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
}

func managedMarkdown(id string, title string, slug string, docType string, status string, tags []string, actors []string, authors []string, source string, updated string, body string) string {
	return `---
id: ` + id + `
schema_version: 1
title: ` + title + `
slug: ` + slug + `
type: ` + docType + `
status: ` + status + `
created: 2026-05-01T12:00:00Z
updated: ` + updated + `
authors: ` + inlineArray(authors) + `
actors: ` + inlineArray(actors) + `
source: ` + source + `
tags: ` + inlineArray(tags) + `
---

# ` + title + `

` + body + `
`
}

func inlineArray(values []string) string {
	if len(values) == 0 {
		return "[]"
	}
	result := "["
	for i, value := range values {
		if i > 0 {
			result += ", "
		}
		result += value
	}
	return result + "]"
}
