module github.com/scmhub/ibsync

go 1.25

require (
	github.com/rs/zerolog v1.34.0
	github.com/scmhub/ibapi v0.10.41-0.20251206071903-e17274c71b27
)

require (
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/robaho/fixed v0.0.0-20251201003256-beee5759f86a // indirect
	golang.org/x/sys v0.38.0 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

// Use local version for development
// replace github.com/scmhub/ibapi => ../ibapi
