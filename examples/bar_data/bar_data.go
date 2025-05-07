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
	host = "localhost"
	// port specifies the IB API server port
	port = 7497
	// clientID is the unique identifier for this client connection
	clientID = 5
)

func main() {
	// We set logger for pretty logs to console
	log := ibsync.Logger()
	ibsync.SetLogLevel(int(zerolog.DebugLevel))
	ibsync.SetConsoleWriter()

	// New IB client & Connect
	ib := ibsync.NewIB()

	err := ib.ConnectWithGracefulShutdown(
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

	// HeadStamp
	headStamp, err := ib.ReqHeadTimeStamp(eurusd, "MIDPOINT", true, 1)
	if err != nil {
		log.Error().Err(err).Msg("ReqHeadTimeStamp")
		return
	}
	fmt.Printf("The Headstamp for EURSUD is %v\n", headStamp)

	lastWednesday12ESTString := ibsync.FormatIBTimeUSEastern(ibsync.LastWednesday12EST())

	// Historical Data
	endDateTime := lastWednesday12ESTString // format "yyyymmdd HH:mm:ss ttt", where "ttt" is an optional time zone
	duration := "1 D"                       // "60 S", "30 D", "13 W", "6 M", "10 Y". The unit must be specified (S for seconds, D for days, W for weeks, etc.).
	barSize := "15 mins"                    // "1 secs", "5 secs", "10 secs", "15 secs", "30 secs", "1 min", "2 mins", "5 mins", etc.
	whatToShow := "MIDPOINT"                // "TRADES", "MIDPOINT", "BID", "ASK", "BID_ASK", "HISTORICAL_VOLATILITY", etc.
	useRTH := true                          // `true` limits data to regular trading hours (RTH), `false` includes all data.
	formatDate := 1                         // `1` for the "yyyymmdd HH:mm:ss ttt" format, or `2` for Unix timestamps.
	barChan, _ := ib.ReqHistoricalData(eurusd, endDateTime, duration, barSize, whatToShow, useRTH, formatDate)

	var bars []ibsync.Bar
	for bar := range barChan {
		fmt.Println(bar)
		bars = append(bars, bar)
	}

	fmt.Println("Number of bars:", len(bars))
	fmt.Println("First Bar", bars[0])
	fmt.Println("Last Bar", bars[len(bars)-1])

	// Historical Data with realtime Updates
	duration = "60 S"
	barSize = "1 secs"
	barChan, cancel := ib.ReqHistoricalDataUpToDate(eurusd, duration, barSize, whatToShow, useRTH, formatDate)

	go func() {
		for bar := range barChan {
			fmt.Println(bar)
			bars = append(bars, bar)
		}

	}()

	time.Sleep(10 * time.Second)
	cancel()

	// Historical schedule
	historicalSchedule, err := ib.ReqHistoricalSchedule(eurusd, endDateTime, duration, useRTH)
	if err != nil {
		log.Error().Err(err).Msg("ReqHistoricalSchedule")
		return
	}

	fmt.Printf("Historical schedule start date: %v, end date: %v, time zone: %v\n", historicalSchedule.StartDateTime, historicalSchedule.EndDateTime, historicalSchedule.TimeZone)
	for i, session := range historicalSchedule.Sessions {
		fmt.Printf("session %v: %v\n", i, session)
	}

	// Real time bars
	whatToShow = "MIDPOINT" // "TRADES", "MIDPOINT", "BID" or "ASK"
	rtBarChan, cancel := ib.ReqRealTimeBars(eurusd, 5, whatToShow, useRTH)

	var rtBars []ibsync.RealTimeBar
	go func() {
		for rtBar := range rtBarChan {
			fmt.Println(rtBar)
			rtBars = append(rtBars, rtBar)
		}

	}()

	time.Sleep(10 * time.Second)
	cancel()

	fmt.Println("Number of RT bars:", len(rtBars))
	fmt.Println("First RT Bar", rtBars[0])
	fmt.Println("Last RT Bar", rtBars[len(rtBars)-1])

	time.Sleep(1 * time.Second)
	log.Info().Msg("Good Bye!!!")
}
