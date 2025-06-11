// Note:
//   - Most of the tests in this file depend on the market being open to function correctly.
//     Ensure the market is open before running these tests to avoid failures or unexpected behavior.
//   - The tests are designed to run exclusively on paper trading accounts.
//     Attempting to run these tests on a live trading account will do nothing and mark all the tests as failed.
package ibsync

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"testing"
	"time"
)

const (
	testHost           = "localhost"
	testPort           = 7497
	testClientID       = 1973
	testTimeOut        = 30 * time.Second
	account            = "DU5352527"
	modelCode          = ""
	contractID   int64 = 756733
)

var testConfig = NewConfig(
	WithHost(testHost),
	WithPort(testPort),
	WithClientID(testClientID),
	WithTimeout(testTimeOut),
)
var prettyFlag bool
var logLevel int

func init() {
	testing.Init()
	flag.IntVar(&logLevel, "logLevel", 1, "log Level: -1:trace, 0:debug, 1:info, 2:warning, 3:error, 4:fatal, 5:panic")
	flag.BoolVar(&prettyFlag, "pretty", false, "enable pretty printing")
}

var globalIB *IB // Global IB client for batch testing

func getIB() *IB {
	if globalIB != nil {
		return globalIB
	}
	globalIB = NewIB(testConfig)
	if err := globalIB.Connect(); err != nil {
		panic("Failed to connect to IB: " + err.Error())
	}

	if !globalIB.IsPaperAccount() {
		panic("Tests must run on a paper trading account")
	}
	return globalIB
}

// TestMain handles setup and teardown for the entire test suite.
func TestMain(m *testing.M) {
	// Setup phase

	// Parse flags first
	flag.Parse()

	// Log level
	SetLogLevel(logLevel)

	// Pretty
	if prettyFlag {
		SetConsoleWriter()
	}

	// Run the tests
	code := m.Run()

	// Teardown phase
	if globalIB != nil {
		if err := globalIB.Disconnect(); err != nil {
			panic("Failed to disconnect IB client: " + err.Error())
		}
	}

	// Exit with the test result code
	os.Exit(code)
}

func TestConnection(t *testing.T) {
	ib := getIB()

	if !ib.IsConnected() {
		t.Fatal("client not connected")
	}

	// Disconnect
	if err := ib.Disconnect(); err != nil {
		t.Errorf("Failed to disconnect IB client: %v", err)
	}
	time.Sleep(1 * time.Second)
	if ib.IsConnected() {
		t.Errorf("client should be disconnected")
	}

	// Reconnect
	if err := ib.Connect(); err != nil {
		panic("Failed to reconnect to IB: " + err.Error())
	}

	if !ib.IsConnected() {
		t.Errorf("client should be disconnected")
	}

	// CurrentTime
	time.Sleep(1 * time.Second)
	currentTime, err := ib.ReqCurrentTime()
	if err != nil {
		t.Errorf("ReqCurrentTime: %v", err)
		return
	}
	lag := time.Since(currentTime)
	t.Logf("CurrentTime: %v, lag: %v.\n", currentTime, lag)
	if lag >= 3*time.Second {
		t.Error("CurrentTime lag is too high", lag)
	}

	// CurrentTimeInMillis
	time.Sleep(1 * time.Second)
	currentTimeInMillis, err := ib.ReqCurrentTimeInMillis()
	if err != nil {
		t.Errorf("ReqCurrentTimeInMillis: %v", err)
		return
	}
	lag = time.Since(time.UnixMilli(currentTimeInMillis))
	t.Logf("CurrentTimeMillis: %v, lag: %v.\n", currentTimeInMillis, lag)
	if lag >= 3*time.Second {
		t.Error("CurrentTimeInMillis lag is too high", lag)
	}

	// Server version
	serverVersion := ib.ServerVersion()
	t.Log("Server version", serverVersion)

	// Managed accounts
	managedAccounts := ib.ManagedAccounts()
	if len(managedAccounts) < 1 {
		t.Fatal("no accounts")
	}
	if testing.Verbose() {
		for i, ma := range managedAccounts {
			t.Logf("Managed account %v: %v\n", i, ma)
		}
	}

	// Account values
	accountValues := ib.AccountValues()
	if len(accountValues) < 1 {
		t.Error("no account values")
	}
	if testing.Verbose() {
		for i, av := range accountValues {
			t.Logf("Account values %v: %v\n", i, av)
		}
	}

	// Account summary
	accountSummary := ib.AccountSummary()
	if len(accountSummary) < 1 {
		t.Error("no account summary")
	}
	if testing.Verbose() {
		for i, as := range accountSummary {
			t.Logf("Account summary %v: %v\n", i, as)
		}
	}

	// Portfolio
	portfolio := ib.Portfolio()
	if len(portfolio) < 1 {
		t.Error("no portfolio")
	}
	if testing.Verbose() {
		for i, p := range portfolio {
			t.Logf("Portfolio item %v: %v\n", i, p)
		}
	}
	// Orders
	orders := ib.Orders()
	for i, order := range orders {
		t.Logf("Order %v: %v\n", i, order)
	}

	// OpenOrders
	openOrders := ib.OpenOrders()
	for i, openOrder := range openOrders {
		t.Logf("Open order %v: %v\n", i, openOrder)
	}

	// TWS connection time
	ct := ib.TWSConnectionTime()
	if testing.Verbose() {
		t.Logf("Connection time: %v\n", ct)
	}
}

