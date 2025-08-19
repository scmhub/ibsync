module github.com/scmhub/ibsync

go 1.23.0

require (
	github.com/rs/zerolog v1.34.0
	github.com/scmhub/ibapi v0.10.40-0.20250819062216-e7b2cb15a941
)

require (
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/robaho/fixed v0.0.0-20250130054609-fd0e46fcd988 // indirect
	golang.org/x/sys v0.35.0 // indirect
	google.golang.org/protobuf v1.36.7 // indirect
)

// Use local version for development
// replace github.com/scmhub/ibapi => ../ibapi
