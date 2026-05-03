package mcpschema

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestV1ToolsMatchAcceptedADRSurface(t *testing.T) {
	want := []ToolName{
		ToolGetBootstrap,
		ToolCaptureNote,
		ToolPromoteDocument,
		ToolArchiveDocument,
		ToolReadDocument,
		ToolFindDocument,
		ToolSearchContext,
		ToolListDocuments,
		ToolValidateWorkspace,
		ToolSyncStatus,
		ToolSyncPull,
		ToolSyncPush,
		ToolIndexStatus,
		ToolIndexRefresh,
	}

	definitions := V1Tools()
	if len(definitions) != len(want) {
		t.Fatalf("unexpected tool count: got %d want %d", len(definitions), len(want))
	}

	seen := map[ToolName]bool{}
	for _, definition := range definitions {
		seen[definition.Name] = true
		if definition.Mutability == "" {
			t.Fatalf("tool %s is missing mutability", definition.Name)
		}
		if definition.Purpose == "" {
			t.Fatalf("tool %s is missing purpose", definition.Name)
		}
	}
	for _, name := range want {
		if !seen[name] {
			t.Fatalf("missing ADR tool %s", name)
		}
	}
}

func TestEveryV1ToolHasConcreteRequestAndResponseSchema(t *testing.T) {
	definitions := V1SchemaDefinitions()
	if len(definitions) != len(V1Tools()) {
		t.Fatalf("schema definition count must match tool count: got %d want %d", len(definitions), len(V1Tools()))
	}

	seen := map[ToolName]bool{}
	for _, definition := range definitions {
		if definition.RequestType == "" {
			t.Fatalf("%s is missing request schema", definition.Tool)
		}
		if definition.ResponseType == "" {
			t.Fatalf("%s is missing response schema", definition.Tool)
		}
		if !strings.HasPrefix(definition.ResponseType, "Envelope[") {
			t.Fatalf("%s response must use common envelope, got %s", definition.Tool, definition.ResponseType)
		}
		seen[definition.Tool] = true
	}
	for _, tool := range V1Tools() {
		if !seen[tool.Name] {
			t.Fatalf("tool %s is missing schema definition", tool.Name)
		}
	}
}

func TestV1ToolsExcludeHardDeleteAndPurge(t *testing.T) {
	for _, definition := range V1Tools() {
		name := string(definition.Name)
		if strings.Contains(name, "delete") || strings.Contains(name, "purge") {
			t.Fatalf("MCP v1 must not expose hard delete or purge, got %s", name)
		}
	}
}

func TestEnvelopeCarriesWarningsUnavailableNextStepsAndProvenance(t *testing.T) {
	response := Envelope[SearchContextData]{
		OK: true,
		Data: SearchContextData{
			Results: []SearchResult{{
				Path:      "runbooks/auth-timeouts.md",
				Title:     "Auth Timeout Runbook",
				Type:      "runbook",
				Status:    "canonical",
				Slug:      "auth-timeout-runbook",
				Tags:      []string{"auth", "timeouts"},
				Updated:   time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC),
				Score:     0.91,
				MatchType: SearchModeMetadata,
				Snippet:   "Retry auth token requests with bounded exponential backoff.",
				Provenance: ItemProvenance{
					Authors: []string{"matt"},
					Actors:  []string{"codex"},
					Source:  "promotion",
				},
			}},
			AttemptedModes: []SearchMode{SearchModeMetadata, SearchModeFullText, SearchModeSemantic},
		},
		Warnings: []Warning{{
			Code:    WarningIndexDegraded,
			Message: "semantic search unavailable; returned local results",
		}},
		Unavailable: []UnavailableMode{{
			Mode:      string(SearchModeSemantic),
			Reason:    WarningRemoteService,
			Message:   "remote indexer is not configured",
			Retryable: false,
		}},
		NextSteps: []NextStep{{
			Action: "index_status",
			Label:  "Check index availability",
		}},
		Provenance: ResponseProvenance{
			Profile:        ProfileLocal,
			WorkspaceID:    "pod-1",
			GeneratedAt:    time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC),
			Source:         "local_index",
			AttemptedModes: []string{"metadata", "full_text", "semantic"},
		},
	}

	content, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}
	jsonText := string(content)
	for _, fragment := range []string{
		`"ok":true`,
		`"warnings":[`,
		`"unavailable":[`,
		`"next_steps":[`,
		`"provenance":`,
		`"attempted_modes":["metadata","full_text","semantic"]`,
		`"match_type":"metadata"`,
	} {
		if !strings.Contains(jsonText, fragment) {
			t.Fatalf("expected JSON fragment %s in %s", fragment, jsonText)
		}
	}
}

