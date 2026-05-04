package mcpschema

import "time"

type ToolName string

const (
	ToolGetBootstrap      ToolName = "get_bootstrap"
	ToolCaptureNote       ToolName = "capture_note"
	ToolPromoteDocument   ToolName = "promote_document"
	ToolArchiveDocument   ToolName = "archive_document"
	ToolReadDocument      ToolName = "read_document"
	ToolFindDocument      ToolName = "find_document"
	ToolSearchContext     ToolName = "search_context"
	ToolListDocuments     ToolName = "list_documents"
	ToolValidateWorkspace ToolName = "validate_workspace"
	ToolSyncStatus        ToolName = "sync_status"
	ToolSyncPull          ToolName = "sync_pull"
	ToolSyncPush          ToolName = "sync_push"
	ToolIndexStatus       ToolName = "index_status"
	ToolIndexRefresh      ToolName = "index_refresh"
)

type Mutability string

const (
	MutabilityRead             Mutability = "read"
	MutabilityLocalWrite       Mutability = "local_write"
	MutabilityRemoteWrite      Mutability = "remote_write"
	MutabilityRemoteLocalWrite Mutability = "remote_write_local_write"
)

type Profile string

const (
	ProfileLocal     Profile = "local"
	ProfilePodRemote Profile = "pod-remote"
)

type ReadMode string

const (
	ReadModeSummary     ReadMode = "summary"
	ReadModeFrontmatter ReadMode = "frontmatter"
	ReadModeTOC         ReadMode = "toc"
	ReadModeSections    ReadMode = "sections"
	ReadModeFull        ReadMode = "full"
)

type SearchMode string

const (
	SearchModeAuto     SearchMode = "auto"
	SearchModeMetadata SearchMode = "metadata"
	SearchModeFullText SearchMode = "full_text"
	SearchModeSemantic SearchMode = "semantic"
)

type WarningCode string

const (
	WarningValidation      WarningCode = "validation"
	WarningSyncDivergence  WarningCode = "sync_divergence"
	WarningIndexDegraded   WarningCode = "index_degraded"
	WarningProfileMissing  WarningCode = "profile_missing"
	WarningRemoteAuth      WarningCode = "remote_auth_unavailable"
	WarningRemoteService   WarningCode = "remote_service_unavailable"
	WarningProgressiveRead WarningCode = "progressive_read_recommended"
)

type ToolDefinition struct {
	Name       ToolName   `json:"name"`
	Mutability Mutability `json:"mutability"`
	Purpose    string     `json:"purpose"`
}

type SchemaDefinition struct {
	Tool         ToolName `json:"tool"`
	RequestType  string   `json:"request_type"`
	ResponseType string   `json:"response_type"`
}

func V1Tools() []ToolDefinition {
	return []ToolDefinition{
		{Name: ToolGetBootstrap, Mutability: MutabilityRead, Purpose: "Return compact workspace onboarding context and next steps."},
		{Name: ToolCaptureNote, Mutability: MutabilityLocalWrite, Purpose: "Capture agent-authored markdown under agents/{actor}."},
		{Name: ToolPromoteDocument, Mutability: MutabilityLocalWrite, Purpose: "Promote an existing document to a type/status/destination."},
		{Name: ToolArchiveDocument, Mutability: MutabilityLocalWrite, Purpose: "Archive a document without hard deletion."},
		{Name: ToolReadDocument, Mutability: MutabilityRead, Purpose: "Read document metadata, structure, sections, summary, or full text."},
		{Name: ToolFindDocument, Mutability: MutabilityRead, Purpose: "Find documents by id, slug, title, path, type, status, or tag."},
		{Name: ToolSearchContext, Mutability: MutabilityRead, Purpose: "Search local and optional derived context."},
		{Name: ToolListDocuments, Mutability: MutabilityRead, Purpose: "List managed documents by filters."},
		{Name: ToolValidateWorkspace, Mutability: MutabilityRead, Purpose: "Validate managed markdown and sync/index metadata health."},
		{Name: ToolSyncStatus, Mutability: MutabilityRead, Purpose: "Report local/remote sync state."},
		{Name: ToolSyncPull, Mutability: MutabilityRemoteLocalWrite, Purpose: "Pull remote workspace changes when safe."},
		{Name: ToolSyncPush, Mutability: MutabilityRemoteWrite, Purpose: "Push local workspace changes when safe."},
		{Name: ToolIndexStatus, Mutability: MutabilityRead, Purpose: "Report local/remote index availability and freshness."},
		{Name: ToolIndexRefresh, Mutability: MutabilityRemoteLocalWrite, Purpose: "Refresh configured index artifacts."},
	}
}

