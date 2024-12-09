package main

import (
	"fmt"
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
	//ibsync.SetLogLevel(int(zerolog.DebugLevel))
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

	// Retrieve the list of managed accounts.
	managedAccounts := ib.ManagedAccounts()
	log.Info().Strs("accounts", managedAccounts).Msg("Managed accounts list")

	account := managedAccounts[0] // Use the first managed account for PnL requests.
	modelCode := ""               // Optional model code
	contractID := int64(756733)   // Example contract ID

	// Request and handle PnL updates for the account.
	ib.ReqPnL(account, modelCode)
	pnlChan := ib.PnlChan(account, modelCode)

	// Start a goroutine to listen for PnL updates.
	go func() {
		for pnl := range pnlChan {
			fmt.Println("Received PnL from channel:", pnl)
		}
	}()

	// Allow time for updates and display current PnL.
	time.Sleep(5 * time.Second)
	pnl := ib.Pnl(account, modelCode)
	fmt.Println("Current PnL:", pnl)

	// Cancel PnL requests and check the status.
	ib.CancelPnL(account, modelCode)
	time.Sleep(2 * time.Second)
	pnl = ib.Pnl(account, modelCode)
	fmt.Println("PnL after cancellation:", pnl)

	// Request and handle single PnL data.
	ib.ReqPnLSingle(account, modelCode, contractID)
	pnlSingleChan := ib.PnlSingleChan(account, modelCode, contractID)

	// Start a goroutine to listen for single PnL updates.
	go func() {
		for pnlSingle := range pnlSingleChan {
			fmt.Println("Received single PnL from channel:", pnlSingle)
		}
	}()

	// Allow time for updates and display current single PnL.
	time.Sleep(5 * time.Second)
	pnlSingle := ib.PnlSingle(account, modelCode, contractID)
	fmt.Println("Current single PnL:", pnlSingle)

	// Cancel single PnL requests and check the status.
	ib.CancelPnLSingle(account, modelCode, contractID)
	time.Sleep(2 * time.Second)
	pnlSingle = ib.PnlSingle(account, modelCode, contractID)
	fmt.Println("Single PnL after cancellation:", pnlSingle)

	time.Sleep(1 * time.Second) // Allow time for final logs to flush.
	log.Info().Msg("Good Bye!!!")
}
