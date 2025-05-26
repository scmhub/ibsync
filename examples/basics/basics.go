// the basics file demonstrates the usage of the ibsync library for
// connecting to Interactive Brokers API and retrieving various
// financial account and market information.
//
// This example script showcases:
// - Configuring and connecting to Interactive Brokers
// - Logging configuration
// - Retrieving account-related information
// - Subscribing to and managing real-time data streams
// - Handling trades, orders, and executions
// - Receiving news bulletins
//
// Key operations include:
// - Establishing a connection with specific configuration
// - Fetching managed accounts
// - Retrieving account values and portfolio
// - Subscribing to positions and P&L
// - Monitoring trades and orders
// - Requesting and processing news bulletins
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
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
	ibsync.SetLogLevel(int(zerolog.InfoLevel))
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
			// ibsync.WithoutSync(),
			ibsync.WithTimeout(10*time.Second),
		),
	)
	if err != nil {
		log.Error().Err(err).Msg("Connect")
		return
	}
	defer ib.Disconnect()

	// Managed accounts
	managedAccounts := ib.ManagedAccounts()
	log.Info().Strs("accounts", managedAccounts).Msg("Managed accounts list")

	// Account Values
	accountValues := ib.AccountValues()
	fmt.Println("acount values", ibsync.AccountSummary(accountValues))

	// Account Summary
	accountSummary := ib.AccountSummary()
	fmt.Println("account summary", accountSummary)

	// Portfolio
	portfolio := ib.Portfolio()
	fmt.Println("portfolio", portfolio)

	// Positions
	// Subscribe to Postion
	ib.ReqPositions()
	// Position Channel
	posChan := ib.PositionChan()
	go func() {
		for pos := range posChan {
			fmt.Println("Position from chan:", pos)
		}
	}()
	time.Sleep(1 * time.Second)
	positions := ib.Positions()
	fmt.Println("positions", positions)
	// Cancel position subscription
	ib.CancelPositions()

	// Pnl
	// Subscribe to P&L
	ib.ReqPnL(managedAccounts[0], "")
	// P&l Channel
	pnlChan := ib.PnlChan(managedAccounts[0], "")
	go func() {
		for pnl := range pnlChan {
			fmt.Println("P&L from chan:", pnl)
		}
	}()
	time.Sleep(1 * time.Second)
	// Get P&L for specific account
	pnl := ib.Pnl(managedAccounts[0], "")
	fmt.Println("pnl", pnl)
	// Cancel P&l subscription
	ib.CancelPnL(managedAccounts[0], "")

	// Trades
	trades := ib.Trades()
	fmt.Println("trades", trades)
	openTrades := ib.OpenTrades()
	fmt.Println("open trades", openTrades)

	// Orders
	orders := ib.Orders()
	fmt.Println("orders", orders)
	openOrders := ib.OpenOrders()
	fmt.Println("open orders", openOrders)

	// Get previous sessions executions and fills
	_, err = ib.ReqExecutions()
	if err != nil {
		log.Error().Err(err).Msg("ReqExecutions")
		return
	}

	// Fills
	fills := ib.Fills()
	fmt.Println("fills", fills)

	// Executions
	executions := ib.Executions()
	fmt.Println("executions", executions)

	// User info
	whiteBrandingId, _ := ib.ReqUserInfo()
	fmt.Println("whiteBrandingId", whiteBrandingId)

	// News bulletins Channel
	nbChan := ib.NewsBulletinsChan()
	ctx, cancel := context.WithCancel(ib.Context())
	defer cancel()
	go func() {
		var i int
		for {
			select {
			case <-ctx.Done():
				return
			case bulletin, ok := <-nbChan:
				if !ok {
					return
				}
				fmt.Printf("News bulletin from channel %v: %v\n", i, bulletin)
				i++
			}
		}
	}()

	// Request news bulletins
	ib.ReqNewsBulletins(true)

	// Wait for bulletins
	time.Sleep(10 * time.Second)

	// Recorded bulletins
	bulletins := ib.NewsBulletins()

	for i, bulletin := range bulletins {
		fmt.Printf("News bulletin %v: %v\n", i, bulletin)
	}
	ib.CancelNewsBulletins()

	time.Sleep(1 * time.Second)
	log.Info().Msg("Good Bye!!!")
}
