package mcpops

import (
	"context"

	"cairn/internal/document"
	"cairn/internal/mcpschema"
	"cairn/internal/workspace"
)

func (l *Local) ValidateWorkspace(ctx context.Context, req mcpschema.ValidateWorkspaceRequest) (mcpschema.Envelope[mcpschema.ValidateWorkspaceData], error) {
	data, err := workspace.Validate(ctx, l.Root, workspace.ValidateOptions{
		Paths: req.Paths,
		Mode:  validationMode(req.Mode),
	})
	if err != nil {
		return mcpschema.Envelope[mcpschema.ValidateWorkspaceData]{}, err
	}
	return mcpschema.Envelope[mcpschema.ValidateWorkspaceData]{
		OK:         data.Healthy,
		Data:       data,
		Provenance: l.provenance("workspace_validation"),
		NextSteps:  validateWorkspaceNextSteps(data),
	}, nil
}

func validationMode(mode string) document.ValidationMode {
	if mode == string(document.ValidationModeDurableBoundary) {
		return document.ValidationModeDurableBoundary
	}
	return document.ValidationModeDiscovery
}

func validateWorkspaceNextSteps(data mcpschema.ValidateWorkspaceData) []mcpschema.NextStep {
	if data.Healthy && len(data.Findings) == 0 {
		return []mcpschema.NextStep{{
			Action: string(mcpschema.ToolIndexStatus),
			Label:  "Check index status",
			Reason: "Workspace validation passed.",
		}}
	}
	return []mcpschema.NextStep{{
		Action: string(mcpschema.ToolValidateWorkspace),
		Label:  "Address validation findings",
		Reason: "Fix reported document or metadata health issues, then validate again.",
	}}
}
