package version

// Version is the release identifier shown by `tars --version`.
// Release builds set this via -ldflags, for example:
//
//	go build -ldflags="-X tars/internal/version.Version=v1.0.0" ./cmd/tars
var Version = "dev"
