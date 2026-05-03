package document

import "testing"

const validDocument = `---
id: cairn:01JZ8K4Z8V5X9R8P9M6T2Q3N4A
schema_version: 1
title: Example Document
slug: example-document
type: note
status: working
created: 2026-05-02T12:00:00Z
updated: 2026-05-02T12:00:00Z
authors:
  - matt
actors:
  - codex
source: capture
tags: []
---

# Example
`

func TestParseMarkdownCoreFrontmatter(t *testing.T) {
	result, err := ParseMarkdown(validDocument)
	if err != nil {
		t.Fatalf("ParseMarkdown returned error: %v", err)
	}
	if !result.HasFrontmatter {
		t.Fatal("expected frontmatter")
	}
	if result.Metadata.ID != "cairn:01JZ8K4Z8V5X9R8P9M6T2Q3N4A" {
		t.Fatalf("unexpected id: %q", result.Metadata.ID)
	}
	if result.Metadata.SchemaVersion != 1 {
		t.Fatalf("unexpected schema version: %d", result.Metadata.SchemaVersion)
	}
	if result.Metadata.Title != "Example Document" {
		t.Fatalf("unexpected title: %q", result.Metadata.Title)
	}
	if result.Metadata.Slug != "example-document" {
		t.Fatalf("unexpected slug: %q", result.Metadata.Slug)
	}
	if result.Metadata.Type != "note" {
		t.Fatalf("unexpected type: %q", result.Metadata.Type)
	}
	if result.Metadata.Status != "working" {
		t.Fatalf("unexpected status: %q", result.Metadata.Status)
	}
	if len(result.Metadata.Authors) != 1 || result.Metadata.Authors[0] != "matt" {
		t.Fatalf("unexpected authors: %#v", result.Metadata.Authors)
	}
	if len(result.Metadata.Actors) != 1 || result.Metadata.Actors[0] != "codex" {
		t.Fatalf("unexpected actors: %#v", result.Metadata.Actors)
	}
	if len(result.Metadata.Tags) != 0 {
		t.Fatalf("unexpected tags: %#v", result.Metadata.Tags)
	}
}

func TestValidateValidFrontmatterHasNoFindings(t *testing.T) {
	result, err := ParseMarkdown(validDocument)
	if err != nil {
		t.Fatalf("ParseMarkdown returned error: %v", err)
	}
	validation := Validate(result, ValidationModeDurableBoundary)
	if len(validation.Findings) != 0 {
		t.Fatalf("expected no findings, got %#v", validation.Findings)
	}
}

func TestMissingFrontmatterWarnsDuringDiscovery(t *testing.T) {
	result, err := ParseMarkdown("# Plain Markdown\n")
	if err != nil {
		t.Fatalf("ParseMarkdown returned error: %v", err)
	}
	validation := Validate(result, ValidationModeDiscovery)
	if validation.Blocking() {
		t.Fatalf("expected discovery validation to be non-blocking: %#v", validation.Findings)
	}
	assertFinding(t, validation, SeverityWarning, "frontmatter")
}

func TestMissingFrontmatterBlocksDurableBoundary(t *testing.T) {
	result, err := ParseMarkdown("# Plain Markdown\n")
	if err != nil {
		t.Fatalf("ParseMarkdown returned error: %v", err)
	}
	validation := Validate(result, ValidationModeDurableBoundary)
	if !validation.Blocking() {
		t.Fatalf("expected durable-boundary validation to block: %#v", validation.Findings)
	}
	assertFinding(t, validation, SeverityError, "frontmatter")
}

func TestUnknownFrontmatterFieldWarns(t *testing.T) {
	result, err := ParseMarkdown(`---
id: cairn:01JZ8K4Z8V5X9R8P9M6T2Q3N4A
schema_version: 1
title: Example Document
slug: example-document
type: note
status: working
created: 2026-05-02T12:00:00Z
updated: 2026-05-02T12:00:00Z
authors: [matt]
actors: [codex]
source: capture
tags: []
team: platform
---
`)
	if err != nil {
		t.Fatalf("ParseMarkdown returned error: %v", err)
	}
	validation := Validate(result, ValidationModeDurableBoundary)
	if validation.Blocking() {
		t.Fatalf("expected unknown field to be non-blocking: %#v", validation.Findings)
	}
	assertFinding(t, validation, SeverityWarning, "team")
}

