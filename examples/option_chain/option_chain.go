package main

import (
	"fmt"
	"math"
	"time"

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
	// We set logger for pretty logs to console
	log := ibsync.Logger()
	// ibsync.SetLogLevel(int(zerolog.DebugLevel))
	ibsync.SetConsoleWriter()

	// New IB client & Connect
	ib := ibsync.NewIB()

	err := ib.Connect(
		ibsync.NewConfig(
			ibsync.WithHost(host),
			ibsync.WithPort(port),
			ibsync.WithClientID(clientID),
		),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Connect")
	}
	defer ib.Disconnect()

	// Requests delayed "frozen" data for a user without market data subscriptions.
	ib.ReqMarketDataType(4)

	// S&P index
	spx := ibsync.NewIndex("SPX", "CBOE", "USD")

	err = ib.QualifyContract(spx)
	if err != nil {
		panic(fmt.Errorf("qualify spx: %v", err))
	}

	spxTicker, err := ib.Snapshot(spx)
	fmt.Println(spxTicker)
	if err != nil && err != ibsync.WarnDelayedMarketData {
		panic(fmt.Errorf("spx snapshot: %v", err))
	}
	spxPrice := spxTicker.MarketPrice()
	fmt.Println("SPX market price:", spxPrice)

	// Options chain
	chains, err := ib.ReqSecDefOptParams(spx.Symbol, "", spx.SecType, spx.ConID)
	if err != nil {
		panic(fmt.Errorf("security defimition option parameters: %v", err))
	}

	for i, oc := range chains {
		fmt.Printf("Option chain %v: %v\n", i, oc)
	}

	maturity := time.Now().AddDate(0, 3, 0).Format("200601") // three month from now
	strike := math.Round(spxPrice/250) * 250
	call := ibsync.NewOption("SPX", maturity, strike, "C", "SMART", "100", "USD")
	call.TradingClass = "SPX"

	err = ib.QualifyContract(call)
	if err != nil {
		panic(fmt.Errorf("qualify option: %v", err))
	}

	// Get option market price (if available) or model price.
	callTicker, err := ib.Snapshot(call)
	if err != nil && err != ibsync.WarnCompetingLiveSession && err != ibsync.WarnDelayedMarketData {
		panic(fmt.Errorf("call snapshot: %v", err))
	}

	greeks := callTicker.Greeks()
	fmt.Println(callTicker)
	fmt.Printf("Option price: %.2f, implied volatility: %.2f%%, vega:%.2f\n", greeks.OptPrice, greeks.ImpliedVol*100, greeks.Vega)

	// Option price for a given implied volatility & underlying price
	optionPrice, err := ib.CalculateOptionPrice(call, greeks.ImpliedVol+0.01, greeks.UndPrice)
	if err != nil {
		panic(fmt.Errorf("calculate option price: %v", err))
	}
	fmt.Printf("Option price: %.2f, was expecting: %.2f\n", optionPrice.OptPrice, greeks.OptPrice+greeks.Vega)

	// Implied Volatility for a given option price & underlying price
	impliedVol, err := ib.CalculateImpliedVolatility(call, greeks.OptPrice+greeks.Vega, greeks.UndPrice)
	if err != nil {
		panic(fmt.Errorf("calculate implied volatility: %v", err))
	}
	fmt.Printf("Implied Volatility: %.2f%%, was expecting: %.2f%%\n", impliedVol.ImpliedVol*100, greeks.ImpliedVol*100+1)

	log.Info().Msg("Good Bye!!!")
}
