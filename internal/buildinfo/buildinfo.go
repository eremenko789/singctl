// Package buildinfo holds compile-time identity of the singctl binary.
package buildinfo

import "fmt"

// VersionIdentity fields — overridable via -ldflags "-X ...".
var (
	Name    = "singctl"
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

// Format returns the canonical version payload for stdout
// (identical for `version` and `--version`).
func Format() string {
	return fmt.Sprintf("%s version %s\ncommit: %s\ndate: %s\n", Name, Version, Commit, Date)
}