func TestInvalidKnownFieldsFollowValidationMode(t *testing.T) {
	result, err := ParseMarkdown(`---
id: cairn:bad/id
schema_version: 1
title: Example Document
slug: Example Document
type: unknown
status: done
created: not-a-time
updated: 2026-05-02T12:00:00Z
authors: [matt]
actors: [codex]
source: capture
tags: ["#inline"]
---
`)
	if err != nil {
		t.Fatalf("ParseMarkdown returned error: %v", err)
	}

	discovery := Validate(result, ValidationModeDiscovery)
	if discovery.Blocking() {
		t.Fatalf("expected discovery invalid fields to warn: %#v", discovery.Findings)
	}
	assertFinding(t, discovery, SeverityWarning, "id")
	assertFinding(t, discovery, SeverityWarning, "slug")
	assertFinding(t, discovery, SeverityWarning, "type")
	assertFinding(t, discovery, SeverityWarning, "status")
	assertFinding(t, discovery, SeverityWarning, "created")
	assertFinding(t, discovery, SeverityWarning, "tags")

	durable := Validate(result, ValidationModeDurableBoundary)
	if !durable.Blocking() {
		t.Fatalf("expected durable invalid fields to block: %#v", durable.Findings)
	}
	assertFinding(t, durable, SeverityError, "id")
	assertFinding(t, durable, SeverityError, "slug")
	assertFinding(t, durable, SeverityError, "type")
	assertFinding(t, durable, SeverityError, "status")
	assertFinding(t, durable, SeverityError, "created")
	assertFinding(t, durable, SeverityError, "tags")
}

func TestMissingRequiredFieldsFollowValidationMode(t *testing.T) {
	result, err := ParseMarkdown(`---
id: cairn:01JZ8K4Z8V5X9R8P9M6T2Q3N4A
title: Example Document
---
`)
	if err != nil {
		t.Fatalf("ParseMarkdown returned error: %v", err)
	}

	discovery := Validate(result, ValidationModeDiscovery)
	if discovery.Blocking() {
		t.Fatalf("expected discovery missing fields to warn: %#v", discovery.Findings)
	}
	assertFinding(t, discovery, SeverityWarning, "schema_version")
	assertFinding(t, discovery, SeverityWarning, "slug")

	durable := Validate(result, ValidationModeDurableBoundary)
	if !durable.Blocking() {
		t.Fatalf("expected durable missing fields to block: %#v", durable.Findings)
	}
	assertFinding(t, durable, SeverityError, "schema_version")
	assertFinding(t, durable, SeverityError, "slug")
}

func TestInvalidFieldTypesFollowValidationMode(t *testing.T) {
	result, err := ParseMarkdown(`---
id: cairn:01JZ8K4Z8V5X9R8P9M6T2Q3N4A
schema_version: abc
title: Example Document
slug: example-document
type: note
status: working
created: 2026-05-02T12:00:00Z
updated: 2026-05-02T12:00:00Z
authors: matt
actors: [codex]
source: capture
tags: docs
---
`)
	if err != nil {
		t.Fatalf("ParseMarkdown returned error: %v", err)
	}

	discovery := Validate(result, ValidationModeDiscovery)
	if discovery.Blocking() {
		t.Fatalf("expected discovery invalid field types to warn: %#v", discovery.Findings)
	}
	assertFinding(t, discovery, SeverityWarning, "schema_version")
	assertFinding(t, discovery, SeverityWarning, "authors")
	assertFinding(t, discovery, SeverityWarning, "tags")

	durable := Validate(result, ValidationModeDurableBoundary)
	if !durable.Blocking() {
		t.Fatalf("expected durable invalid field types to block: %#v", durable.Findings)
	}
	assertFinding(t, durable, SeverityError, "schema_version")
	assertFinding(t, durable, SeverityError, "authors")
	assertFinding(t, durable, SeverityError, "tags")
}

func assertFinding(t *testing.T, validation ValidationResult, severity Severity, field string) {
	t.Helper()
	for _, finding := range validation.Findings {
		if finding.Severity == severity && finding.Field == field {
			return
		}
	}
	t.Fatalf("missing %s finding for %q in %#v", severity, field, validation.Findings)
}