func TestMultipleConnections(t *testing.T) {
	// Client #1
	ib1 := NewIB(NewConfig(
		WithHost(testHost),
		WithPort(testPort),
		WithClientID(1001),
	))
	if err := ib1.Connect(); err != nil {
		panic("Failed to connect to IB: " + err.Error())
	}
	defer ib1.Disconnect()

	// Client #2
	ib2 := NewIB(NewConfig(
		WithHost(testHost),
		WithPort(testPort),
		WithClientID(1002),
	))
	if err := ib2.Connect(); err != nil {
		panic("Failed to connect to IB: " + err.Error())
	}
	defer ib2.Disconnect()

	// Client #3
	ib3 := NewIB(NewConfig(
		WithHost(testHost),
		WithPort(testPort),
		WithClientID(1003),
	))
	if err := ib3.Connect(); err != nil {
		panic("Failed to connect to IB: " + err.Error())
	}
	defer ib3.Disconnect()

	// Client #4
	ib4 := NewIB(NewConfig(
		WithHost(testHost),
		WithPort(testPort),
		WithClientID(1004),
	))
	if err := ib4.Connect(); err != nil {
		panic("Failed to connect to IB: " + err.Error())
	}
	defer ib4.Disconnect()

	if !ib1.IsConnected() {
		t.Fatal("client 1 not connected")
	}
	if !ib2.IsConnected() {
		t.Fatal("client 2 not connected")
	}
	if !ib3.IsConnected() {
		t.Fatal("client 3 not connected")
	}
	if !ib4.IsConnected() {
		t.Fatal("client 4 not connected")
	}
}

func TestPositions(t *testing.T) {
	ib := getIB()

	ib.ReqPositions()
	defer ib.CancelPositions()

	posChan := ib.PositionChan()
	go func() {
		for pos := range posChan {
			t.Log("Position from chan:", pos)
		}
	}()

	time.Sleep(1 * time.Second)

	positions := ib.Positions()
	t.Log("positions", positions)

}

func TestPnl(t *testing.T) {
	ib := getIB()

	ib.ReqPnL(account, modelCode)

	pnlChan := ib.PnlChan(account, modelCode)
	go func() {
		for pnl := range pnlChan {
			t.Log("pnl from chan", pnl)
		}
	}()

	time.Sleep(1 * time.Second)

	pnl := ib.Pnl(account, modelCode)
	t.Log("pnl", pnl)

	time.Sleep(3 * time.Second)

	ib.CancelPnL(account, modelCode)

	time.Sleep(1 * time.Second)

	pnl = ib.Pnl(account, modelCode)
	t.Log("pnl", pnl)
}

func TestPnlSingle(t *testing.T) {
	ib := getIB()

	ib.ReqPnLSingle(account, modelCode, contractID)

	pnlSingleChan := ib.PnlSingleChan(account, modelCode, contractID)
	go func() {
		for pnlSingle := range pnlSingleChan {
			t.Log("pnl from chan", pnlSingle)
		}
	}()

	time.Sleep(1 * time.Second)

	pnlSingle := ib.PnlSingle(account, modelCode, contractID)
	t.Log("pnlSingle", pnlSingle)

	time.Sleep(3 * time.Second)

	ib.CancelPnLSingle(account, modelCode, contractID)

	time.Sleep(2 * time.Second)

	pnlSingle = ib.PnlSingle(account, modelCode, contractID)
	t.Log("pnlSingle", pnlSingle)
}

func TestMidPoint(t *testing.T) {
	ib := getIB()
	eurusd := NewForex("EUR", "IDEALPRO", "USD")

	midpoint, err := ib.MidPoint(eurusd)
	if err != nil {
		t.Fatalf("Failed to get midpoint: %v", err)
	}
	if testing.Verbose() {
		t.Logf("MidPoint: %v", midpoint)
	}
}

