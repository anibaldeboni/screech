package version

import (
	"runtime/debug"
	"strings"
	"time"
)

var (
	// Version will be the version tag if the binary is built with "go install url/tool@version".
	// If the binary is built some other way, it will be "(devel)".
	Version = "unknown"
	// Revision is taken from the vcs.revision tag in Go 1.18+.
	Revision = "unknown"
	// LastCommit is taken from the vcs.time tag in Go 1.18+.
	LastCommit time.Time
	// DirtyBuild is taken from the vcs.modified tag in Go 1.18+.
	DirtyBuild = true
)

func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	if info.Main.Version != "" {
		Version = info.Main.Version
	}
	for _, kv := range info.Settings {
		if kv.Value == "" {
			continue
		}
		switch kv.Key {
		case "vcs.revision":
			Revision = kv.Value
		case "vcs.time":
			LastCommit, _ = time.Parse(time.RFC3339, kv.Value)
		case "vcs.modified":
			DirtyBuild = kv.Value == "true"
		}
	}
}

func Short() string {
	parts := make([]string, 0, 3)
	if Version != "unknown" && Version != "(devel)" {
		parts = append(parts, Version)
	}
	if Revision != "unknown" && Revision != "" {
		parts = append(parts, "rev")
		commit := Revision
		if len(commit) > 7 {
			commit = commit[:7]
		}
		parts = append(parts, commit)
		if DirtyBuild {
			parts = append(parts, "dirty")
		}
	}
	if len(parts) == 0 {
		return "devel"
	}
	return strings.Join(parts, "-")
}
