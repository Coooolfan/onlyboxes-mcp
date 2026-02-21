package httpapi

import (
	"strings"

	"github.com/onlyboxes/onlyboxes/console/internal/buildinfo"
)

const defaultConsoleVersion = "dev"

func consoleVersion() string {
	version := strings.TrimSpace(buildinfo.Version)
	if version == "" {
		return defaultConsoleVersion
	}
	return version
}

func consoleRepoURL() string {
	return buildinfo.RepoURL
}
