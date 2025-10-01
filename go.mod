module github.com/scmhub/ibsync

go 1.25

require (
	github.com/rs/zerolog v1.34.0
	github.com/scmhub/ibapi v0.10.41-0.20251001145137-48f7a91a8c76
)

require (
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/robaho/fixed v0.0.0-20250130054609-fd0e46fcd988 // indirect
	golang.org/x/sys v0.36.0 // indirect
	google.golang.org/protobuf v1.36.9 // indirect
)

// Use local version for development
// replace github.com/scmhub/ibapi => ../ibapi
