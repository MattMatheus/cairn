package syncstate

import "sort"

type ChangeType string

const (
	ChangeCreate  ChangeType = "create"
	ChangeEdit    ChangeType = "edit"
	ChangeMove    ChangeType = "move"
	ChangeArchive ChangeType = "archive"
	ChangeDelete  ChangeType = "delete"
)

type Change struct {
	Type         ChangeType `json:"type"`
	Path         string     `json:"path"`
	PreviousPath string     `json:"previous_path,omitempty"`
	DocumentID   string     `json:"document_id,omitempty"`
}

type Conflict struct {
	Local  Change `json:"local"`
	Remote Change `json:"remote"`
}

type Comparison struct {
	LocalChanges  []Change   `json:"local_changes"`
	RemoteChanges []Change   `json:"remote_changes"`
	Conflicts     []Conflict `json:"conflicts,omitempty"`
	Diverged      bool       `json:"diverged"`
}

func Compare(base Manifest, local Manifest, remote Manifest) Comparison {
	localChanges := Changes(base, local)
	remoteChanges := Changes(base, remote)
	comparison := Comparison{
		LocalChanges:  localChanges,
		RemoteChanges: remoteChanges,
		Diverged:      len(localChanges) > 0 && len(remoteChanges) > 0,
	}
	if comparison.Diverged {
		comparison.Conflicts = conflicts(localChanges, remoteChanges)
	}
	return comparison
}

func Changes(base Manifest, current Manifest) []Change {
	normalizeEntries(base.Entries)
	normalizeEntries(current.Entries)

	baseByPath := map[string]Entry{}
	currentByPath := map[string]Entry{}
	baseByDoc := map[string]Entry{}
	currentByDoc := map[string]Entry{}

	for _, entry := range base.Entries {
		baseByPath[entry.Path] = entry
		if entry.DocumentID != "" {
			baseByDoc[entry.DocumentID] = entry
		}
	}
	for _, entry := range current.Entries {
		currentByPath[entry.Path] = entry
		if entry.DocumentID != "" {
			currentByDoc[entry.DocumentID] = entry
		}
	}

	var changes []Change
	handledCurrent := map[string]bool{}

	for _, baseEntry := range base.Entries {
		currentEntry, exists := currentByPath[baseEntry.Path]
		if exists {
			handledCurrent[currentEntry.Path] = true
			if entryEdited(baseEntry, currentEntry) {
				changes = append(changes, Change{
					Type:       ChangeEdit,
					Path:       currentEntry.Path,
					DocumentID: currentEntry.DocumentID,
				})
			}
			continue
		}

		if baseEntry.DocumentID != "" {
			if movedEntry, ok := currentByDoc[baseEntry.DocumentID]; ok {
				handledCurrent[movedEntry.Path] = true
				changeType := ChangeMove
				if movedEntry.Status == "archived" || hasArchivePrefix(movedEntry.Path) {
					changeType = ChangeArchive
				}
				changes = append(changes, Change{
					Type:         changeType,
					Path:         movedEntry.Path,
					PreviousPath: baseEntry.Path,
					DocumentID:   movedEntry.DocumentID,
				})
				if entryEdited(baseEntry, movedEntry) {
					changes = append(changes, Change{
						Type:         ChangeEdit,
						Path:         movedEntry.Path,
						PreviousPath: baseEntry.Path,
						DocumentID:   movedEntry.DocumentID,
					})
				}
				continue
			}
		}

		changes = append(changes, Change{
			Type:       ChangeDelete,
			Path:       baseEntry.Path,
			DocumentID: baseEntry.DocumentID,
		})
	}

	for _, currentEntry := range current.Entries {
		if handledCurrent[currentEntry.Path] {
			continue
		}
		if _, existed := baseByPath[currentEntry.Path]; existed {
			continue
		}
		if currentEntry.DocumentID != "" {
			if _, existed := baseByDoc[currentEntry.DocumentID]; existed {
				continue
			}
		}
		changes = append(changes, Change{
			Type:       ChangeCreate,
			Path:       currentEntry.Path,
			DocumentID: currentEntry.DocumentID,
		})
	}

	sort.Slice(changes, func(i, j int) bool {
		if changes[i].Path == changes[j].Path {
			return changes[i].Type < changes[j].Type
		}
		return changes[i].Path < changes[j].Path
	})
	return changes
}

func entryEdited(base Entry, current Entry) bool {
	return base.Hash != current.Hash || base.Size != current.Size || base.Status != current.Status || base.Type != current.Type
}

func hasArchivePrefix(path string) bool {
	return path == "archive" || len(path) > len("archive/") && path[:len("archive/")] == "archive/"
}

func conflicts(localChanges []Change, remoteChanges []Change) []Conflict {
	var result []Conflict
	for _, local := range localChanges {
		for _, remote := range remoteChanges {
			if local.DocumentID != "" && local.DocumentID == remote.DocumentID {
				result = append(result, Conflict{Local: local, Remote: remote})
				continue
			}
			if local.Path != "" && local.Path == remote.Path {
				result = append(result, Conflict{Local: local, Remote: remote})
			}
		}
	}
	if len(result) == 0 {
		result = append(result, Conflict{Local: localChanges[0], Remote: remoteChanges[0]})
	}
	return result
}