func TestCalculateOption(t *testing.T) {
	ib := getIB()

	ib.ReqMarketDataType(4)

	spx := NewIndex("SPX", "CBOE", "USD")

	err := ib.QualifyContract(spx)
	if err != nil {
		t.Fatal("Qualify contract")
	}

	ticker, err := ib.Snapshot(spx)
	if err != nil && err != WarnDelayedMarketData {
		t.Fatalf("Failed to get ticker: %v", err)
	}

	maturity := time.Now().AddDate(0, 3, 0).Format("200601") // three month from now
	strike := math.Round(ticker.MarketPrice()/250) * 250
	call := NewOption("SPX", maturity, strike, "C", "SMART", "100", "USD")
	call.TradingClass = "SPX"

	err = ib.QualifyContract(call)
	if err != nil {
		t.Fatal("Qualify options")
	}

	ticker, err = ib.Snapshot(call)
	greeks := ticker.Greeks()

	if testing.Verbose() {
		t.Log("err", err)
		t.Log("Greeks", greeks)
	}

	optionPrice, err := ib.CalculateOptionPrice(call, greeks.ImpliedVol+0.01, greeks.UndPrice)
	if err != nil {
		t.Errorf("CalculateOptionPrice: %v\n", err)
		return
	}

	if testing.Verbose() {
		t.Logf("Option Price: %v, was expecting: %v", optionPrice.OptPrice, greeks.OptPrice+greeks.Vega)
	}

	impliedVol, err := ib.CalculateImpliedVolatility(call, greeks.OptPrice+greeks.Vega, greeks.UndPrice)
	if err != nil {
		t.Errorf("CalculateImpliedVolatility: %v\n", err)
		return
	}

	if testing.Verbose() {
		t.Logf("Implied Volatility: %.2f%%, was expecting: %.2f%%", impliedVol.ImpliedVol*100, greeks.ImpliedVol*100+1)
	}
}

func TestPlaceOrder(t *testing.T) {
	ib := getIB()

	eurusd := NewForex("EUR", "IDEALPRO", "USD")

	midpoint, err := ib.MidPoint(eurusd)
	if err != nil {
		t.Fatalf("Failed to get midpoint: %v", err)
	}
	price := math.Round(95*midpoint.MidPoint) / 100
	modifiedPrice := math.Round(105*midpoint.MidPoint) / 100

	// Place orders
	order1 := LimitOrder("SELL", StringToDecimal("20001"), price)
	trade1 := ib.PlaceOrder(eurusd, order1)

	order2 := LimitOrder("BUY", StringToDecimal("20002"), price)
	trade2 := ib.PlaceOrder(eurusd, order2)

	// Wait for order1 to be acknowledged
	select {
	case <-trade1.Done():
	case <-time.After(5 * time.Second):
		t.Fatal("Order1 placement timed out")
	}

	// Trades
	trades := ib.Trades()
	for _, trade := range trades {
		if trade.Equal(trade1) {
			t.Log("found trade1 in Trades:", trade)
		}
	}

	// OpenTrades
	openTrades := ib.OpenTrades()
	for _, openTrade := range openTrades {
		if openTrade.Equal(trade2) {
			t.Log("found trade2 in openTrades:", openTrade)
		}
	}

	// Orders
	orders := ib.Orders()
	for _, order := range orders {
		if order.HasSameID(order1) {
			t.Log("found order1 in Orders:", order)
		}
	}

	// OpenOrders
	openOrders := ib.OpenOrders()
	for _, openOrder := range openOrders {
		if openOrder.HasSameID(order2) {
			t.Log("found order2 in openOrders:", openOrder)
		}
	}

	// Modify order
	order2.LmtPrice = modifiedPrice
	_ = ib.PlaceOrder(eurusd, order2)
	// Wait for order2 to be acknowledged
	select {
	case <-trade2.Done():
	case <-time.After(5 * time.Second):
		t.Fatal("Order2 modification timed out")
	}

	// Request executions
	execs, err := ib.ReqExecutions()
	if err != nil {
		t.Error("Request Executions")
		return
	}
	if len(execs) < 2 {
		t.Errorf("not enough executions: %v", len(execs))
		return
	}

	// Request fills
	fills, err := ib.ReqFills()
	if err != nil {
		t.Error("Fill Executions")
		return
	}
	if len(fills) < 2 {
		t.Errorf("not enough fills: %v", len(fills))
		return
	}

}