func V1SchemaDefinitions() []SchemaDefinition {
	return []SchemaDefinition{
		{Tool: ToolGetBootstrap, RequestType: "EmptyRequest", ResponseType: "Envelope[GetBootstrapData]"},
		{Tool: ToolCaptureNote, RequestType: "CaptureNoteRequest", ResponseType: "Envelope[MutationResult]"},
		{Tool: ToolPromoteDocument, RequestType: "PromoteDocumentRequest", ResponseType: "Envelope[MutationResult]"},
		{Tool: ToolArchiveDocument, RequestType: "ArchiveDocumentRequest", ResponseType: "Envelope[MutationResult]"},
		{Tool: ToolReadDocument, RequestType: "ReadDocumentRequest", ResponseType: "Envelope[ReadDocumentData]"},
		{Tool: ToolFindDocument, RequestType: "FindDocumentRequest", ResponseType: "Envelope[DocumentListData]"},
		{Tool: ToolSearchContext, RequestType: "SearchContextRequest", ResponseType: "Envelope[SearchContextData]"},
		{Tool: ToolListDocuments, RequestType: "ListDocumentsRequest", ResponseType: "Envelope[DocumentListData]"},
		{Tool: ToolValidateWorkspace, RequestType: "ValidateWorkspaceRequest", ResponseType: "Envelope[ValidateWorkspaceData]"},
		{Tool: ToolSyncStatus, RequestType: "EmptyRequest", ResponseType: "Envelope[SyncStatusData]"},
		{Tool: ToolSyncPull, RequestType: "SyncRequest", ResponseType: "Envelope[SyncMutationData]"},
		{Tool: ToolSyncPush, RequestType: "SyncRequest", ResponseType: "Envelope[SyncMutationData]"},
		{Tool: ToolIndexStatus, RequestType: "IndexStatusRequest", ResponseType: "Envelope[IndexStatusData]"},
		{Tool: ToolIndexRefresh, RequestType: "IndexRefreshRequest", ResponseType: "Envelope[IndexRefreshData]"},
	}
}

type Envelope[T any] struct {
	OK          bool               `json:"ok"`
	Data        T                  `json:"data,omitempty"`
	Warnings    []Warning          `json:"warnings,omitempty"`
	Unavailable []UnavailableMode  `json:"unavailable,omitempty"`
	NextSteps   []NextStep         `json:"next_steps,omitempty"`
	Provenance  ResponseProvenance `json:"provenance,omitempty"`
	Error       *Error             `json:"error,omitempty"`
}

type Warning struct {
	Code       WarningCode       `json:"code"`
	Message    string            `json:"message"`
	Path       string            `json:"path,omitempty"`
	DocumentID string            `json:"document_id,omitempty"`
	Details    map[string]string `json:"details,omitempty"`
}

type UnavailableMode struct {
	Mode      string      `json:"mode"`
	Reason    WarningCode `json:"reason"`
	Message   string      `json:"message"`
	Retryable bool        `json:"retryable"`
}

type NextStep struct {
	Action string `json:"action"`
	Label  string `json:"label"`
	Reason string `json:"reason,omitempty"`
}

type ResponseProvenance struct {
	Profile        Profile   `json:"profile,omitempty"`
	WorkspaceID    string    `json:"workspace_id,omitempty"`
	GeneratedAt    time.Time `json:"generated_at,omitempty"`
	Source         string    `json:"source,omitempty"`
	AttemptedModes []string  `json:"attempted_modes,omitempty"`
}

