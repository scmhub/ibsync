// This package demonstrates how to interact with ibsunc f
// for forex trading operations including placing, modifying, and canceling orders.
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
	host = "localhost"
	// port specifies the IB API server port
	port = 7497
	// clientID is the unique identifier for this client connection
	clientID = 5
)

// printTrades prints the current state of trades.
// It displays both all trades and open trades separately.
//
// Parameters:
//   - header: descriptive text to identify the trading state
//   - count: sequential number for tracking multiple trade printouts
//   - printPrevious: when true, prints all trades including those with OrderID=0
//   - ib: pointer to the IB connection instance
func printTrades(header string, count int, printPrevious bool, ib *ibsync.IB) {
	fmt.Println("***   ***   ***")
	// Header
	fmt.Printf("*%v* *** %v ***\n", count, header)
	// Print Trades
	trades := ib.Trades()
	fmt.Printf("*%v* # trades: %v\n", count, len(trades))
	for i, t := range trades {
		if printPrevious || t.Order.OrderID != 0 {
			fmt.Printf("*%v-%v* %v\n", count, i, t)
		}
	}
	fmt.Println("***")
	// Print Open Trades
	openTrades := ib.OpenTrades()
	fmt.Printf("*%v* # open trades: %v\n", count, len(openTrades))
	for i, t := range openTrades {
		if printPrevious || t.Order.OrderID != 0 {
			fmt.Printf("*%v-%v* %v\n", count, i, t)
		}
	}
	fmt.Println("***   ***   ***")
}

