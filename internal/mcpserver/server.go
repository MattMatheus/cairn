package mcpserver

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"cairn/internal/mcpops"
	"cairn/internal/mcpschema"
)

type Server struct {
	local *mcpops.Local
	tools map[mcpschema.ToolName]Tool
}

type Option func(*Server)

type Tool struct {
	Name        mcpschema.ToolName `json:"name"`
	Description string             `json:"description"`
	InputSchema map[string]any     `json:"inputSchema"`
	handler     func(context.Context, json.RawMessage) (any, error)
}

func New(local *mcpops.Local, opts ...Option) *Server {
	server := &Server{local: local, tools: map[mcpschema.ToolName]Tool{}}
	server.registerReadOnlyTools()
	for _, opt := range opts {
		opt(server)
	}
	return server
}

func WithLocalWrites() Option {
	return func(s *Server) {
		s.registerLocalWriteTools()
	}
}

func WithRemoteWrites() Option {
	return func(s *Server) {
		s.registerRemoteWriteTools()
	}
}

func (s *Server) Tools() []Tool {
	tools := make([]Tool, 0, len(s.tools))
	for _, tool := range s.tools {
		tools = append(tools, tool)
	}
	sort.Slice(tools, func(i, j int) bool {
		return tools[i].Name < tools[j].Name
	})
	return tools
}

func (s *Server) CallTool(ctx context.Context, name mcpschema.ToolName, arguments json.RawMessage) (any, error) {
	tool, ok := s.tools[name]
	if !ok {
		return nil, fmt.Errorf("unknown MCP tool %q", name)
	}
	return tool.handler(ctx, arguments)
}

func (s *Server) Handle(ctx context.Context, request Request) Response {
	switch request.Method {
	case "initialize":
		return resultResponse(request.ID, map[string]any{
			"protocolVersion": "2024-11-05",
			"serverInfo": map[string]any{
				"name":    "cairn",
				"version": "0.1.0",
			},
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
		})
	case "tools/list":
		return resultResponse(request.ID, map[string]any{"tools": s.Tools()})
	case "tools/call":
		var params callParams
		if err := json.Unmarshal(request.Params, &params); err != nil {
			return errorResponse(request.ID, -32602, err.Error())
		}
		result, err := s.CallTool(ctx, mcpschema.ToolName(params.Name), params.Arguments)
		if err != nil {
			return errorResponse(request.ID, -32000, err.Error())
		}
		return resultResponse(request.ID, toolResult(result))
	default:
		return errorResponse(request.ID, -32601, "method not found")
	}
}

