package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"

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

func save2File(xmlData string, filePath string) {
	file, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = file.WriteString(xmlData)
	if err != nil {
		panic(err)
	}
}

func parseXML(xmlData string) ([]string, error) {
	// Define types for XML structure
	type AbstractField struct {
		Code string `xml:"code"`
	}
	type RangeFilter struct {
		AbstractFields []AbstractField `xml:"AbstractField"`
	}
	type FilterList struct {
		RangeFilters []RangeFilter `xml:"RangeFilter"`
	}

	type ScanParameterResponse struct {
		XMLName     xml.Name     `xml:"ScanParameterResponse"`
		FilterLists []FilterList `xml:"FilterList"`
	}

	// Parse the XML
	var response ScanParameterResponse
	err := xml.Unmarshal([]byte(xmlData), &response)
	if err != nil {
		return nil, err
	}

	// Extract all <code> values
	var codes []string
	for _, fls := range response.FilterLists {
		for _, rfs := range fls.RangeFilters {
			for _, af := range rfs.AbstractFields {
				codes = append(codes, af.Code)
			}
		}
	}

	return codes, nil
}

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

	// Scanner Parameter
	xml, err := ib.ReqScannerParameters()
	if err != nil {
		log.Error().Err(err).Msg("Request scanner parameters")
		return
	}
	// Save xml scanner parameters to scanner_parameters.xml file
	save2File(xml, "scanner_parameters.xml")

	// Scanner subscription
	scanSubscription := ibsync.NewScannerSubscription()
	scanSubscription.Instrument = "STK"
	scanSubscription.LocationCode = "STK.US.MAJOR"
	scanSubscription.ScanCode = "TOP_PERC_GAIN"

	scanDatas, err := ib.ReqScannerSubscription(scanSubscription)
	if err != nil {
		log.Error().Err(err).Msg("Request scanner subscription")
		return
	}
	fmt.Println("scanner subscription")
	for _, scan := range scanDatas {
		fmt.Printf("%v\n", scan)
	}

	// Filter scanner the old way
	oldFilterSubscrition := scanSubscription
	oldFilterSubscrition.AbovePrice = 5
	oldFilterSubscrition.BelowPrice = 50
	oldFilterScanDatas, err := ib.ReqScannerSubscription(oldFilterSubscrition)
	if err != nil {
		log.Error().Err(err).Msg("Request old filter scanner subscription")
		return
	}

	fmt.Println("old filter scanner subscription")
	for _, scan := range oldFilterScanDatas {
		fmt.Printf("%v\n", scan)
	}

	// Filter scanner data the new way
	// View all tags
	tags, err := parseXML(xml)
	if err != nil {
		log.Error().Err(err).Msg("parsing xml string")
		return
	}
	fmt.Printf("nb tags: %v, first 10 tags:%v...\n", len(tags), strings.Join(tags[:10], ", "))

	// AbovPrice is now priceAbove
	opts := ibsync.ScannerSubscriptionOptions{
		FilterOptions: []ibsync.TagValue{
			{Tag: "changePercAbove", Value: "20"},
			{Tag: "priceAbove", Value: "5"},
			{Tag: "priceBelow", Value: "50"},
		},
	}

	newFilterScanDatas, err := ib.ReqScannerSubscription(scanSubscription, opts)
	if err != nil {
		log.Error().Err(err).Msg("Request new filter scanner subscription")
		return
	}
	fmt.Println("new filter scanner subscription")
	for i, scan := range newFilterScanDatas {
		fmt.Printf("%v - %v\n", i, scan)
	}

	log.Info().Msg("Good Bye!!!")
}
