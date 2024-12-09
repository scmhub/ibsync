package main

import (
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/scmhub/ibsync"
)

// Connection constants for the Interactive Brokers API.
const (
	// host specifies the IB API server address
	host = "10.74.0.9"
	// port specifies the IB API server port
	port = 4002
	// clientID is the unique identifier for this client connection
	clientID = 5
)

func main() {
	// We get ibsync logger
	log := ibsync.Logger()
	// Set log level to Debug
	ibsync.SetLogLevel(int(zerolog.DebugLevel))
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

	// Market depth exchange list
	mktDepthsExchanges, err := ib.ReqMktDepthExchanges()
	if err != nil {
		log.Error().Err(err).Msg("ReqMktDepthExchanges")
		return
	}
	for i, mde := range mktDepthsExchanges {
		fmt.Printf("Depth market data description %v: %v\n", i, mde)
	}

	eurusd := ibsync.NewForex("EUR", "IDEALPRO", "USD")
	ib.QualifyContract(eurusd)

	// Request Market depth
	ticker, err := ib.ReqMktDepth(eurusd, 10, false)
	if err != nil {
		log.Error().Err(err).Msg("ReqMktDepth")
		return
	}

	time.Sleep(1 * time.Second)

	// Cancel Market depth
	err = ib.CancelMktDepth(eurusd, false)
	if err != nil {
		log.Error().Err(err).Msg("CancelMktDepth")
		return
	}
	bids := ticker.DomBids()
	asks := ticker.DomAsks()
	fmt.Println("DOM")
	for i := range min(len(bids), len(asks)) {
		fmt.Printf("level %v: Bid: %v | Ask %v\n", i, bids[i], asks[i])
	}

	log.Info().Msg("Good Bye!!!")
}
