package ado

import (
	"strings"
	"testing"
)

func TestBuildCandidateFromPRCompletedPayload(t *testing.T) {
	candidate, err := BuildCandidate("pr-completed", []byte(`{
	  "resource": {
	    "pullRequestId": 42,
	    "title": "Add checkout retry",
	    "description": "Retries transient checkout failures.",
	    "sourceRefName": "refs/heads/feature/retry",
	    "targetRefName": "refs/heads/main",
	    "url": "https://dev.azure.com/org/project/_git/payments/pullrequest/42",
	    "repository": {"name": "payments-api"},
	    "closedBy": {"displayName": "Ada Lovelace"}
	  }
	}`))
	if err != nil {
		t.Fatalf("BuildCandidate() error = %v", err)
	}
	if candidate.Title != "ADO PR completed: Add checkout retry" {
		t.Fatalf("unexpected title: %s", candidate.Title)
	}
	for _, expected := range []string{
		"Pull request: 42",
		"Repository: payments-api",
		"Source branch: feature/retry",
		"Target branch: main",
		"Actor: Ada Lovelace",
		"Retries transient checkout failures.",
		"Review this candidate and promote it only if it should become durable pod knowledge.",
	} {
		if !strings.Contains(candidate.Body, expected) {
			t.Fatalf("candidate body missing %q:\n%s", expected, candidate.Body)
		}
	}
	if strings.Join(candidate.Tags, ",") != "ado,candidate,payments-api" {
		t.Fatalf("unexpected tags: %#v", candidate.Tags)
	}
}

func TestBuildCandidateRejectsUnsupportedEvent(t *testing.T) {
	if _, err := BuildCandidate("work-item-closed", []byte(`{}`)); err == nil {
		t.Fatal("expected unsupported event error")
	}
}
