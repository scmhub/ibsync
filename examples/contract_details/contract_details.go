package main

import (
	"fmt"
	"strings"
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

	// Request matching symbols

	cdescs, err := ib.ReqMatchingSymbols("amd")
	if err != nil {
		log.Error().Err(err).Msg("request matching symbols")
	}
	for i, cdesc := range cdescs {
		fmt.Printf("Contract description %v: %v\n", i, cdesc)
	}

	// Request contract details

	// NewStock("XXX", "XXX", "XXX") -> No security definition has been found
	amd := ibsync.NewStock("XXX", "XXX", "XXX")
	cds, err := ib.ReqContractDetails(amd)
	if err != nil {
		log.Error().Err(err).Msg("request contract details")
	}
	fmt.Println("number of contract found for request NewStock(\"AMD\", \"\", \"\") :", len(cds))

	// NewStock("AMD", "", "")
	amd = ibsync.NewStock("AMD", "", "")
	cds, err = ib.ReqContractDetails(amd)
	if err != nil {
		log.Error().Err(err).Msg("request contract details")
	}
	fmt.Println("number of contract found for request NewStock(\"AMD\", \"\", \"\") :", len(cds))

	// NewStock("AMD", "", "USD")
	amd = ibsync.NewStock("AMD", "", "USD")
	cds, err = ib.ReqContractDetails(amd)
	if err != nil {
		log.Error().Err(err).Msg("request contract details")
	}
	fmt.Println("number of contract found for request NewStock(\"AMD\", \"\", \"USD\") :", len(cds))

	// NewStock("AMD", "SMART", "USD")
	amd = ibsync.NewStock("AMD", "SMART", "USD")
	cds, err = ib.ReqContractDetails(amd)
	if err != nil {
		log.Error().Err(err).Msg("request contract details")
	}
	fmt.Println("number of contract found for request NewStock(\"AMD\", \"SMART\", \"USD\") :", len(cds))
	fmt.Println(cds[0])

	// get all US T-Notes
	tNote := ibsync.NewBond("US-T", "SMART", "USD")
	cds, err = ib.ReqContractDetails(tNote)
	if err != nil {
		log.Error().Err(err).Msg("request contract details")
	}
	fmt.Println("number of contract found for request NewBond(\"US-T\", \"SMART\", \"USD\") :", len(cds))
	// Get the 10 years Notes
	for i, cd := range cds {
		split := strings.Split(cd.DescAppend, " ")
		maturity, err := time.Parse("01/02/06", split[len(split)-1])
		if err != nil {
			log.Error().Err(err).Msg("couldn't parse maturity")
		}
		if maturity.Year() == time.Now().Year()+10 {
			fmt.Printf("bond %v: %v\n", i, cd)
		}

	}

	// Qualify contract

	// NewStock("XXX", "XXX", "XXX") -> No security definition has been found
	amd = ibsync.NewStock("XXX", "XXX", "XXX")
	err = ib.QualifyContract(amd)
	if err != nil {
		log.Error().Err(err).Msg("qualify contract details")
	}

	// NewStock("AMD", "", "") -> ambiguous contract
	amd = ibsync.NewStock("AMD", "", "")
	err = ib.QualifyContract(amd)
	if err != nil {
		log.Error().Err(err).Msg("qualify contract details")
	}

	// Qualifiable
	amd = ibsync.NewStock("AMD", "SMART", "USD")
	log.Info().Stringer("AMD", amd).Msg("AMD before qualifiying")
	err = ib.QualifyContract(amd)
	if err != nil {
		log.Error().Err(err).Msg("qualify contract details")
	}
	log.Info().Stringer("AMD", amd).Msg("AMD after qualifiying")

	// Qualify bond from CUSIP
	tNote = ibsync.NewBond("91282CLW9", "SMART", "USD")
	err = ib.QualifyContract(tNote)
	if err != nil {
		log.Error().Err(err).Msg("qualify contract details")
	}
	log.Info().Stringer("91282CLW9", tNote).Msg("US 10 years Notes")

	log.Info().Msg("Good Bye!!!")
}
