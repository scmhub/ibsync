module github.com/scmhub/ibsync

go 1.26

require (
	github.com/rs/zerolog v1.35.1
	github.com/scmhub/ibapi v0.10.48-0.20260629091025-cce2c5a92c7d
)

require (
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.22 // indirect
	github.com/robaho/fixed v0.0.0-20251201003256-beee5759f86a // indirect
	golang.org/x/sys v0.46.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

// Use local version for development
// replace github.com/scmhub/ibapi => ../ibapi