func TestCancelOrder(t *testing.T) {
	ib := getIB()

	eurusd := NewForex("EUR", "IDEALPRO", "USD")

	midpoint, err := ib.MidPoint(eurusd)
	if err != nil {
		t.Fatalf("Failed to get midpoint: %v", err)
	}
	price := math.Round(95*midpoint.MidPoint) / 100

	// Place order
	order := LimitOrder("BUY", StringToDecimal("20001"), price)
	trade := ib.PlaceOrder(eurusd, order)

	// Cancel the order
	ib.CancelOrder(order, NewOrderCancel())

	// Wait for order to be cancelled
	select {
	case <-trade.Done():
	case <-time.After(5 * time.Second):
		t.Fatal("Order cancel timed out")
	}

	if trade.OrderStatus.Status != Cancelled {
		t.Errorf("Expected Cancelled status, got %v", trade.OrderStatus.Status)
	}
}

func TestGlobalCancel(t *testing.T) {
	ib := getIB()

	eurusd := NewForex("EUR", "IDEALPRO", "USD")

	midpoint, err := ib.MidPoint(eurusd)
	if err != nil {
		t.Fatalf("Failed to get midpoint: %v", err)
	}
	price := math.Round(95*midpoint.MidPoint) / 100

	// Place orders
	order1 := LimitOrder("BUY", StringToDecimal("20001"), price)
	trade1 := ib.PlaceOrder(eurusd, order1)

	order2 := LimitOrder("BUY", StringToDecimal("20002"), price)
	trade2 := ib.PlaceOrder(eurusd, order2)

	// Execute global cancel
	ib.ReqGlobalCancel()

	// Wait for order1 to be cancelled
	select {
	case <-trade1.Done():
	case <-time.After(5 * time.Second):
		t.Fatal("Order cancel timed out")
	}
	// Wait for order2 to be cancelled
	select {
	case <-trade2.Done():
	case <-time.After(5 * time.Second):
		t.Fatal("Order cancel timed out")
	}
	if !(trade1.OrderStatus.Status == Cancelled || trade1.OrderStatus.Status == ApiCancelled) {
		t.Errorf("Expected Cancelled status for trade1, got %v", trade1.OrderStatus.Status)
	}
	if !(trade2.OrderStatus.Status == Cancelled || trade2.OrderStatus.Status == ApiCancelled) {
		t.Errorf("Expected Cancelled status for trade2, got %v", trade2.OrderStatus.Status)
	}
}

func TestSnapshot(t *testing.T) {
	ib := getIB()

	eurusd := NewForex("EUR", "IDEALPRO", "USD")

	ticker, err := ib.Snapshot(eurusd)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if testing.Verbose() {
		t.Logf("Market price: %v\n", ticker.MarketPrice())
		t.Logf("Snapshot: %v\n", ticker)

	}

}

func TestReqSmartComponent(t *testing.T) {
	ib := getIB()

	smartComponents, err := ib.ReqSmartComponents("c70003")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if testing.Verbose() {
		for i, sc := range smartComponents {
			t.Logf("Smart component %v: %v\n", i, sc)
		}
	}

}

func TestReqMarketRule(t *testing.T) {
	ib := getIB()

	priceIncrement, err := ib.ReqMarketRule(26)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if testing.Verbose() {
		t.Logf("Price increment: %v\n", priceIncrement)
	}
}

func TestReqTickByTickData(t *testing.T) {
	ib := getIB()

	eurusd := NewForex("EUR", "IDEALPRO", "USD")

	tickByTick := ib.ReqTickByTickData(eurusd, "BidAsk", 100, true)

	time.Sleep(5 * time.Second)

	err := ib.CancelTickByTickData(eurusd, "BidAsk")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if testing.Verbose() {
		t.Logf("Tick by tick: %v\n", tickByTick)
	}

}

func TestReqContractDetails(t *testing.T) {
	ib := getIB()

	amd := NewStock("XXX", "XXX", "XXX")
	_, err := ib.ReqContractDetails(amd)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	amd = NewStock("AMD", "", "")
	cds, err := ib.ReqContractDetails(amd)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(cds) < 2 {
		t.Errorf("Expected more contract details got: %v", len(cds))
	}
	if testing.Verbose() {
		for i, cd := range cds {
			t.Logf("contract detail %v:, %v\n", i, cd)
		}
	}

	amd = NewStock("AMD", "SMART", "USD")
	cds, err = ib.ReqContractDetails(amd)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(cds) > 1 {
		t.Errorf("Expected one contract details got: %v", len(cds))
	}
	if testing.Verbose() {
		for i, cd := range cds {
			t.Logf("contract detail %v:, %v\n", i, cd)
		}
	}
}

