package main

import (
	"time"

	"github.com/scmhub/ibsync"
)

func main() {
	log := ibsync.Logger()
	ibsync.SetLogLevel(0)
	ibsync.SetConsoleWriter()

	// New IB client & Connect
	ib := ibsync.NewIB()

	err := ib.Connect(
		ibsync.NewConfig(
			ibsync.WithHost("127.0.0.1"),
			ibsync.WithPort(7497),
			ibsync.WithClientID(10),
		),
	)
	if err != nil {
		log.Error().Err(err).Msg("Connect")
		return
	}
	defer ib.Disconnect()

	// CPAY
	// cpay := ibsync.NewStock("CPAY", "NASDAQ", "USD")
	// order := ibsync.MarketOrder("BUY", ibsync.StringToDecimal("100"))
	// trade := ib.PlaceOrder(cpay, order)

	// Create EUR/USD forex instrument
	eurusd := ibsync.NewForex("EUR", "IDEALPRO", "USD")
	orderEurusd := ibsync.MarketOrder("BUY", ibsync.StringToDecimal("20000"))

	trade := ib.PlaceOrder(eurusd, orderEurusd)

	log.Info().Msgf("Waiting for trade to complete: %s", trade)
	<-trade.Done()
	log.Info().Msgf("Trade completed: %s", trade)

	time.Sleep(5 * time.Second) // Wait for fills to be processed
	fills := trade.Fills()
	log.Info().Msgf("Fills: %d", len(fills)) // always 0 fills in this example
	for _, fill := range fills {
		log.Info().Msgf("Fill: %v", fill)
	}
}
