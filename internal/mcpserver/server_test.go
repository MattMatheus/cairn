package mcpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cairn/internal/localindex"
	"cairn/internal/mcpops"
	"cairn/internal/mcpschema"
)

func TestServerRegistersReadOnlyToolsOnly(t *testing.T) {
	server := New(&mcpops.Local{Root: t.TempDir()})
	var names []mcpschema.ToolName
	for _, tool := range server.Tools() {
		names = append(names, tool.Name)
		if tool.InputSchema["type"] != "object" {
			t.Fatalf("tool %s missing object schema: %#v", tool.Name, tool.InputSchema)
		}
	}
	want := []mcpschema.ToolName{
		mcpschema.ToolFindDocument,
		mcpschema.ToolGetBootstrap,
		mcpschema.ToolIndexStatus,
		mcpschema.ToolListDocuments,
		mcpschema.ToolReadDocument,
		mcpschema.ToolSearchContext,
	}
	if strings.Join(toolNames(names), ",") != strings.Join(toolNames(want), ",") {
		t.Fatalf("tools = %#v want %#v", names, want)
	}
	forbidden := []mcpschema.ToolName{
		mcpschema.ToolCaptureNote,
		mcpschema.ToolPromoteDocument,
		mcpschema.ToolArchiveDocument,
		mcpschema.ToolSyncPull,
		mcpschema.ToolSyncPush,
		mcpschema.ToolIndexRefresh,
	}
	for _, name := range forbidden {
		if _, err := server.CallTool(context.Background(), name, nil); err == nil {
			t.Fatalf("forbidden tool %s was callable", name)
		}
	}
}

func TestServerCallToolPreservesEnvelope(t *testing.T) {
	root := t.TempDir()
	local := indexedLocal(t, root)
	defer local.Close()
	server := New(local)

	result, err := server.CallTool(context.Background(), mcpschema.ToolSearchContext, json.RawMessage(`{"query":"alpha","mode":"full_text"}`))
	if err != nil {
		t.Fatalf("CallTool() error = %v", err)
	}
	envelope, ok := result.(mcpschema.Envelope[mcpschema.SearchContextData])
	if !ok {
		t.Fatalf("unexpected result type %T", result)
	}
	if !envelope.OK || len(envelope.Data.Results) != 1 {
		t.Fatalf("unexpected envelope %#v", envelope)
	}
	if envelope.Data.Results[0].Path != "working/searchable.md" {
		t.Fatalf("unexpected result path %#v", envelope.Data.Results)
	}
}

func TestServerJSONRPCListAndCall(t *testing.T) {
	root := t.TempDir()
	local := indexedLocal(t, root)
	defer local.Close()
	server := New(local)

	list := server.Handle(context.Background(), Request{JSONRPC: "2.0", ID: float64(1), Method: "tools/list"})
	if list.Error != nil {
		t.Fatalf("tools/list error %#v", list.Error)
	}
	listJSON, _ := json.Marshal(list.Result)
	if !strings.Contains(string(listJSON), string(mcpschema.ToolReadDocument)) {
		t.Fatalf("read_document missing from list: %s", string(listJSON))
	}

	call := server.Handle(context.Background(), Request{
		JSONRPC: "2.0",
		ID:      float64(2),
		Method:  "tools/call",
		Params:  json.RawMessage(`{"name":"read_document","arguments":{"path":"working/searchable.md","mode":"frontmatter"}}`),
	})
	if call.Error != nil {
		t.Fatalf("tools/call error %#v", call.Error)
	}
	result, ok := call.Result.(map[string]any)
	if !ok {
		t.Fatalf("unexpected call result type %T", call.Result)
	}
	content := result["content"].([]map[string]string)
	if !strings.Contains(content[0]["text"], `"ok":true`) || !strings.Contains(content[0]["text"], "Searchable") {
		t.Fatalf("unexpected call result: %#v", call.Result)
	}
}

func TestServerServeLineDelimitedJSONRPC(t *testing.T) {
	server := New(&mcpops.Local{Root: t.TempDir()})
	input := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize"}` + "\n" + `{"jsonrpc":"2.0","id":2,"method":"tools/list"}` + "\n")
	var output bytes.Buffer

	if err := server.Serve(context.Background(), input, &output); err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("unexpected output lines: %q", output.String())
	}
	if !strings.Contains(lines[0], "protocolVersion") || !strings.Contains(lines[1], "tools") {
		t.Fatalf("unexpected output: %q", output.String())
	}
}

func indexedLocal(t *testing.T, root string) *mcpops.Local {
	t.Helper()
	writeFile(t, root, "working/searchable.md", validDoc("cairn:search", "Searchable", "searchable", "alpha body"))
	index, err := localindex.Open(root)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if _, err := index.IndexWorkspace(context.Background(), root); err != nil {
		t.Fatalf("IndexWorkspace() error = %v", err)
	}
	return &mcpops.Local{Root: root, Index: index}
}

func writeFile(t *testing.T, root string, rel string, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

func validDoc(id string, title string, slug string, body string) string {
	return `---
id: ` + id + `
schema_version: 1
title: ` + title + `
slug: ` + slug + `
type: note
status: working
created: 2026-05-03T00:00:00Z
updated: 2026-05-03T00:00:00Z
authors: [foundry]
actors: [codex]
source: test
tags: [test]
---
# ` + title + `

` + body + `
`
}

func toolNames(names []mcpschema.ToolName) []string {
	out := make([]string, 0, len(names))
	for _, name := range names {
		out = append(out, string(name))
	}
	return out
}