type Error struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

type ActorContext struct {
	Actor   string  `json:"actor,omitempty"`
	Profile Profile `json:"profile,omitempty"`
}

type DocumentRef struct {
	ID   string `json:"id,omitempty"`
	Path string `json:"path,omitempty"`
	Slug string `json:"slug,omitempty"`
}

type DocumentSummary struct {
	ID      string    `json:"id,omitempty"`
	Path    string    `json:"path"`
	Title   string    `json:"title,omitempty"`
	Slug    string    `json:"slug,omitempty"`
	Type    string    `json:"type,omitempty"`
	Status  string    `json:"status,omitempty"`
	Tags    []string  `json:"tags,omitempty"`
	Updated time.Time `json:"updated,omitempty"`
	Authors []string  `json:"authors,omitempty"`
	Actors  []string  `json:"actors,omitempty"`
	Source  string    `json:"source,omitempty"`
}

type ChangedPath struct {
	Path         string `json:"path"`
	PreviousPath string `json:"previous_path,omitempty"`
	Kind         string `json:"kind,omitempty"`
}

type MutationResult struct {
	DocumentID   string        `json:"document_id,omitempty"`
	ChangedPaths []ChangedPath `json:"changed_paths"`
}

type EmptyRequest struct {
	ActorContext
}

type GetBootstrapData struct {
	WorkspaceID string   `json:"workspace_id,omitempty"`
	Summary     string   `json:"summary,omitempty"`
	NextSteps   []string `json:"next_steps,omitempty"`
}

type CaptureNoteRequest struct {
	ActorContext
	Title   string   `json:"title"`
	Body    string   `json:"body"`
	Type    string   `json:"type,omitempty"`
	Authors []string `json:"authors,omitempty"`
	Tags    []string `json:"tags,omitempty"`
}

type PromoteDocumentRequest struct {
	ActorContext
	DocumentRef
	Type   string `json:"type,omitempty"`
	Status string `json:"status,omitempty"`
}

type ArchiveDocumentRequest struct {
	ActorContext
	DocumentRef
	Reason string `json:"reason,omitempty"`
}

type ReadDocumentRequest struct {
	ActorContext
	DocumentRef
	Mode     ReadMode `json:"mode"`
	Sections []string `json:"sections,omitempty"`
}

type ReadDocumentData struct {
	Document    DocumentSummary   `json:"document"`
	Mode        ReadMode          `json:"mode"`
	Summary     string            `json:"summary,omitempty"`
	Frontmatter map[string]any    `json:"frontmatter,omitempty"`
	TOC         []Heading         `json:"toc,omitempty"`
	Sections    []DocumentSection `json:"sections,omitempty"`
	Content     string            `json:"content,omitempty"`
}

type Heading struct {
	Level int    `json:"level"`
	Text  string `json:"text"`
	ID    string `json:"id,omitempty"`
}

type DocumentSection struct {
	Heading string `json:"heading"`
	Content string `json:"content"`
}

type FindDocumentRequest struct {
	ActorContext
	ID     string `json:"id,omitempty"`
	Slug   string `json:"slug,omitempty"`
	Title  string `json:"title,omitempty"`
	Path   string `json:"path,omitempty"`
	Type   string `json:"type,omitempty"`
	Status string `json:"status,omitempty"`
	Tag    string `json:"tag,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

type ListDocumentsRequest struct {
	ActorContext
	Type        string   `json:"type,omitempty"`
	Status      string   `json:"status,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	ActorFilter string   `json:"actor,omitempty"`
	Source      string   `json:"source,omitempty"`
	Limit       int      `json:"limit,omitempty"`
	PageToken   string   `json:"page_token,omitempty"`
}

type DocumentListData struct {
	Documents     []DocumentSummary `json:"documents"`
	NextPageToken string            `json:"next_page_token,omitempty"`
}

type SearchContextRequest struct {
	ActorContext
	Query string     `json:"query"`
	Mode  SearchMode `json:"mode,omitempty"`
	Types []string   `json:"types,omitempty"`
	Tags  []string   `json:"tags,omitempty"`
	Limit int        `json:"limit,omitempty"`
}

