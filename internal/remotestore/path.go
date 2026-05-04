package remotestore

import (
	"path"
	"path/filepath"
	"strings"
)

func CleanPath(value string) string {
	clean := path.Clean(strings.TrimPrefix(filepath.ToSlash(value), "/"))
	if clean == "." {
		return ""
	}
	return clean
}

func JoinPrefix(prefix string, workspacePath string) string {
	prefix = CleanPath(prefix)
	workspacePath = CleanPath(workspacePath)
	switch {
	case prefix == "":
		return workspacePath
	case workspacePath == "":
		return prefix
	default:
		return prefix + "/" + workspacePath
	}
}

func StripPrefix(prefix string, objectName string) string {
	prefix = CleanPath(prefix)
	objectName = CleanPath(objectName)
	if prefix == "" {
		return objectName
	}
	if objectName == prefix {
		return ""
	}
	return strings.TrimPrefix(objectName, prefix+"/")
}
