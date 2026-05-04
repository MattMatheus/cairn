package syncstate

import (
	"cairn/internal/mcpschema"
)

type PlanDirection string

const (
	PlanDirectionClean   PlanDirection = "clean"
	PlanDirectionPull    PlanDirection = "pull"
	PlanDirectionPush    PlanDirection = "push"
	PlanDirectionRefused PlanDirection = "refused"
)

type Plan struct {
	Direction PlanDirection
	Safe      bool
	Changes   []Change
	Conflicts []Conflict
	Diverged  bool
}

func PlanFromStatus(status Status) Plan {
	comparison := status.Comparison
	if comparison.Diverged {
		return Plan{
			Direction: PlanDirectionRefused,
			Safe:      false,
			Conflicts: comparison.Conflicts,
			Diverged:  true,
		}
	}
	if len(comparison.LocalChanges) > 0 {
		return Plan{
			Direction: PlanDirectionPush,
			Safe:      true,
			Changes:   comparison.LocalChanges,
		}
	}
	if len(comparison.RemoteChanges) > 0 {
		return Plan{
			Direction: PlanDirectionPull,
			Safe:      true,
			Changes:   comparison.RemoteChanges,
		}
	}
	return Plan{Direction: PlanDirectionClean, Safe: true}
}

func (p Plan) Schema() mcpschema.SyncPlanData {
	return mcpschema.SyncPlanData{
		Direction:      mcpschema.SyncDirection(p.Direction),
		Safe:           p.Safe,
		PlannedChanges: schemaChanges(p.Changes),
		Conflicts:      schemaConflicts(p.Conflicts),
		Diverged:       p.Diverged,
	}
}

func (p Plan) Warnings() []mcpschema.Warning {
	if p.Direction != PlanDirectionRefused {
		return nil
	}
	return []mcpschema.Warning{{
		Code:    mcpschema.WarningSyncDivergence,
		Message: "sync dry-run refused because local and remote changes diverged",
	}}
}

func (p Plan) NextSteps() []mcpschema.NextStep {
	switch p.Direction {
	case PlanDirectionClean:
		return []mcpschema.NextStep{{
			Action: string(mcpschema.ToolValidateWorkspace),
			Label:  "Validate workspace",
			Reason: "No sync changes were detected.",
		}}
	case PlanDirectionPull:
		return []mcpschema.NextStep{{
			Action: string(mcpschema.ToolSyncPull),
			Label:  "Pull remote changes when ready",
			Reason: "Dry-run found only remote changes.",
		}}
	case PlanDirectionPush:
		return []mcpschema.NextStep{{
			Action: string(mcpschema.ToolSyncPush),
			Label:  "Push local changes when ready",
			Reason: "Dry-run found only local changes.",
		}}
	default:
		return []mcpschema.NextStep{
			{Action: "review_conflicts", Label: "Review conflicting paths", Reason: "Manual reconciliation is required before pull or push."},
			{Action: string(mcpschema.ToolSyncStatus), Label: "Run sync status again", Reason: "Retry after reconciling local or remote changes."},
		}
	}
}

func PlanEnvelope(status Status) mcpschema.Envelope[mcpschema.SyncMutationData] {
	plan := PlanFromStatus(status)
	envelope := mcpschema.Envelope[mcpschema.SyncMutationData]{
		OK: plan.Safe,
		Data: mcpschema.SyncMutationData{
			Diverged: plan.Diverged,
			Plan:     ptr(plan.Schema()),
		},
		Warnings:  plan.Warnings(),
		NextSteps: plan.NextSteps(),
	}
	if !status.RemoteAvailable {
		envelope.Warnings = append(envelope.Warnings, mcpschema.Warning{
			Code:    mcpschema.WarningRemoteService,
			Message: "remote manifest is unavailable in local profile",
			Path:    ".cairn/remote-manifest.json",
		})
		envelope.Unavailable = append(envelope.Unavailable, mcpschema.UnavailableMode{
			Mode:      "remote_manifest",
			Reason:    mcpschema.WarningRemoteService,
			Message:   "remote manifest fixture is not available",
			Retryable: false,
		})
	}
	return envelope
}

func ptr[T any](value T) *T {
	return &value
}
