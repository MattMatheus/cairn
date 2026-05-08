package ado

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Candidate struct {
	Title string
	Body  string
	Tags  []string
}

func BuildCandidate(event string, payload []byte) (Candidate, error) {
	if event != "pr-completed" {
		return Candidate{}, fmt.Errorf("unsupported ADO event %q", event)
	}
	var data map[string]any
	if err := json.Unmarshal(payload, &data); err != nil {
		return Candidate{}, err
	}
	resource := mapValue(data, "resource")
	repo := mapValue(resource, "repository")
	createdBy := mapValue(resource, "createdBy")
	closedBy := mapValue(resource, "closedBy")

	prTitle := firstString(resource, "title")
	if prTitle == "" {
		prTitle = firstString(data, "message", "detailedMessage")
	}
	if prTitle == "" {
		prTitle = "ADO PR Completed"
	}
	prID := firstString(resource, "pullRequestId", "pullRequestId")
	repoName := firstString(repo, "name")
	sourceRef := trimRef(firstString(resource, "sourceRefName"))
	targetRef := trimRef(firstString(resource, "targetRefName"))
	actor := firstString(closedBy, "displayName", "uniqueName")
	if actor == "" {
		actor = firstString(createdBy, "displayName", "uniqueName")
	}
	url := firstString(resource, "url", "remoteUrl")
	description := firstString(resource, "description")

	title := "ADO PR completed: " + prTitle
	var body strings.Builder
	body.WriteString("# ")
	body.WriteString(title)
	body.WriteString("\n\n")
	body.WriteString("## ADO Origin\n\n")
	writeField(&body, "Event", event)
	writeField(&body, "Pull request", prID)
	writeField(&body, "Repository", repoName)
	writeField(&body, "Source branch", sourceRef)
	writeField(&body, "Target branch", targetRef)
	writeField(&body, "Actor", actor)
	writeField(&body, "URL", url)
	body.WriteString("\n## Summary\n\n")
	if description != "" {
		body.WriteString(description)
		body.WriteString("\n")
	} else {
		body.WriteString("Candidate knowledge captured from an Azure DevOps PR completion event.\n")
	}
	body.WriteString("\n## Recommended Next Action\n\n")
	body.WriteString("Review this candidate and promote it only if it should become durable pod knowledge.\n")

	tags := []string{"ado", "candidate"}
	if repoName != "" {
		tags = append(tags, safeTag(repoName))
	}
	return Candidate{Title: title, Body: body.String(), Tags: tags}, nil
}

func mapValue(values map[string]any, key string) map[string]any {
	if values == nil {
		return nil
	}
	if nested, ok := values[key].(map[string]any); ok {
		return nested
	}
	return nil
}

func firstString(values map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := values[key]
		if !ok {
			continue
		}
		switch typed := value.(type) {
		case string:
			if strings.TrimSpace(typed) != "" {
				return strings.TrimSpace(typed)
			}
		case float64:
			return fmt.Sprintf("%.0f", typed)
		}
	}
	return ""
}

func trimRef(value string) string {
	return strings.TrimPrefix(value, "refs/heads/")
}

func writeField(builder *strings.Builder, label string, value string) {
	if value == "" {
		return
	}
	builder.WriteString("- ")
	builder.WriteString(label)
	builder.WriteString(": ")
	builder.WriteString(value)
	builder.WriteString("\n")
}

func safeTag(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	lastDash := false
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			builder.WriteRune('-')
			lastDash = true
		}
	}
	tag := strings.Trim(builder.String(), "-")
	if tag == "" {
		return "repo"
	}
	return tag
}
