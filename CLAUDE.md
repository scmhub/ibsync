# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

`ibsync` is a Go client for the Interactive Brokers API, providing a synchronous, easy-to-use interface on top of the async [`ibapi`](https://github.com/scmhub/ibapi) package (inspired by Python's `ib_insync`/`ib_async`). It wraps `ibapi.EClient`/`EWrapper` so callers can make blocking request calls and read live, in-sync state (accounts, positions, trades, tickers, PnL, etc.) instead of handling raw async callbacks.

## Commands

```bash
go build ./...           # build
go vet ./...             # static checks
go test ./... -run '<Pattern>'   # run a specific unit test
golangci-lint run ./...  # lint (config in .golangci.yml, also run in CI)
```

Running `go test ./...` as a whole will attempt to run `ib_test.go`, which requires a **live connection to TWS/IB Gateway on a paper trading account** (`localhost:7497` by default, see `testConfig` in `ib_test.go`) with the market open. These tests will hang/fail without that live connection. Pure unit tests that don't need a connection live in `config_test.go`, `encoder_test.go`, `pubsub_test.go`, `utils_test.go`, and `trade_test.go` — prefer running these explicitly (`go test -run TestName`) unless a live TWS/paper session is confirmed available.

There is a `replace` directive commented out in `go.mod` (`// replace github.com/scmhub/ibapi => ../ibapi`) for local development against a sibling checkout of `ibapi` — leave it commented unless the user asks to develop against a local `ibapi`.

## Architecture

The package is a thin synchronous layer over the async `ibapi.EClient`/`EWrapper`. Key pieces, in `IB` (`ib.go`):

- **`ibapi.EClient`** — sends requests to TWS/IBG (unchanged from `ibapi`, reqID handling included).
- **`WrapperSync`** (`wrapper.go`) — implements `ibapi.EWrapper`. Every callback from TWS lands here. It does two things with each callback: (1) updates the shared `ibState`, and (2) publishes an encoded message onto the `PubSub` bus, keyed by reqID (or another topic key).
- **`ibState`** (`state.go`) — the single mutex-protected struct holding all synced data: accounts, positions, portfolio, trades (`permID2Trade`), fills, tickers (`reqID2Ticker`, `ticker2ReqID`), PnL subscriptions, etc. Callers must lock/unlock `state.mu` themselves when touching it directly (see the comment at the top of `state.go`).
- **`PubSub`** (`pubsub.go`) — an in-process pub/sub over `chan string`, topic = `fmt.Sprint(key)` (usually a reqID). This is how request methods turn async callbacks into blocking calls.

**Request pattern**: most `ReqXxx` methods on `IB` follow the same shape (see `ReqContractDetails` in `ib.go` for a canonical example):
1. Get a new reqID (`ib.NextID()`).
2. `ib.pubSub.Subscribe(reqID, bufSize)` to get a channel + unsubscribe func (deferred).
3. Call the matching `ib.eClient.ReqXxx(reqID, ...)`.
4. Loop on `select` between `ctx.Done()` (timeout, from `ib.config.Timeout`) and the subscribed channel, decoding messages until an `"end"` sentinel or an error message (`isErrorMsg`/`msg2Error`) arrives.

**Message encoding** (`encoder.go`): messages pushed through `PubSub` are plain `string`s, but payloads are typically gob-encoded + base64 (`Encode`/`Decode`). New struct types passed this way must be registered with `gob.Register` in `encoder.go`'s `init()`. `Key`/`Join`/`Split` build/parse `::`-separated composite keys (e.g. `Key(account, tag, currency)` for map keys in `ibState`).

**Re-exports** (`ibapi.go`): this file re-exports `ibapi` types, constants, and constructors (contracts, orders, tick types, etc.) under `ibsync` so consumers only need to import one package. When adding support for new `ibapi` fields/types, mirror them here rather than requiring users to import `ibapi` directly.

**Domain types**:
- `Trade` (`trade.go`) — a contract + order + live `OrderStatusData`, with `Fills()`/`Logs()`, and `Done()`/`Ack()` channels closed when the order reaches a terminal state / is acknowledged.
- `Ticker` (`ticker.go`) — thread-safe live market data (bid/ask/last, greeks, DOM, tick-by-tick) for a contract, updated via `WrapperSync`.
- `Contract` helpers (`contract.go`) — `NewStock`, `NewOption`, `NewFuture`, `NewForex`, etc., convenience constructors over `ibapi.Contract`.
- `Config` (`config.go`) — functional-options config (`WithHost`, `WithPort`, `WithClientID`, `WithTimeout`, `WithoutSync`, ...) for `NewIB`/`Connect`.

`examples/` contains small runnable programs (one directory per feature area: `basics`, `orders`, `bar_data`, `tick_data`, `option_chain`, `market_depth`, `pnl`, `scanners`, `contract_details`) — check the matching example when working on a feature area.

## Notes

- The package is beta; TWS/IBKR error codes are triaged in `errors.go` (`warningCodes` list) — some IB error codes are intentionally downgraded to warnings rather than treated as request failures.
- Logging goes through `zerolog` via `ibapi.Logger()`/`SetLogger`/`SetConsoleWriter` (re-exported in `ibapi.go`); `config.go` sets the package-level `log`.