func TestQualifyContract(t *testing.T) {
	ib := getIB()

	amd := NewStock("XXX", "XXX", "XXX")
	err := ib.QualifyContract(amd)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	t.Logf("No security definition has been found for the request, %v", err)

	amd = NewStock("AMD", "", "")
	err = ib.QualifyContract(amd)
	if err == nil {
		t.Error("Expected error, got nil")
	}
	t.Logf("ambiguous contract, %v", err)

	amd = NewStock("AMD", "SMART", "USD")
	t.Logf("AMD before qualifiying:, %v", amd)
	err = ib.QualifyContract(amd)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	t.Logf("AMD after qualifiying:, %v", amd)
}

func TestJsonCOntract(t *testing.T) {
	ib := getIB()

	amd := NewStock("AMD", "SMART", "USD")
	err := ib.QualifyContract(amd)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	byteContract, err := json.Marshal(amd)
	if err != nil {
		t.Errorf("json Marshall error: %v", err)
	}
	if testing.Verbose() {
		t.Logf("json contract: %v\n", string(byteContract))
	}

	var decodedContract Contract
	err = json.Unmarshal(byteContract, &decodedContract)
	if err != nil {
		t.Errorf("json Unmarshall error: %v", err)
	}
	if !decodedContract.Equal(amd) {
		t.Errorf("Decoded contract does not match original contract: %v", amd)
	}
}

func TestReqMktDepthExchanges(t *testing.T) {
	ib := getIB()

	mdes, err := ib.ReqMktDepthExchanges()

	if err != nil {
		t.Errorf("Unexpected error %v\n", err)
	}

	if len(mdes) < 1 {
		t.Error("no market depth exchanges")
	}

	t.Logf("No security definition has been found for the request, %v", err)
	if testing.Verbose() {
		for i, mde := range mdes {
			t.Logf("Depth market data description %v: %v\n", i, mde)
		}
	}
}

func TestReqMktDepth(t *testing.T) {
	ib := getIB()

	aapl := NewStock("AAPL", "NYSE", "USD")

	ticker, err := ib.ReqMktDepth(aapl, 5, false)

	if err == ErrAdditionalSubscriptionRequired {
		t.Log("no market data subscription for Market depth")
		return
	}

	if err != nil {
		t.Fatalf("Unexpected error %v\n", err)
	}

	time.Sleep(2 * time.Second)

	bids := ticker.DomBids()
	asks := ticker.DomAsks()

	if len(bids) == 0 || len(asks) == 0 {
		t.Error("no market depth data")
	}
	if testing.Verbose() {
		for level := range len(bids) {
			t.Logf("level %v, bid:%v, ask:%v", level, bids[level], asks[level])
		}
	}

	ib.CancelMktDepth(aapl, false)
}

func TestNewsBulletins(t *testing.T) {
	ib := getIB()

	nbChan := ib.NewsBulletinsChan()
	ctx, cancel := context.WithCancel(ib.eClient.Ctx)
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
				t.Logf("News bulletin from channel %v: %v\n", i, bulletin)
				i++
			}
		}
	}()

	ib.ReqNewsBulletins(true)
	defer ib.CancelNewsBulletins()

	time.Sleep(2 * time.Second)

	bulletins := ib.NewsBulletins()

	for i, bulletin := range bulletins {
		t.Logf("News bulletin %v: %v\n", i, bulletin)
	}

}

func TestRequestFA(t *testing.T) {
	ib := getIB()

	cxml, err := ib.RequestFA(FaDataType(1))

	if err == ErrNotFinancialAdvisor {
		t.Log("RequestFA not allowed on non FA account")
		return
	}

	if err != nil {
		t.Errorf("Unexpected error %v\n", err)
	}

	if testing.Verbose() {
		t.Logf("FA cxml:\n %v\n", cxml)
	}
}

func TestReplaceFA(t *testing.T) {
	ib := getIB()

	cxml, err := ib.RequestFA(FaDataType(1))

	if err == ErrNotFinancialAdvisor {
		t.Log("ReplaceFA not allowed on non FA account")
		return
	}

	cxml, err = ib.ReplaceFA(FaDataType(1), cxml)

	if err != nil {
		t.Errorf("Unexpected error %v\n", err)
	}

	if testing.Verbose() {
		t.Logf("FA cxml:\n %v\n", cxml)
	}
}

