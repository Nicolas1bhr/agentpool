// Package version reports the build identity of the agentpool binary.
//
// Version and Commit are injected at link time by the Makefile and the release workflow:
//
//	-ldflags "-X github.com/Nicolas1bhr/agentpool/internal/version.Version=0.1.0"
package version

import "fmt"

var (
	// Version is the semantic version without a leading "v", or "dev" for local builds.
	Version = "dev"
	// Commit is the abbreviated git commit the binary was built from.
	Commit = "unknown"
)

// String returns the human-readable build identity.
func String() string {
	return fmt.Sprintf("agentpool %s (%s)", Version, Commit)
}
