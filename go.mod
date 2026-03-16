module github.com/scmhub/ibsync

go 1.26

require (
	github.com/rs/zerolog v1.34.0
	github.com/scmhub/ibapi v0.10.45-0.20260315191916-45a8382701ff
)

require (
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/robaho/fixed v0.0.0-20251201003256-beee5759f86a // indirect
	golang.org/x/sys v0.42.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

// Use local version for development
// replace github.com/scmhub/ibapi => ../ibapi
