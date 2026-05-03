package workspace

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type InitOptions struct {
	WorkspaceID string
}

type InitResult struct {
	WorkspaceID string
	Created     []string
	Existing    []string
}

var standardFolders = []string{
	"inbox",
	"agents",
	"working",
	"decisions",
	"runbooks",
	"projects",
	"services",
	"handoffs",
	"onboarding",
	"archive",
	".cairn",
	".cairn/schemas",
	".cairn/generated",
	".cairn/index",
}

func Init(root string, opts InitOptions) (InitResult, error) {
	workspaceID := strings.TrimSpace(opts.WorkspaceID)
	if workspaceID == "" {
		var err error
		workspaceID, err = newWorkspaceID()
		if err != nil {
			return InitResult{}, err
		}
	}

	result := InitResult{WorkspaceID: workspaceID}
	for _, dir := range standardFolders {
		created, err := ensureDir(root, dir)
		if err != nil {
			return InitResult{}, err
		}
		result.addPath(dir, created)
	}

	files := map[string]string{
		".cairn/config.yaml":          configYAML(workspaceID),
		".cairnignore":                cairnIgnore(),
		".cairn/schemas/core.yaml":    coreSchemaYAML(),
		".cairn/schemas/README.md":    schemasReadme(),
		"onboarding/team-context.md":  teamContext(),
		"onboarding/agent-setup.md":   agentSetup(),
		"onboarding/workspace-map.md": workspaceMap(),
		"AGENTS.md":                   agentsPointer(),
		"CLAUDE.md":                   claudePointer(),
	}
	for _, path := range sortedKeys(files) {
		created, err := ensureFile(root, path, files[path])
		if err != nil {
			return InitResult{}, err
		}
		result.addPath(path, created)
	}

	return result, nil
}

func ensureDir(root string, rel string) (bool, error) {
	path := filepath.Join(root, filepath.FromSlash(rel))
	if info, err := os.Stat(path); err == nil {
		if !info.IsDir() {
			return false, fmt.Errorf("workspace init path exists and is not a directory: %s", rel)
		}
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, err
	}
	return true, os.MkdirAll(path, 0o755)
}

func ensureFile(root string, rel string, content string) (bool, error) {
	path := filepath.Join(root, filepath.FromSlash(rel))
	if info, err := os.Stat(path); err == nil {
		if info.IsDir() {
			return false, fmt.Errorf("workspace init path exists and is a directory: %s", rel)
		}
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, err
	}
	return true, os.WriteFile(path, []byte(content), 0o644)
}

func (result *InitResult) addPath(path string, created bool) {
	if created {
		result.Created = append(result.Created, path)
		return
	}
	result.Existing = append(result.Existing, path)
}

func newWorkspaceID() (string, error) {
	var bytes [8]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return "cairn:workspace:" + hex.EncodeToString(bytes[:]), nil
}

func sortedKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func configYAML(workspaceID string) string {
	return fmt.Sprintf(`schema_version: 1
workspace_id: %s
managed_folders:
  - inbox
  - agents
  - working
  - decisions
  - runbooks
  - projects
  - services
  - handoffs
  - onboarding
  - archive
document_types:
  note: inbox
  handoff: handoffs
  investigation: working
  decision: decisions
  runbook: runbooks
  project: projects
  service: services
  onboarding: onboarding
profiles:
  local:
    enabled: true
  pod-remote:
    enabled: false
    provider: azure_blob
    container: ""
    prefix: ""
required_skills: []
`, workspaceID)
}

func cairnIgnore() string {
	return `.DS_Store
.git/
.cairn/index/
.cairn/generated/
`
}

func schemasReadme() string {
	return `# Cairn Schemas

Place custom Cairn document schemas in this directory as YAML files.
`
}

func coreSchemaYAML() string {
	return `schema_version: 1
name: core
required_fields:
  - id
  - schema_version
  - title
  - slug
  - type
  - status
  - created
  - updated
  - authors
  - actors
  - source
  - tags
`
}

func teamContext() string {
	return starterOnboardingDoc("cairn:OnboardingTeamContext", "Team Context", "team-context", `# Team Context

Capture the pod purpose, current priorities, core systems, and collaboration norms here.
`)
}

func agentSetup() string {
	return starterOnboardingDoc("cairn:OnboardingAgentSetup", "Agent Setup", "agent-setup", `# Agent Setup

Use Cairn tools to search, validate, capture, promote, sync, and index workspace context.
`)
}

func workspaceMap() string {
	return starterOnboardingDoc("cairn:OnboardingWorkspaceMap", "Workspace Map", "workspace-map", `# Workspace Map

- inbox: inbound or unreviewed content
- agents: agent-authored notes and handoffs
- working: drafts and in-progress work
- decisions: canonical ADR-style decisions
- runbooks: operational procedures
- projects: project documents
- services: service documentation
- handoffs: transition notes
- onboarding: team and agent setup
- archive: archived managed documents
`)
}

func starterOnboardingDoc(id string, title string, slug string, body string) string {
	return fmt.Sprintf(`---
id: %s
schema_version: 1
title: %s
slug: %s
type: onboarding
status: draft
created: 2026-05-03T00:00:00Z
updated: 2026-05-03T00:00:00Z
authors:
  - cairn
actors:
  - cairn
source: init
tags: []
---

%s`, id, title, slug, body)
}

func agentsPointer() string {
	return "# AGENTS\n\n" +
		"This is a Cairn workspace. Use Cairn tools for workspace bootstrap, search, validation, capture, promotion, sync, and indexing.\n\n" +
		"Start with `/onboarding/agent-setup.md` and `/onboarding/workspace-map.md`.\n"
}

func claudePointer() string {
	return "# CLAUDE\n\n" +
		"This is a Cairn workspace. Use Cairn tools for workspace bootstrap, search, validation, capture, promotion, sync, and indexing.\n\n" +
		"Start with `/onboarding/agent-setup.md` and `/onboarding/workspace-map.md`.\n"
}
