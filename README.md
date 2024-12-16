[![Go Report Card](https://goreportcard.com/badge/github.com/scmhub/ibsync)](https://goreportcard.com/report/github.com/scmhub/ibsync)
[![Go Reference](https://pkg.go.dev/badge/github.com/scmhub/ibsync.svg)](https://pkg.go.dev/github.com/scmhub/ibsync)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

# Interactive Brokers Synchronous Golang Client 

`ibsync` is a Go package designed to simplify interaction with the [Interactive Brokers](https://www.interactivebrokers.com/en/home.php) API. It is inspired by the great [ib_insync](https://github.com/erdewit/ib_insync) Python library and based on [ibapi](https://github.com/scmhub/ibapi). It provides a synchronous, easy-to-use interface for account management, trade execution, real-time and historical market data within the IB ecosystem.

> [!CAUTION]
> This package is in the **beta phase**. While functional, it may still have bugs or incomplete features. Please test extensively in non-production environments.

## Getting Started

### Prerequisites
- **Go** version 1.23 or higher (recommended)
- An **Interactive Brokers** account with TWS or IB Gateway installed and running

### Installation
Install the package via `go get`:

```bash
go get -u github.com/scmhub/ibsync
```

## Quick start
Here’s a basic example to connect and get the managed accounts list:
```go
package main

import "github.com/scmhub/ibsync"

func main() {
	// Get the logger (zerolog)
	log := ibapi.Logger()
	ibsync.SetConsoleWriter() // pretty logs to console, for dev and test.

	// New IB client & Connect
	ib := ibsync.NewIB()

	err := ib.Connect()
	if err != nil {
		log.Error().Err(err).Msg("Connect")
		return
	}
	defer ib.Disconnect()

	managedAccounts := ib.ManagedAccounts()
	log.Info().Strs("accounts", managedAccounts).Msg("Managed accounts list")
}
```

## Usage guide

### Configuration
Connect with a different configuration.
```go
// New IB client & Connect
ib := ibsync.NewIB()

err := ib.Connect(
    ibsync.NewConfig(
        ibsync.WithHost("10.74.0.9"),       // Default: "127.0.0.1".
        ibsync.WithPort(4002),              // Default: 7497.
        ibsync.WithClientID(5),             // Default: a random number. If set to 0 it will also retreive manual orders.
        ibsync.WithTimeout(1*time.Second),  // Default is 30 seconds.
    ),
)
if err != nil {
	log.Error().Err(err).Msg("Connect")
	return
}
defer ib.Disconnect()
```

### Account
Account value, summary, positions, trades...
```go
// Account Values
accountValues := ib.AccountValues()

// Account Summary
accountSummary := ib.AccountSummary()

// Portfolio
portfolio := ib.Portfolio()

// Positions
// Subscribe to Postion
ib.ReqPositions()
// Position Channel
posChan := ib.PositionChan()

// Trades
trades := ib.Trades()
openTrades := ib.OpenTrades()

```

### Contract details
Request contract details from symbol, exchange
```go
// NewStock("AMD", "", "")
amd := ibsync.NewStock("AMD", "", "")
cd, err := ib.ReqContractDetails(amd)
if err != nil {
	log.Error().Err(err).Msg("request contract details")
	return
}
fmt.Println("number of contract found for request NewStock(\"AMD\", \"\", \"\") :", len(cd))
```

### Pnl
Subscribe to the pnl stream
```go
// Request PnL subscription for the account.
ib.ReqPnL(account, modelCode)

// Get a PnL channel to receive updates...
pnlChan := ib.PnlChan(account, modelCode)

go func() {
	for pnl := range pnlChan {
		fmt.Println("Received PnL from channel:", pnl)
	}
}()

//... Or read the last PnL on the client state
pnl := ib.Pnl(account, modelCode)
fmt.Println("Current PnL:", pnl)
```

### Orders
Place an order and create a new trade. Modify or cancel the trade. Cancel all trades
```go
// Create the contract
eurusd := ibsync.NewForex("EUR", "IDEALPRO", "USD")

// Create the order
order := ibsync.LimitOrder("BUY", ibsync.StringToDecimal("20000"), 1.05)

// Place the order
trade := ib.PlaceOrder(eurusd, order)

go func() {
		<-trade.Done()
		fmt.Println("The trade is done!!!")
	}()

// Cancel the order
ib.CancelOrder(Order, ibsync.NewOrderCancel())

// Cancel all orders
ib.ReqGlobalCancel()
```

### Bar data
Real time and historical bar data.
```go
// Historical data
barChan, _ := ib.ReqHistoricalData(eurusd, endDateTime, duration, barSize, whatToShow, useRTH, formatDate)
var bars []ibsync.Bar
for bar := range barChan {
    bars = append(bars, bar)
}

// Historical data up to date
barChan, cancel := ib.ReqHistoricalDataUpToDate(eurusd, duration, barSize, whatToShow, useRTH, formatDate)
go func() {
    for bar := range barChan {
        bars = append(bars, bar)
    }
}()
time.Sleep(10 * time.Second)
cancel()

// Real time bars
rtBarChan, cancel := ib.ReqRealTimeBars(eurusd, 5, "MIDPOINT", useRTH)
<-rtBarChan
cancel()
```

### Tick data
Real time and historical tick data.
```go
// Snapshot - Market price
snapshot, err := ib.Snapshot(eurusd)
if err != nil {
    panic(fmt.Errorf("snapshot eurusd: %v", err))
}
fmt.Println("Snapshot market price", snapshot.MarketPrice())

// Tick by tick data
tickByTick := ib.ReqTickByTickData(eurusd, "BidAsk", 100, true)
time.Sleep(5 * time.Second)
ib.CancelTickByTickData(eurusd, "BidAsk")

// HistoricalTicks
historicalTicks, err, done := ib.ReqHistoricalTicks(aapl, startDateTime, time.Time{}, 100, true, true)
```

### Scanner
Request scanner parameters and scanner subscritpion
```go
// Scanner Parameters
xml, err := ib.ReqScannerParameters()

// Scanner subscription
scanSubscription := ibsync.NewScannerSubscription()
scanSubscription.Instrument = "STK"
scanSubscription.LocationCode = "STK.US.MAJOR"
scanSubscription.ScanCode = "TOP_PERC_GAIN"

scanData, err := ib.ReqScannerSubscription(scanSubscription)

// Scanner subcscription with filter option
opts := ibsync.ScannerSubscriptionOptions{
	FilterOptions: []ibsync.TagValue{
		{Tag: "changePercAbove", Value: "20"},
		{Tag: "priceAbove", Value: "5"},
		{Tag: "priceBelow", Value: "50"},
	},
}

filterScanData, err := ib.ReqScannerSubscription(scanSubscription, opts)
```

## Documentation
For more information on how to use this package, please refer to the [GoDoc](https://pkg.go.dev/github.com/scmhub/ibsync) documentation and check the [examples](https://github.com/scmhub/ibsync/tree/main/examples) directory. You can also have a look at the `ib_test.go` file

## Acknowledgments
- [ibapi](https://github.com/scmhub/ibapi) for core API functionality.
- [ib_insync](https://github.com/erdewit/ib_insync) for API inspiration. (ib_insync is now [ib_async](https://github.com/ib-api-reloaded/ib_async))

## Notice of Non-Affiliation and Disclaimer
> [!CAUTION]
> This project is in the **beta phase** and is still undergoing testing and development. Users are advised to thoroughly test the software in non-production environments before relying on it for live trading. Features may be incomplete, and bugs may exist. Use at your own risk.

> [!IMPORTANT]
>This project is **not affiliated** with Interactive Brokers Group, Inc. All references to Interactive Brokers, including trademarks, logos, and brand names, belong to their respective owners. The use of these names is purely for informational purposes and does not imply endorsement by Interactive Brokers.

> [!IMPORTANT]
>The authors of this package make **no guarantees** regarding the software's reliability, accuracy, or suitability for any particular purpose, including trading or financial decisions. **No liability** will be accepted for any financial losses, damages, or misinterpretations arising from the use of this software.

## License
Distributed under the MIT License. See [LICENSE](./LICENSE) for more information.

## Author
**Philippe Chavanne** - [contact](https://scm.cx/contact)