func TestReqHistoricalData(t *testing.T) {
	ib := getIB()

	eurusd := NewForex("EUR", "IDEALPRO", "USD")

	lastWednesday12ESTString := FormatIBTimeUSEastern(LastWednesday12EST())

	// Request Historical Data
	endDateTime := lastWednesday12ESTString // format "yyyymmdd HH:mm:ss ttt", where "ttt" is an optional time zone
	duration := "1 D"                       // "60 S", "30 D", "13 W", "6 M", "10 Y". The unit must be specified (S for seconds, D for days, W for weeks, etc.).
	barSize := "15 mins"                    // "1 secs", "5 secs", "10 secs", "15 secs", "30 secs", "1 min", "2 mins", "5 mins", etc.
	whatToShow := "MIDPOINT"                // "TRADES", "MIDPOINT", "BID", "ASK", "BID_ASK", "HISTORICAL_VOLATILITY", etc.
	useRTH := true                          // `true` limits data to regular trading hours (RTH), `false` includes all data.
	formatDate := 1                         // `1` for the "yyyymmdd HH:mm:ss ttt" format, or `2` for Unix timestamps.

	barChan, cancel := ib.ReqHistoricalData(eurusd, endDateTime, duration, barSize, whatToShow, useRTH, formatDate)
	cancel()
	<-barChan
	t.Log("Historical request cancelled")

	barChan, _ = ib.ReqHistoricalData(eurusd, endDateTime, duration, barSize, whatToShow, useRTH, formatDate)
	var bars []Bar
	var i int
	for bar := range barChan {
		bars = append(bars, bar)
		if testing.Verbose() {
			t.Logf("bar %v: %v\n", i, bar)
			i++
		}
	}

	if len(bars) < 1 {
		t.Error("No bars retreived")
		return
	}

	t.Log("Number of bars:", len(bars))
	t.Log("FirstBar:", bars[0])
	t.Log("LastBar", bars[len(bars)-1])
}

func TestReqHistoricalDataUpToDate(t *testing.T) {
	ib := getIB()

	eurusd := NewForex("EUR", "IDEALPRO", "USD")
	duration := "60 S"
	barSize := "5 secs"
	whatToShow := "MIDPOINT"
	useRTH := true
	formatDate := 1

	barChan, cancel := ib.ReqHistoricalDataUpToDate(eurusd, duration, barSize, whatToShow, useRTH, formatDate)

	var bars []Bar
	go func() {
		var i int
		for bar := range barChan {
			bars = append(bars, bar)
			if testing.Verbose() {
				t.Logf("bar %v: %v\n", i, bar)
				i++
			}
		}
	}()

	time.Sleep(30 * time.Second)
	cancel()

	if len(bars) < 1 {
		t.Error("No bars retreived")
		return
	}

	t.Log("Number of bars:", len(bars))
	t.Log("FirstBar:", bars[0])
	t.Log("LastBar", bars[len(bars)-1])
}

func TestReqHistoricalSchedule(t *testing.T) {
	ib := getIB()

	eurusd := NewForex("EUR", "IDEALPRO", "USD")

	historicalSchedule, err := ib.ReqHistoricalSchedule(eurusd, "", "1 W", true)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}
	t.Logf("Historical schedule start date:, %v", historicalSchedule.StartDateTime)
	t.Logf("Historical schedule end date:, %v", historicalSchedule.EndDateTime)
	t.Logf("Historical schedule time zone:, %v", historicalSchedule.TimeZone)
	if testing.Verbose() {
		for i, session := range historicalSchedule.Sessions {
			t.Logf("Session %v:, %v", i, session)
		}
	}
}

func TestReqHeadTimeStamp(t *testing.T) {
	ib := getIB()

	eurusd := NewForex("EUR", "IDEALPRO", "USD")

	headStamp, err := ib.ReqHeadTimeStamp(eurusd, "MIDPOINT", true, 1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}
	if !(headStamp.Before(time.Now()) && headStamp.After(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))) {
		t.Errorf("Unexpected error: %v", err)
	}
	t.Logf("headStamp:, %v", headStamp)
}

func TestReqHistogramData(t *testing.T) {
	ib := getIB()

	aapl := NewStock("AAPL", "SMART", "USD")

	histogramDatas, err := ib.ReqHistogramData(aapl, true, "1 day")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if testing.Verbose() {
		for i, hd := range histogramDatas {
			t.Logf("Histogram Data %v: %v\n", i, hd)
		}
	}

}

func TestReqHistoricalTicks(t *testing.T) {
	ib := getIB()

	aapl := NewStock("AAPL", "SMART", "USD")

	ticks, err, done := ib.ReqHistoricalTicks(aapl, LastWednesday12EST(), time.Time{}, 100, true, true)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if testing.Verbose() {
		t.Logf("Historical Ticks number %v, is done? %v\n", len(ticks), done)
		for i, hd := range ticks {
			t.Logf("%v: %v\n", i, hd)
		}
	}

}

