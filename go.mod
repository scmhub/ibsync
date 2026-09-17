module github.com/scmhub/ibsync

go 1.26.0

require (
	github.com/rs/zerolog v1.35.1
	github.com/scmhub/ibapi v0.10.51-0.20260917090330-c78d3df66579
)

require (
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/robaho/fixed v0.0.0-20251201003256-beee5759f86a // indirect
	golang.org/x/sys v0.48.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
)

// Use local version for development
// replace github.com/scmhub/ibapi => ../ibapi