func TestMutationSchemasExposeChangedPathsAndDurableIDs(t *testing.T) {
	for name, response := range map[ToolName]MutationResult{
		ToolCaptureNote: {
			DocumentID:   "cairn:01JZ8K4Z8V5X9R8P9M6T2Q3N4A",
			ChangedPaths: []ChangedPath{{Path: "agents/codex/example.md", Kind: "created"}},
		},
		ToolPromoteDocument: {
			DocumentID:   "cairn:01JZ8K4Z8V5X9R8P9M6T2Q3N4A",
			ChangedPaths: []ChangedPath{{Path: "notes/example.md", PreviousPath: "agents/codex/example.md", Kind: "moved"}},
		},
		ToolArchiveDocument: {
			DocumentID:   "cairn:01JZ8K4Z8V5X9R8P9M6T2Q3N4A",
			ChangedPaths: []ChangedPath{{Path: "archive/notes/example.md", PreviousPath: "notes/example.md", Kind: "archived"}},
		},
		ToolSyncPull: {
			ChangedPaths: []ChangedPath{{Path: ".cairn/sync-state.json", Kind: "updated"}},
		},
		ToolSyncPush: {
			ChangedPaths: []ChangedPath{{Path: ".cairn/remote-manifest.json", Kind: "updated"}},
		},
		ToolIndexRefresh: {
			ChangedPaths: []ChangedPath{{Path: ".cairn/index/cairn.db", Kind: "updated"}},
		},
	} {
		if len(response.ChangedPaths) == 0 {
			t.Fatalf("%s mutation response must include changed paths", name)
		}
		if name == ToolCaptureNote || name == ToolPromoteDocument || name == ToolArchiveDocument {
			if response.DocumentID == "" {
				t.Fatalf("%s mutation response must include durable document id", name)
			}
		}
	}
}

func TestReadAndSearchModesSupportProgressiveDisclosure(t *testing.T) {
	readModes := []ReadMode{ReadModeSummary, ReadModeFrontmatter, ReadModeTOC, ReadModeSections, ReadModeFull}
	searchModes := []SearchMode{SearchModeAuto, SearchModeMetadata, SearchModeFullText, SearchModeSemantic}

	if len(readModes) != 5 {
		t.Fatalf("unexpected read mode count: %d", len(readModes))
	}
	if len(searchModes) != 4 {
		t.Fatalf("unexpected search mode count: %d", len(searchModes))
	}

	readRequest := ReadDocumentRequest{
		DocumentRef: DocumentRef{Path: "docs/service.md"},
		Mode:        ReadModeSections,
		Sections:    []string{"Operations"},
	}
	searchRequest := SearchContextRequest{
		Query: "auth timeout",
		Mode:  SearchModeAuto,
		Limit: 5,
	}
	if readRequest.Mode != ReadModeSections || len(readRequest.Sections) != 1 {
		t.Fatalf("unexpected read request: %#v", readRequest)
	}
	if searchRequest.Mode != SearchModeAuto {
		t.Fatalf("unexpected search request: %#v", searchRequest)
	}
}