func TestReqHistoricalTickLast(t *testing.T) {
	ib := getIB()

	aapl := NewStock("AAPL", "SMART", "USD")

	ticks, err, done := ib.ReqHistoricalTickLast(aapl, LastWednesday12EST(), time.Time{}, 100, true, true)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if testing.Verbose() {
		t.Logf("Historical Last Ticks number %v, is done? %v\n", len(ticks), done)
		for i, hd := range ticks {
			t.Logf("%v: %v\n", i, hd)
		}
	}

}

func TestReqHistoricalTickBidAsk(t *testing.T) {
	ib := getIB()

	aapl := NewStock("AAPL", "SMART", "USD")

	ticks, err, done := ib.ReqHistoricalTickBidAsk(aapl, LastWednesday12EST(), time.Time{}, 100, true, true)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if testing.Verbose() {
		t.Logf("Historical Bid Ask Ticks number %v, is done? %v\n", len(ticks), done)
		for i, hd := range ticks {
			t.Logf("%v: %v\n", i, hd)
		}
	}

}

func TestReqScannerParameters(t *testing.T) {
	ib := getIB()

	xml, err := ib.ReqScannerParameters()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if testing.Verbose() {
		fmt.Println(xml)
	}

}

func TestReqScannerSubscription(t *testing.T) {
	ib := getIB()

	ss := NewScannerSubscription()
	ss.Instrument = "STK"
	ss.LocationCode = "STK.US.MAJOR"
	ss.ScanCode = "HOT_BY_VOLUME"

	scanDatas, err := ib.ReqScannerSubscription(ss)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if testing.Verbose() {
		for i, sd := range scanDatas {
			t.Logf("Scan data %v: %v\n", i, sd)
		}
	}

}

func TestReqRealTimeBars(t *testing.T) {
	ib := getIB()

	eurusd := NewForex("EUR", "IDEALPRO", "USD")

	useRTH := true
	whatToShow := "MIDPOINT" // "TRADES", "MIDPOINT", "BID" or "ASK"
	rtBarChan, cancel := ib.ReqRealTimeBars(eurusd, 5, whatToShow, useRTH)

	var rtBars []RealTimeBar
	go func() {
		var i int
		for rtBar := range rtBarChan {
			rtBars = append(rtBars, rtBar)
			if testing.Verbose() {
				t.Logf("real time bar %v: %v\n", i, rtBar)
				i++
			}
		}
	}()

	time.Sleep(10 * time.Second)
	cancel()

	if len(rtBars) < 1 {
		t.Error("No real time bars retreived")
		return
	}

	t.Log("Number of bars:", len(rtBars))
	t.Log("FirstBar:", rtBars[0])
	t.Log("LastBar", rtBars[len(rtBars)-1])
}

func TestReqFundamentalData(t *testing.T) {
	ib := getIB()

	aapl := NewStock("AAPL", "SMART", "USD")

	// "ReportSnapshot"
	data, err := ib.ReqFundamentalData(aapl, "ReportSnapshot")

	if err != nil {
		switch err {
		case ErrMissingReportType, ErrNewsFeedNotAllowed:
			t.Log(err)
		default:
			t.Errorf("Unexpected error: %v", err)
		}
	}

	if testing.Verbose() && data != "" {
		t.Logf("fundamental data: ReportSnapshot. \n%v", data)
	}

	// "ReportsFinSummary"
	data, err = ib.ReqFundamentalData(aapl, "ReportsFinSummary")

	if err != nil {
		switch err {
		case ErrMissingReportType, ErrNewsFeedNotAllowed:
			t.Log(err)
		default:
			t.Errorf("Unexpected error: %v", err)
		}
	}

	if testing.Verbose() && data != "" {
		t.Logf("fundamental data: Financial Summary. \n%v", data)
	}

	// "ReportRatios"
	data, err = ib.ReqFundamentalData(aapl, "ReportRatios")

	if err != nil {
		switch err {
		case ErrMissingReportType, ErrNewsFeedNotAllowed:
			t.Log(err)
		default:
			t.Errorf("Unexpected error: %v", err)
		}
	}

	if testing.Verbose() && data != "" {
		t.Logf("fundamental data: Financial Ratio. \n%v", data)
	}

	// "ReportsFinStatements"
	data, err = ib.ReqFundamentalData(aapl, "ReportsFinStatements")

	if err != nil {
		switch err {
		case ErrMissingReportType, ErrNewsFeedNotAllowed:
			t.Log(err)
		default:
			t.Errorf("Unexpected error: %v", err)
		}
	}

	if testing.Verbose() && data != "" {
		t.Logf("fundamental data: Financial Statement. \n%v", data)
	}

	// "RESC
	data, err = ib.ReqFundamentalData(aapl, "RESC")

	if err != nil {
		switch err {
		case ErrMissingReportType, ErrNewsFeedNotAllowed:
			t.Log(err)
		default:
			t.Errorf("Unexpected error: %v", err)
		}
	}

	if testing.Verbose() && data != "" {
		t.Logf("fundamental data: Analyst estimates. \n%v", data)
	}

	// "CalendarReport"
	data, err = ib.ReqFundamentalData(aapl, "CalendarReport")

	if err != nil {
		switch err {
		case ErrMissingReportType, ErrNewsFeedNotAllowed:
			t.Log(err)
		default:
			t.Errorf("Unexpected error: %v", err)
		}
	}

	if testing.Verbose() && data != "" {
		t.Logf("fundamental data: Calendar Report. \n%v", data)
	}
}