func (s *Server) Serve(ctx context.Context, input io.Reader, output io.Writer) error {
	scanner := bufio.NewScanner(input)
	encoder := json.NewEncoder(output)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var request Request
		if err := json.Unmarshal(line, &request); err != nil {
			if err := encoder.Encode(errorResponse(nil, -32700, err.Error())); err != nil {
				return err
			}
			continue
		}
		if request.ID == nil {
			continue
		}
		if err := encoder.Encode(s.Handle(ctx, request)); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func (s *Server) registerReadOnlyTools() {
	s.register(mcpschema.ToolGetBootstrap, "Return compact workspace onboarding context and next steps.", emptySchema(), func(ctx context.Context, raw json.RawMessage) (any, error) {
		var req mcpschema.EmptyRequest
		if err := decode(raw, &req); err != nil {
			return nil, err
		}
		return s.local.GetBootstrap(ctx, req)
	})
	s.register(mcpschema.ToolSearchContext, "Search local and optional derived context.", objectSchema(map[string]any{
		"query": map[string]any{"type": "string"},
		"mode":  map[string]any{"type": "string"},
		"limit": map[string]any{"type": "integer"},
	}), func(ctx context.Context, raw json.RawMessage) (any, error) {
		var req mcpschema.SearchContextRequest
		if err := decode(raw, &req); err != nil {
			return nil, err
		}
		return s.local.SearchContext(ctx, req)
	})
	s.register(mcpschema.ToolListDocuments, "List managed documents by filters.", objectSchema(map[string]any{
		"type":   map[string]any{"type": "string"},
		"status": map[string]any{"type": "string"},
		"tags":   map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		"limit":  map[string]any{"type": "integer"},
	}), func(ctx context.Context, raw json.RawMessage) (any, error) {
		var req mcpschema.ListDocumentsRequest
		if err := decode(raw, &req); err != nil {
			return nil, err
		}
		return s.local.ListDocuments(ctx, req)
	})
	s.register(mcpschema.ToolFindDocument, "Find documents by id, slug, title, path, type, status, or tag.", objectSchema(map[string]any{
		"id":     map[string]any{"type": "string"},
		"slug":   map[string]any{"type": "string"},
		"title":  map[string]any{"type": "string"},
		"path":   map[string]any{"type": "string"},
		"type":   map[string]any{"type": "string"},
		"status": map[string]any{"type": "string"},
		"tag":    map[string]any{"type": "string"},
		"limit":  map[string]any{"type": "integer"},
	}), func(ctx context.Context, raw json.RawMessage) (any, error) {
		var req mcpschema.FindDocumentRequest
		if err := decode(raw, &req); err != nil {
			return nil, err
		}
		return s.local.FindDocument(ctx, req)
	})
	s.register(mcpschema.ToolIndexStatus, "Report local/remote index availability and freshness.", emptySchema(), func(ctx context.Context, raw json.RawMessage) (any, error) {
		var req mcpschema.IndexStatusRequest
		if err := decode(raw, &req); err != nil {
			return nil, err
		}
		return s.local.IndexStatus(ctx, req)
	})
	s.register(mcpschema.ToolSyncStatus, "Report local and remote sync state without mutating files.", emptySchema(), func(ctx context.Context, raw json.RawMessage) (any, error) {
		var req mcpschema.EmptyRequest
		if err := decode(raw, &req); err != nil {
			return nil, err
		}
		return s.local.SyncStatus(ctx, req)
	})
	s.register(mcpschema.ToolValidateWorkspace, "Validate managed markdown and sync/index metadata health.", objectSchema(map[string]any{
		"paths": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		"mode":  map[string]any{"type": "string"},
	}), func(ctx context.Context, raw json.RawMessage) (any, error) {
		var req mcpschema.ValidateWorkspaceRequest
		if err := decode(raw, &req); err != nil {
			return nil, err
		}
		return s.local.ValidateWorkspace(ctx, req)
	})
	s.register(mcpschema.ToolReadDocument, "Read document metadata, structure, sections, summary, or full text.", objectSchema(map[string]any{
		"id":       map[string]any{"type": "string"},
		"path":     map[string]any{"type": "string"},
		"slug":     map[string]any{"type": "string"},
		"mode":     map[string]any{"type": "string"},
		"sections": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
	}), func(ctx context.Context, raw json.RawMessage) (any, error) {
		var req mcpschema.ReadDocumentRequest
		if err := decode(raw, &req); err != nil {
			return nil, err
		}
		return s.local.ReadDocument(ctx, req)
	})
}

func (s *Server) registerLocalWriteTools() {
	s.register(mcpschema.ToolCaptureNote, "Capture agent-authored markdown under agents/{actor}.", objectSchema(map[string]any{
		"actor":   map[string]any{"type": "string"},
		"title":   map[string]any{"type": "string"},
		"body":    map[string]any{"type": "string"},
		"type":    map[string]any{"type": "string"},
		"authors": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		"tags":    map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
	}), func(ctx context.Context, raw json.RawMessage) (any, error) {
		var req mcpschema.CaptureNoteRequest
		if err := decode(raw, &req); err != nil {
			return nil, err
		}
		return s.local.CaptureNote(ctx, req)
	})
	s.register(mcpschema.ToolPromoteDocument, "Promote an existing document to a type/status/destination.", objectSchema(map[string]any{
		"id":     map[string]any{"type": "string"},
		"path":   map[string]any{"type": "string"},
		"type":   map[string]any{"type": "string"},
		"status": map[string]any{"type": "string"},
	}), func(ctx context.Context, raw json.RawMessage) (any, error) {
		var req mcpschema.PromoteDocumentRequest
		if err := decode(raw, &req); err != nil {
			return nil, err
		}
		return s.local.PromoteDocument(ctx, req)
	})
	s.register(mcpschema.ToolArchiveDocument, "Archive a document without hard deletion.", objectSchema(map[string]any{
		"id":     map[string]any{"type": "string"},
		"path":   map[string]any{"type": "string"},
		"reason": map[string]any{"type": "string"},
	}), func(ctx context.Context, raw json.RawMessage) (any, error) {
		var req mcpschema.ArchiveDocumentRequest
		if err := decode(raw, &req); err != nil {
			return nil, err
		}
		return s.local.ArchiveDocument(ctx, req)
	})
}

func (s *Server) registerRemoteWriteTools() {
	s.register(mcpschema.ToolSyncPull, "Pull remote workspace changes when safe.", objectSchema(map[string]any{
		"dry_run": map[string]any{"type": "boolean"},
	}), func(ctx context.Context, raw json.RawMessage) (any, error) {
		var req mcpschema.SyncRequest
		if err := decode(raw, &req); err != nil {
			return nil, err
		}
		return s.local.SyncPull(ctx, req)
	})
	s.register(mcpschema.ToolSyncPush, "Push local workspace changes when safe.", objectSchema(map[string]any{
		"dry_run": map[string]any{"type": "boolean"},
	}), func(ctx context.Context, raw json.RawMessage) (any, error) {
		var req mcpschema.SyncRequest
		if err := decode(raw, &req); err != nil {
			return nil, err
		}
		return s.local.SyncPush(ctx, req)
	})
	s.register(mcpschema.ToolIndexRefresh, "Refresh configured index artifacts.", objectSchema(map[string]any{
		"mode":    map[string]any{"type": "string"},
		"dry_run": map[string]any{"type": "boolean"},
	}), func(ctx context.Context, raw json.RawMessage) (any, error) {
		var req mcpschema.IndexRefreshRequest
		if err := decode(raw, &req); err != nil {
			return nil, err
		}
		return s.local.IndexRefresh(ctx, req)
	})
}

func (s *Server) register(name mcpschema.ToolName, description string, inputSchema map[string]any, handler func(context.Context, json.RawMessage) (any, error)) {
	s.tools[name] = Tool{Name: name, Description: description, InputSchema: inputSchema, handler: handler}
}

func decode(raw json.RawMessage, out any) error {
	if len(raw) == 0 || string(raw) == "null" {
		raw = []byte("{}")
	}
	return json.Unmarshal(raw, out)
}

func emptySchema() map[string]any {
	return objectSchema(map[string]any{})
}

func objectSchema(properties map[string]any) map[string]any {
	return map[string]any{
		"type":                 "object",
		"properties":           properties,
		"additionalProperties": true,
	}
}

func toolResult(result any) map[string]any {
	content, err := json.Marshal(result)
	if err != nil {
		content = []byte(fmt.Sprintf(`{"ok":false,"error":{"code":"marshal_error","message":%q}}`, err.Error()))
	}
	return map[string]any{
		"content": []map[string]string{{
			"type": "text",
			"text": string(content),
		}},
		"isError": false,
	}
}
