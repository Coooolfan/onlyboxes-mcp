package buildinfo

const defaultVersion = "dev"

// Version is injected at build time via:
// go build -ldflags "-X github.com/onlyboxes/onlyboxes/worker/worker-sys/internal/buildinfo.Version=<version>"
var Version = defaultVersion