func TestReqNewsProviders(t *testing.T) {
	ib := getIB()

	newsProvider, err := ib.ReqNewsProviders()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if testing.Verbose() {
		for i, np := range newsProvider {
			t.Logf("News provider %v: %v.\n", i, np)
		}
	}
}

func TestReqHistoricalNews(t *testing.T) {
	ib := getIB()

	aapl := NewStock("AAPL", "SMART", "USD")
	err := ib.QualifyContract(aapl)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Briefing.com General Market Columns -> BRFG.
	// Briefing.com Analyst Actions -> BRFUPDN.
	// Dow Jones News Service -> DJ-N.
	// Dow Jones Real-Time News Asia Pacific -> DJ-RTA.
	// Dow Jones Real-Time News Europe -> DJ-RTE.
	// Dow Jones Real-Time News Global -> DJ-RTG.
	// Dow Jones Real-Time News Pro -> DJ-RTPRO.
	// Dow Jones Newsletters -> DJNL.

	provider := "BRFG"
	startDateTime := time.Now().AddDate(0, 0, -7)
	endDateTime := time.Now()

	historicalNews, err, hasMore := ib.ReqHistoricalNews(aapl.ConID, provider, startDateTime, endDateTime, 50)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if testing.Verbose() {
		t.Logf("Number of headlines: %v, has more? %v", len(historicalNews), hasMore)
		for i, hn := range historicalNews {
			t.Logf("News headline %v: %v.\n", i, hn)
		}
	}
	if len(historicalNews) > 0 {

		article, err := ib.ReqNewsArticle(provider, historicalNews[0].ArticleID)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if testing.Verbose() {
			t.Log(article)
		}

	}

}

func TestQueryDisplayGroups(t *testing.T) {
	ib := getIB()

	groups, err := ib.QueryDisplayGroups()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	t.Logf("display groups:, %v", groups)
}

func TestReqSecDefOptParams(t *testing.T) {
	ib := getIB()

	optionChains, err := ib.ReqSecDefOptParams("IBM", "", "STK", 8314)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	t.Logf("option chains:, %v", optionChains)
}

func TestReqSoftDollarTiers(t *testing.T) {
	ib := getIB()

	sdts, err := ib.ReqSoftDollarTiers()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	t.Logf("Soft Dollar Tiers:, %v", sdts)
}

func TestReqFamilyCodes(t *testing.T) {
	ib := getIB()

	familyCodes, err := ib.ReqFamilyCodes()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	t.Logf("Family codes:, %v", familyCodes)
}

func TestReqMatchingSymbols(t *testing.T) {
	ib := getIB()

	cds, err := ib.ReqMatchingSymbols("aapl")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	for i, cd := range cds {
		t.Logf("Contract descriptions %v:, %v\n", i, cd)
	}

}

func TestReqWshMetaData(t *testing.T) {
	ib := getIB()

	dataJson, err := ib.ReqWshMetaData()

	if err == ErrNewsFeedNotAllowed {
		t.Log(err)
		return
	}

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	t.Logf("Wall Street Horizon Meta data:, %v", dataJson)
}

func TestReqWshEventData(t *testing.T) {
	ib := getIB()

	dataJson, err := ib.ReqWshEventData(NewWshEventData())

	if err == ErrNewsFeedNotAllowed {
		t.Log(err)
		return
	}

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	t.Logf("Wall Street Horizon Event data:, %v", dataJson)
}

func TestReqUserInfo(t *testing.T) {
	ib := getIB()

	whiteBrandingId, err := ib.ReqUserInfo()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	t.Logf("White Branding ID:, %v", whiteBrandingId)
}