type SearchContextData struct {
	Results        []SearchResult `json:"results"`
	AttemptedModes []SearchMode   `json:"attempted_modes"`
}

type SearchResult struct {
	Path       string         `json:"path"`
	Title      string         `json:"title,omitempty"`
	Type       string         `json:"type,omitempty"`
	Status     string         `json:"status,omitempty"`
	Slug       string         `json:"slug,omitempty"`
	Tags       []string       `json:"tags,omitempty"`
	Updated    time.Time      `json:"updated,omitempty"`
	Score      float64        `json:"score,omitempty"`
	MatchType  SearchMode     `json:"match_type"`
	Snippet    string         `json:"snippet,omitempty"`
	Provenance ItemProvenance `json:"provenance,omitempty"`
}

type ItemProvenance struct {
	Authors []string `json:"authors,omitempty"`
	Actors  []string `json:"actors,omitempty"`
	Source  string   `json:"source,omitempty"`
}

type ValidateWorkspaceRequest struct {
	ActorContext
	Paths []string `json:"paths,omitempty"`
	Mode  string   `json:"mode,omitempty"`
}

type ValidateWorkspaceData struct {
	Findings []ValidationFinding `json:"findings"`
	Healthy  bool                `json:"healthy"`
}

type ValidationFinding struct {
	Severity   string      `json:"severity"`
	Code       WarningCode `json:"code"`
	Message    string      `json:"message"`
	Path       string      `json:"path,omitempty"`
	DocumentID string      `json:"document_id,omitempty"`
}

type SyncRequest struct {
	ActorContext
	DryRun bool `json:"dry_run,omitempty"`
}

type SyncDirection string

const (
	SyncDirectionClean   SyncDirection = "clean"
	SyncDirectionPull    SyncDirection = "pull"
	SyncDirectionPush    SyncDirection = "push"
	SyncDirectionRefused SyncDirection = "refused"
)

type SyncStatusData struct {
	LocalChanges  []SyncChange   `json:"local_changes,omitempty"`
	RemoteChanges []SyncChange   `json:"remote_changes,omitempty"`
	Conflicts     []SyncConflict `json:"conflicts,omitempty"`
	Diverged      bool           `json:"diverged"`
	LastSyncAt    time.Time      `json:"last_sync_at,omitempty"`
	BaseHash      string         `json:"base_hash,omitempty"`
	RemoteHash    string         `json:"remote_hash,omitempty"`
}

type SyncChange struct {
	Type         string `json:"type"`
	Path         string `json:"path"`
	PreviousPath string `json:"previous_path,omitempty"`
	DocumentID   string `json:"document_id,omitempty"`
}

type SyncConflict struct {
	Local  SyncChange `json:"local"`
	Remote SyncChange `json:"remote"`
}

type SyncMutationData struct {
	MutationResult
	Applied  bool          `json:"applied"`
	Diverged bool          `json:"diverged"`
	Plan     *SyncPlanData `json:"plan,omitempty"`
}

type SyncPlanData struct {
	Direction      SyncDirection  `json:"direction"`
	Safe           bool           `json:"safe"`
	PlannedChanges []SyncChange   `json:"planned_changes,omitempty"`
	Conflicts      []SyncConflict `json:"conflicts,omitempty"`
	Diverged       bool           `json:"diverged"`
}

type IndexStatusRequest struct {
	ActorContext
}

type IndexStatusData struct {
	LocalAvailable  bool      `json:"local_available"`
	RemoteAvailable bool      `json:"remote_available"`
	Fresh           bool      `json:"fresh"`
	LastRefreshAt   time.Time `json:"last_refresh_at,omitempty"`
}

type IndexRefreshRequest struct {
	ActorContext
	Mode   SearchMode `json:"mode,omitempty"`
	DryRun bool       `json:"dry_run,omitempty"`
}

type IndexRefreshData struct {
	MutationResult
	LocalRefreshed  bool `json:"local_refreshed"`
	RemoteRefreshed bool `json:"remote_refreshed"`
}