// main demonstrates a complete workflow for forex trading using the IB API.
// It includes:
//   - Establishing connection to IB
//   - Verifying paper trading account
//   - Getting forex midpoint prices
//   - Placing buy and sell orders
//   - Modifying existing orders
//   - Canceling specific orders
//   - Performing global cancellation
//
// The example uses EUR/USD currency pair and includes error handling
// and proper cleanup with deferred disconnect.
func main() {
	// We set logger for pretty logs to console
	log := ibsync.Logger()
	ibsync.SetLogLevel(0)
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
		log.Error().Err(err).Msg("Connect")
		return
	}
	defer ib.Disconnect()

	// Make sure that the account is a paper account.
	if !ib.IsPaperAccount() {
		log.Warn().Msg("This is not a paper trading account! Exiting!")
		log.Warn().Msg("Good Bye!!!")
		return
	}

	// Create EUR/USD forex instrument
	eurusd := ibsync.NewForex("EUR", "IDEALPRO", "USD")

	// Get current market midpoint
	eurusdMidpoint, err := ib.MidPoint(eurusd)
	if err != nil {
		log.Error().Err(err).Msg("Midpoint")
		return
	}
	fmt.Println("EURUSD midpoint: ", eurusdMidpoint)

	// Calculate order price at 95% of midpoint and rounding to three decimals.
	// sell orders will be filled and buy orders will be submitted
	orderprice := math.Round(95*eurusdMidpoint.MidPoint) / 100

	// Print Trades on server before placing orders
	printTrades("Trades from server before placing orders", 1, true, ib)

	// Place Orders

	// Place sell order
	fmt.Println("*** Place Sell order ***")
	sellOrder := ibsync.LimitOrder("SELL", ibsync.StringToDecimal("20000"), orderprice)
	sellTrade := ib.PlaceOrder(eurusd, sellOrder)

	fmt.Println("submitted sell trade:", sellTrade)

	<-sellTrade.Done()
	fmt.Println("The sell trade is done!!!")

	// Place buy orders
	fmt.Println("*** Place Buy orders ***")
	buyOrder := ibsync.LimitOrder("BUY", ibsync.StringToDecimal("20001"), orderprice)
	buyTrade := ib.PlaceOrder(eurusd, buyOrder)

	fmt.Println("submitted buy trade:", buyTrade)
	go func() {
		<-buyTrade.Done()
		fmt.Println("The buy trade is done!!!")
	}()

	// Place additional buy orders
	buyOrder2 := ibsync.LimitOrder("BUY", ibsync.StringToDecimal("20002"), orderprice)
	buyTrade2 := ib.PlaceOrder(eurusd, buyOrder2)
	buyOrder3 := ibsync.LimitOrder("BUY", ibsync.StringToDecimal("20003"), orderprice)
	buyTrade3 := ib.PlaceOrder(eurusd, buyOrder3)
	buyOrder4 := ibsync.LimitOrder("BUY", ibsync.StringToDecimal("20004"), orderprice)
	buyTrade4 := ib.PlaceOrder(eurusd, buyOrder4)
	buyOrder5 := ibsync.LimitOrder("BUY", ibsync.StringToDecimal("20005"), orderprice)
	buyTrade5 := ib.PlaceOrder(eurusd, buyOrder5)

	// Wait one second and check trades
	fmt.Println("*** Updated trades ***")
	time.Sleep(1 * time.Second)
	fmt.Println("updated buy trade:", buyTrade)
	fmt.Println("updated sell trade:", sellTrade)

	// Print Trades after placing orders
	time.Sleep(1 * time.Second)
	printTrades("Trades after placing orders", 2, false, ib)

	// Modify order
	buyOrder2.LmtPrice = math.Round(105*eurusdMidpoint.MidPoint) / 100
	mofifiedbuyTrade2 := ib.PlaceOrder(eurusd, buyOrder2)
	<-buyTrade2.Done()
	<-mofifiedbuyTrade2.Done()

	// Print Trades after modifying order
	printTrades("Trades after modifying order", 3, false, ib)

	// Cancel Orders

	// Cancel buy Order
	fmt.Println("*** Cancel buy order ***")
	ib.CancelOrder(buyOrder3, ibsync.NewOrderCancel())
	<-buyTrade3.Done()
	fmt.Println("cancelled buy trade", buyTrade3)

	// Print Trades after canceling order
	printTrades("Trades after canceling order", 4, false, ib)

	// Global Cancel
	fmt.Println("*** Global cancel ***")
	ib.ReqGlobalCancel()
	<-buyTrade4.Done()
	<-buyTrade5.Done()

	// Print Trades after global cancel
	time.Sleep(1 * time.Second)
	printTrades("Trades after global cancel", 5, false, ib)

	// Executions & Fills

	// Executions & Fills from state, i.e for this session
	fills := ib.Fills()
	fmt.Printf("*** Executions & Fills from state, total number: %v ***\n", len(fills))
	execs := ib.Executions()
	for i, exec := range execs {
		fmt.Printf("*%v* %v \n", i, exec)
	}
	for i, fill := range fills {
		fmt.Printf("*%v* %v \n", i, fill)
	}

	// Executions & Fills after filtered request
	ef := ibsync.NewExecutionFilter()
	ef.Side = "BUY"
	execs, err = ib.ReqExecutions(ef)
	if err != nil {
		log.Error().Err(err).Msg("Request Executions")
		return
	}
	fmt.Printf("*** Executions & Fills after filtered request, total number: %v ***\n ***", len(execs))
	for i, exec := range execs {
		fmt.Printf("*%v* %v \n", i, exec)
	}
	fills, err = ib.ReqFills()
	if err != nil {
		log.Error().Err(err).Msg("Request Fills")
		return
	}
	for i, fill := range fills {
		fmt.Printf("*%v* %v \n", i, fill)
	}

	// Executions & Fills after no filter request
	execs, err = ib.ReqExecutions()
	if err != nil {
		log.Error().Err(err).Msg("Request Executions")
		return
	}
	fmt.Printf("*** Executions & Fills after no filter request, total number: %v ***\n ***", len(execs))
	for i, exec := range execs {
		fmt.Printf("*%v* %v \n", i, exec)
	}
	fills, err = ib.ReqFills()
	if err != nil {
		log.Error().Err(err).Msg("Request Fills")
		return
	}
	for i, fill := range fills {
		fmt.Printf("*%v* %v \n", i, fill)
	}

	log.Info().Msg("Good Bye!!!")
}
