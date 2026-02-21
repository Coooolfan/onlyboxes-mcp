package buildinfo

const (
	defaultVersion = "dev"
	RepoURL        = "https://github.com/Coooolfan/onlyboxes"
)

// Version is injected at build time via:
// go build -ldflags "-X github.com/onlyboxes/onlyboxes/console/internal/buildinfo.Version=<version>"
var Version = defaultVersion
