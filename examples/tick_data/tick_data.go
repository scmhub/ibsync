package main

import (
	"fmt"
	"time"

	"github.com/scmhub/ibsync"
)

// Connection constants for the Interactive Brokers API.
const (
	// host specifies the IB API server address
	host = "localhost"
	// port specifies the IB API server port
	port = 7497
	// clientID is the unique identifier for this client connection
	clientID = 5
)

func main() {
	// We get ibsync logger
	log := ibsync.Logger()
	// Set log level to Debug
	// ibsync.SetLogLevel(int(zerolog.DebugLevel))
	// Set logger for pretty logs to console
	ibsync.SetConsoleWriter()

	// New IB client & Connect
	ib := ibsync.NewIB()

	// Connect ib with config
	err := ib.Connect(
		ibsync.NewConfig(
			ibsync.WithHost(host),
			ibsync.WithPort(port),
			ibsync.WithClientID(clientID),
		),
	)
	if err != nil {
		log.Error().Err(err).Msg("Connect")
		return
	}
	defer ib.Disconnect()

	eurusd := ibsync.NewForex("EUR", "IDEALPRO", "USD")

	err = ib.QualifyContract(eurusd)
	if err != nil {
		panic(fmt.Errorf("qualify eurusd: %v", err))
	}

	// Snapshot - Market price
	snapshot, err := ib.Snapshot(eurusd)
	if err != nil {
		panic(fmt.Errorf("snapshot eurusd: %v", err))
	}
	fmt.Println("Snapshot", snapshot)
	fmt.Println("Snapshot market price", snapshot.MarketPrice())

	// Streaming Tick data
	eurusdTicker := ib.ReqMktData(eurusd, "")
	time.Sleep(5 * time.Second)
	ib.CancelMktData(eurusd)

	fmt.Println("Streaming Tick data:", eurusdTicker)

	// Tick by tick data
	tickByTick := ib.ReqTickByTickData(eurusd, "BidAsk", 100, true)
	time.Sleep(5 * time.Second)
	ib.CancelTickByTickData(eurusd, "BidAsk")

	fmt.Println("Tick By Tick dat:", tickByTick)

	// Historical ticks

	aapl := ibsync.NewStock("AAPL", "SMART", "USD")

	// HistoricalTicks
	historicalTicks, err, done := ib.ReqHistoricalTicks(aapl, ibsync.LastWednesday12EST(), time.Time{}, 100, true, true)

	if err != nil {
		log.Error().Err(err).Msg("ReqHistoricalTicks")
		return
	}

	fmt.Printf("Historical Ticks number %v, is done? %v\n", len(historicalTicks), done)
	for i, ht := range historicalTicks {
		fmt.Printf("%v: %v\n", i, ht)
	}

	// HistoricalTickLast
	historicalTickLast, err, done := ib.ReqHistoricalTickLast(aapl, ibsync.LastWednesday12EST(), time.Time{}, 100, true, true)

	if err != nil {
		log.Error().Err(err).Msg("ReqHistoricalTicks")
		return
	}

	fmt.Printf("Historical Last Ticks number %v, is done? %v\n", len(historicalTickLast), done)
	for i, htl := range historicalTickLast {
		fmt.Printf("%v: %v\n", i, htl)
	}

	// HistoricalTickBidAsk
	historicalTickBidAsk, err, done := ib.ReqHistoricalTickBidAsk(aapl, ibsync.LastWednesday12EST(), time.Time{}, 100, true, true)

	if err != nil {
		log.Error().Err(err).Msg("ReqHistoricalTicks")
		return
	}

	fmt.Printf("Historical Bid Ask Ticks number %v, is done? %v\n", len(historicalTickBidAsk), done)
	for i, htba := range historicalTickBidAsk {
		fmt.Printf("%v: %v\n", i, htba)
	}

	log.Info().Msg("Good Bye!!!")
}
