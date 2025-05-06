package ibsync

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/scmhub/ibapi"
)

const MaxSyncedSubAccounts = 50

// A CancelFunc tells an operation to abandon its work. A CancelFunc does not wait for the work to stop.
// A CancelFunc may be called by multiple goroutines simultaneously. After the first call, subsequent calls to a CancelFunc do nothing.
type CancelFunc func()

// IB struct offers direct access to the current state, such as orders, executions, positions, tickers etc.
// This state is automatically kept in sync with the TWS/IBG application.
// IB has most request methods of EClient, with the same names and parameters (except for the reqId parameter which is not needed anymore).
type IB struct {
	state   *ibState
	pubSub  *PubSub
	eClient *ibapi.EClient
	wrapper *WrapperSync
	config  *Config
}

func NewIB(config ...*Config) *IB {
	state := NewState()
	pubSub := NewPubSub()
	wrapper := NewWrapperSync(state, pubSub)
	client := ibapi.NewEClient(wrapper)

	ib := &IB{
		state:   state,
		pubSub:  pubSub,
		eClient: client,
		wrapper: wrapper,
		config:  NewConfig(),
	}

	if len(config) > 0 {
		ib.config = config[0]
	}

	return ib
}

// SetClientLogLevel sets the log level of the client.
// logLevel can be:
// -1 = trace   // zerolog.TraceLevel
//
//	0 = debug   // zerolog.DebugLevel
//	1 = info    // zerolog.InfoLevel
//	2 = warn    // zerolog.WarnLevel
//	3 = error   // zerolog.ErrorLevel
//	4 = fatal   // zerolog.FatalLevel
//	5 = panic   // zerolog.PanicLevel
func (ib *IB) SetClientLogLevel(logLevel int64) {
	SetLogLevel(int(logLevel))
}

// SetConsoleWriter will set pretty log to the console.
func (in *IB) SetConsoleWriter() {
	SetConsoleWriter()
}

// Connect must be called before any other.
// There is no feedback for a successful connection, but a subsequent attempt to connect will return the message "Already connected."
func (ib *IB) Connect(config ...*Config) error {
	if len(config) > 0 {
		ib.config = config[0] // override config
	}

	err := ib.eClient.Connect(ib.config.Host, ib.config.Port, ib.config.ClientID)
	if err != nil {
		return err
	}

	time.Sleep(500 * time.Millisecond)

	// bind manual orders
	if ib.config.ClientID == 0 {
		ib.ReqAutoOpenOrders(true)
	}

	accounts := ib.ManagedAccounts()
	if ib.config.Account == "" && len(accounts) == 1 {
		ib.config.Account = accounts[0]
	}
	if !ib.config.InSync {
		log.Warn().Msg("this client will not be kept in sync with the TWS/IBG application")
		return nil
	}
	// Start sync
	if !ib.config.ReadOnly {
		// Get and sync open orders
		openOrdersChan, _ := ib.pubSub.Subscribe("OpenOrdersEnd")

		ib.ReqOpenOrders()
		<-openOrdersChan
		// Get and sync completed orders
		completedOrdersChan, _ := ib.pubSub.Subscribe("CompletedOrdersEnd")

		ib.ReqCompletedOrders(false)
		<-completedOrdersChan
	}
	if ib.config.Account != "" {
		accountUpdatesChan, _ := ib.pubSub.Subscribe("AccountDownloadEnd")

		ib.ReqAccountUpdates(true, ib.config.Account)
		<-accountUpdatesChan
	}

	log.Info().Msg("client in sync with the TWS/IBG application")
	return nil
}

// ConnectWithGracefulShutdown connects and sets up signal handling for graceful shutdown.
// This is a convenience for simple apps. Advanced users should handle signals themselves.
func (ib *IB) ConnectWithGracefulShutdown(config ...*Config) error {
	err := ib.Connect(config...)
	if err != nil {
		return err
	}
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Warn().Msg("detected termination signal, shutting down gracefully")
		ib.Disconnect()
		os.Exit(0)
	}()
	return nil
}

// Disconnect terminates the connections with TWS.
// Calling this function does not cancel orders that have already been sent.
func (ib *IB) Disconnect() error {
	return ib.eClient.Disconnect()
}

// IsConnected checks if there is a connection to TWS or GateWay
func (ib *IB) IsConnected() bool {
	return ib.eClient.IsConnected()
}

// Context returns the ibsync Context
func (ib *IB) Context() context.Context {
	return ib.eClient.Ctx
}

// SetTimeout sets the timeout for receiving messages from TWS/IBG.
// Default timeout duration is TIMEOUT = 30 * time.Second
func (ib *IB) SetTimeout(Timeout time.Duration) {
	ib.config.Timeout = Timeout
}

// ManagedAccounts returns a list of account names.
func (ib *IB) ManagedAccounts() []string {
	ib.state.mu.Lock()
	defer ib.state.mu.Unlock()
	return append(ib.state.accounts[:0:0], ib.state.accounts...)
}

// IsPaperAccount checks if the accounts are paper accounts
func (ib *IB) IsPaperAccount() bool {
	return strings.HasPrefix(ib.ManagedAccounts()[0], "D")
}

// IsFinancialAdvisorAccount checks if the accounts is a financial advisor account.
func (ib *IB) IsFinancialAdvisorAccount() bool {
	return strings.HasPrefix(ib.ManagedAccounts()[0], "F")
}

// NextID returns a local next ID. It is initialised at connection.
// NextID = -1 if non initialised.
func (ib *IB) NextID() int64 {
	ib.state.mu.Lock()
	defer ib.state.mu.Unlock()
	currentID := ib.state.nextValidID
	ib.state.nextValidID++
	return currentID
}

// AccountValues returns a slice of account values for the given accounts.
//
// If no account is provided it will return values of all accounts.
// Account values need to be subscribed by ReqAccountUpdates. This is done at start up unless WithoutSync option is used.
func (ib *IB) AccountValues(account ...string) AccountValues {
	ib.state.mu.Lock()
	defer ib.state.mu.Unlock()
	var avs AccountValues
	for _, v := range ib.state.updateAccountValues {
		if len(account) == 0 || slices.Contains(account, v.Account) {
			avs = append(avs, v)
		}
	}
	return avs
}

// AccountSumary returns a slice of account values for the given accounts.
//
// If no account is provided it will return values of all accounts.
// On the first run it is calling ReqAccountSummary and is blocking, after it returns the last summary requested.
// To request a new summary call ReqAccountSummary.
func (ib *IB) AccountSummary(account ...string) AccountSummary {
	ib.state.mu.Lock()
	var as AccountSummary
	for _, v := range ib.state.accountSummary {
		if len(account) == 0 || slices.Contains(account, v.Account) {
			as = append(as, v)
		}
	}
	ib.state.mu.Unlock()
	if len(as) != 0 {
		return as
	}
	tags := []string{"AccountType", "NetLiquidation", "TotalCashValue", "SettledCash", "AccruedCash", "BuyingPower", "EquityWithLoanValue",
		"PreviousDayEquityWithLoanValue", "GrossPositionValue", "RegTEquity", "RegTMargin", "SMA", "InitMarginReq", "MaintMarginReq", "AvailableFunds",
		"ExcessLiquidity", "Cushion", "FullInitMarginReq", "FullMaintMarginReq", "FullAvailableFunds", "FullExcessLiquidity", "LookAheadNextChange",
		"LookAheadInitMarginReq", "LookAheadMaintMarginReq", "LookAheadAvailableFunds", "LookAheadExcessLiquidity", "HighestSeverity", "DayTradesRemaining",
		"DayTradesRemainingT+1", "DayTradesRemainingT+2", "DayTradesRemainingT+3", "DayTradesRemainingT+4", "Leverage,$LEDGER:ALL"}
	tagsString := strings.Join(tags, ",")
	ras, err := ib.ReqAccountSummary("All", tagsString)
	if err != nil {
		return nil
	}
	for _, v := range ras {
		if len(account) == 0 || slices.Contains(account, v.Account) {
			as = append(as, v)
		}
	}
	return as
}

// Portfolio returns a slice of portfolio item for the given accounts.
//
// If no account is provided it will return items of all accounts.
// Portfolios need to be subscribed by ReqAccountUpdates. This is done at start up unless WithoutSync option is used.
func (ib *IB) Portfolio(account ...string) []PortfolioItem {
	ib.state.mu.Lock()
	defer ib.state.mu.Unlock()
	var pis []PortfolioItem
	for acc, piMap := range ib.state.portfolio {
		if len(account) == 0 || slices.Contains(account, acc) {
			for _, pi := range piMap {
				pis = append(pis, pi)
			}
		}
	}
	return pis
}

// ReqPositions subscribes to real-time position stream for all accounts.
func (ib *IB) ReqPositions() {
	ib.eClient.ReqPositions()
}

// CancelPositions cancels real-time position subscription.
func (ib *IB) CancelPositions() {
	ib.eClient.CancelPositions()
}

// Position returns a slice of the last positions received for the given accounts.
//
// If no account is provided it will return positions of all accounts.
// Positions need to be subscribed with ReqPositions first.
func (ib *IB) Positions(account ...string) []Position {
	ib.state.mu.Lock()
	defer ib.state.mu.Unlock()
	var ps []Position
	for acc, pMap := range ib.state.positions {
		if len(account) == 0 || slices.Contains(account, acc) {
			for _, p := range pMap {
				ps = append(ps, p)
			}
		}
	}
	return ps
}

// PositionChan returns a channel that receives a continuous feed of Position updates.
//
// You need to subscribe to positions by calling ReqPositions.
// Do NOT close the channel.
func (ib *IB) PositionChan(account ...string) chan Position {
	ctx := ib.eClient.Ctx
	positionChan := make(chan Position)
	ch, unsubscribe := ib.pubSub.Subscribe("Position")
	var once sync.Once

	go func() {
		defer func() {
			unsubscribe()
			once.Do(func() { close(positionChan) })
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				var pos Position
				if err := Decode(&pos, msg); err != nil {
					return
				}
				if len(account) == 0 || slices.Contains(account, pos.Account) {
					positionChan <- pos
				}
			}
		}
	}()

	return positionChan
}

// ReqPnL requests and subscribe the PnL of assigned account.
func (ib *IB) ReqPnL(account string, modelCode string) {
	reqID := ib.NextID()
	key := Key(account, modelCode)

	ib.state.mu.Lock()
	_, ok := ib.state.pnlKey2ReqID[key]
	if ok {
		log.Warn().Str("account", account).Str("modelCode", modelCode).Msg("Pnl request already made")
		return
	}
	ib.state.pnlKey2ReqID[key] = reqID
	ib.state.reqID2Pnl[reqID] = &Pnl{Account: account, ModelCode: modelCode}
	ib.state.mu.Unlock()

	ib.eClient.ReqPnL(reqID, account, modelCode)
}

// CancelPnL cancels the PnL update of assigned account.
func (ib *IB) CancelPnL(account string, modelCode string) {
	ib.state.mu.Lock()
	reqID, ok := ib.state.pnlKey2ReqID[Key(account, modelCode)]
	if !ok {
		log.Warn().Str("account", account).Str("modelCode", modelCode).Msg("No pnl request to cancel")
		return
	}
	ib.state.mu.Unlock()
	ib.eClient.CancelPnL(reqID)
}

// Pnl returns a slice of Pnl values based on the specified account and model code.
//
// If account is an empty string, it returns Pnls for all accounts.
// If modelCode is an empty string, it returns Pnls for all model codes.
func (ib *IB) Pnl(account string, modelCode string) []Pnl {
	ib.state.mu.Lock()
	defer ib.state.mu.Unlock()
	var pnls []Pnl
	for _, pnl := range ib.state.reqID2Pnl {
		if (account == "" || account == pnl.Account) && (modelCode == "" || modelCode == pnl.ModelCode) {
			pnls = append(pnls, *pnl)
		}
	}
	return pnls
}

// PnlChan returns a channel that receives a continuous feed of Pnl updates.
//
// Do NOT close the channel.
func (ib *IB) PnlChan(account string, modelCode string) chan Pnl {
	ctx := ib.eClient.Ctx
	pnlChan := make(chan Pnl)
	ch, unsubscribe := ib.pubSub.Subscribe("Pnl")
	var once sync.Once

	go func() {
		defer func() {
			unsubscribe()
			once.Do(func() { close(pnlChan) })
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				var pnl Pnl
				if err := Decode(&pnl, msg); err != nil {
					return
				}
				if (account == "" || account == pnl.Account) && (modelCode == "" || modelCode == pnl.ModelCode) {
					pnlChan <- pnl
				}
			}
		}
	}()

	return pnlChan
}

// ReqPnLSingle requests and subscribe the single contract PnL of assigned account.
func (ib *IB) ReqPnLSingle(account string, modelCode string, contractID int64) {
	reqID := ib.NextID()
	key := Key(account, modelCode, contractID)
	ib.state.mu.Lock()
	_, ok := ib.state.pnlSingleKey2ReqID[key]
	if ok {
		log.Warn().Str("account", account).Str("modelCode", modelCode).Int64("contractID", contractID).Msg("Pnl single request already made")
		return
	}
	ib.state.pnlSingleKey2ReqID[key] = reqID
	ib.state.reqID2PnlSingle[reqID] = &PnlSingle{Account: account, ModelCode: modelCode, ConID: contractID}
	ib.state.mu.Unlock()
	ib.eClient.ReqPnLSingle(reqID, account, modelCode, contractID)
}

// CancelPnLSingle cancels the single contract PnL update of assigned account.
func (ib *IB) CancelPnLSingle(account string, modelCode string, contractID int64) {
	ib.state.mu.Lock()
	reqID, ok := ib.state.pnlSingleKey2ReqID[Key(account, modelCode, contractID)]
	if !ok {
		log.Warn().Str("account", account).Str("modelCode", modelCode).Int64("contractID", contractID).Msg("No pnl single request to cancel")
		return
	}
	ib.state.mu.Unlock()
	ib.eClient.CancelPnLSingle(reqID)
}

// PnlSingle returns a slice of PnlSingle values based on the specified account, model code, and contract ID.
//
// If account is an empty string, it returns PnlSingles for all accounts.
// If modelCode is an empty string, it returns PnlSingles for all model codes.
// If contractID is zero, it returns PnlSingles for all contracts.
func (ib *IB) PnlSingle(account string, modelCode string, contractID int64) []PnlSingle {
	ib.state.mu.Lock()
	defer ib.state.mu.Unlock()
	var pnlSingles []PnlSingle
	for _, pnlSingle := range ib.state.reqID2PnlSingle {
		if (account == "" || account == pnlSingle.Account) && (modelCode == "" || modelCode == pnlSingle.ModelCode) && (contractID == 0 || contractID == pnlSingle.ConID) {
			pnlSingles = append(pnlSingles, *pnlSingle)
		}
	}
	return pnlSingles
}

// PnlSingleChan returns a channel that receives a continuous feed of PnlSingle updates.
//
// Do NOT close the channel.
func (ib *IB) PnlSingleChan(account string, modelCode string, contractID int64) chan PnlSingle {
	ctx := ib.eClient.Ctx
	pnlSingleChan := make(chan PnlSingle)
	ch, unsubscribe := ib.pubSub.Subscribe("PnlSingle")
	var once sync.Once

	go func() {
		defer unsubscribe()
		defer func() { once.Do(func() { close(pnlSingleChan) }) }()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				var pnlSingle PnlSingle
				if err := Decode(&pnlSingle, msg); err != nil {
					return
				}
				if (account == "" || account == pnlSingle.Account) && (modelCode == "" || modelCode == pnlSingle.ModelCode) && (contractID == 0 || contractID == pnlSingle.ConID) {
					pnlSingleChan <- pnlSingle
				}
			}
		}
	}()

	return pnlSingleChan
}

// Trades returns a slice of all trades from this session
func (ib *IB) Trades() []*Trade {
	ib.state.mu.Lock()
	defer ib.state.mu.Unlock()
	var ts []*Trade
	for _, t := range ib.state.trades {
		ts = append(ts, t)
	}
	return ts
}

// OpenTrades returns a slice of copies of all open trades from this session
func (ib *IB) OpenTrades() []*Trade {
	ib.state.mu.Lock()
	defer ib.state.mu.Unlock()
	var ts []*Trade
	for _, t := range ib.state.trades {
		if !t.IsDone() {
			ts = append(ts, t)
		}
	}
	return ts
}

// Orders returns a slice of all orders from this session
func (ib *IB) Orders() []Order {
	ib.state.mu.Lock()
	defer ib.state.mu.Unlock()
	var ts []Order
	for _, t := range ib.state.trades {
		ts = append(ts, *t.Order)
	}
	return ts
}

// OpenOrders returns a slice of all open orders from this session
func (ib *IB) OpenOrders() []Order {
	ib.state.mu.Lock()
	defer ib.state.mu.Unlock()
	var ts []Order
	for _, t := range ib.state.trades {
		if !t.IsDone() {
			ts = append(ts, *t.Order)
		}
	}
	return ts
}

// Ticker returns a *Ticker for the provided and contract and a bool to tell if the ticker exists.
func (ib *IB) Ticker(contract *Contract) (*Ticker, bool) {
	ib.state.mu.Lock()
	defer ib.state.mu.Unlock()
	val, exists := ib.state.tickers[contract]
	return val, exists
}

// Tickers returns a slice of all Tickers
func (ib *IB) Tickers() []*Ticker {
	ib.state.mu.Lock()
	defer ib.state.mu.Unlock()
	var ts []*Ticker
	for _, t := range ib.state.tickers {
		ts = append(ts, t)
	}
	return ts
}

// NewTick returns the list of NewsTick
func (ib *IB) NewsTick() []NewsTick {
	ib.state.mu.Lock()
	defer ib.state.mu.Unlock()
	return append(ib.state.newsTicks[:0:0], ib.state.newsTicks...)
}

// ReqCurrentTime asks the current system time on the server side.
// A second call within a secund will not be answered.
func (ib *IB) ReqCurrentTime() (currentTime time.Time, err error) {
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	ch, unsubscribe := ib.pubSub.Subscribe("CurrentTime")
	defer unsubscribe()

	ib.eClient.ReqCurrentTime()
	select {
	case <-ctx.Done():
		return time.Time{}, ctx.Err()
	case msg := <-ch:
		if err = Decode(&currentTime, msg); err != nil {
			return time.Time{}, err
		}
		return currentTime, nil
	}
}

// ReqCurrentTime asks the current system time on the server side.
// A second call within a secund will not be answered.
func (ib *IB) ReqCurrentTimeInMillis() (int64, error) {
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	ch, unsubscribe := ib.pubSub.Subscribe("CurrentTimeInMillis")
	defer unsubscribe()

	ib.eClient.ReqCurrentTimeInMillis()

	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case msg := <-ch:
		var ctim int64
		if err := Decode(&ctim, msg); err != nil {
			return 0, err
		}
		return ctim, nil
	}
}

// ServerVersion returns the version of the TWS instance to which the API application is connected.
func (ib *IB) ServerVersion() int {
	return ib.eClient.ServerVersion()
}

// SetServerLogLevel sets the log level of the server.
// logLevel can be:
// 1 = SYSTEM
// 2 = ERROR	(default)
// 3 = WARNING
// 4 = INFORMATION
// 5 = DETAIL
func (ib *IB) SetServerLogLevel(logLevel int64) {
	ib.eClient.SetServerLogLevel(logLevel)
}

// ConnectionTime is the time the API application made a connection to TWS.
func (ib *IB) TWSConnectionTime() string {
	return ib.eClient.TWSConnectionTime()
}

// SetServerLogLevel setups the log level of server.
// The default detail level is ERROR. For more details, see API Logging.
func (ib *IB) SetLogLevel(logLevel int64) {
	ib.eClient.SetServerLogLevel(logLevel)
}

// ReqMktData request market data stream. It returns a *Ticker that will be updated.
//
// contract contains a description of the Contract for which market data is being requested.
// genericTickList is a commma delimited list of generic tick types. Tick types can be found in the Generic Tick Types page.
// Prefixing w/ 'mdoff' indicates that top mkt data shouldn't tick. You can specify the news source by postfixing w/ ':<source>. Example: "mdoff,292:FLY+BRF"
// For snapshots requests use Snapshot().
// mktDataOptions is for internal use only.Use default value XYZ.
func (ib *IB) ReqMktData(contract *Contract, genericTickList string, mktDataOptions ...TagValue) *Ticker {
	reqID := ib.NextID()

	ib.state.mu.Lock()
	ticker := ib.state.startTicker(reqID, contract, "mktData")
	ib.state.mu.Unlock()

	ib.eClient.ReqMktData(reqID, contract, genericTickList, false, false, mktDataOptions)

	return ticker
}

// CancelMktData stops the market data stream for the specified contract.
//
// Do not use CancelMktData for Snapshot() calls
func (ib *IB) CancelMktData(contract *Contract) {
	ib.state.mu.Lock()
	ticker := ib.state.tickers[contract]
	reqID, ok := ib.state.endTicker(ticker, "mktData")
	ib.state.mu.Unlock()

	if !ok {
		log.Error().Err(errUnknowReqID).Int64("conID", contract.ConID).Msg("<CancelMktData>")
	}

	ib.eClient.CancelMktData(reqID)
	log.Debug().Int64("reqID", reqID).Msg("<CancelMktData>")
}

// Snapshot return a market data snapshot.
//
// contract contains a description of the Contract for which market data is being requested.
// regulatorySnapshot: With the US Value Snapshot Bundle for stocks, regulatory snapshots are available for 0.01 USD each.
func (ib *IB) Snapshot(contract *Contract, regulatorySnapshot ...bool) (*Ticker, error) {
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	reqID := ib.NextID()

	ch, unsubscribe := ib.pubSub.Subscribe(reqID)
	defer unsubscribe()

	ib.state.mu.Lock()
	ticker := ib.state.startTicker(reqID, contract, "snapshot")
	ib.state.mu.Unlock()

	defer func() {
		ib.state.mu.Lock()
		ib.state.endTicker(ticker, "snapshot")
		ib.state.mu.Unlock()
	}()

	regulatory := false
	if len(regulatorySnapshot) > 0 {
		regulatory = regulatorySnapshot[0]
	}

	ib.eClient.ReqMktData(reqID, contract, "", true, regulatory, nil)

	var err error
	for {
		select {
		case <-ctx.Done():
			return ticker, ctx.Err()
		case msg := <-ch:
			if isErrorMsg(msg) {
				err = msg2Error(msg)
				if err != WarnDelayedMarketData && err != ErrPartlyNotSubsribed {
					return ticker, msg2Error(msg)
				}
				break
			}
			switch msg {
			case "TickSnapshotEnd":
				return ticker, err
			default:
				return ticker, errors.New(msg)
			}
		}
	}
}

// ReqMarketDataType changes the market data type.
//
// The API can receive frozen market data from Trader Workstation. Frozen market data is the last data recorded in our system.
// During normal trading hours, the API receives real-time market data.
// If you use this function, you are telling TWS to automatically switch to frozen market data after the close. Then, before the opening of the next
// trading day, market data will automatically switch back to real-time market data.
// marketDataType:
//
//	1 -> realtime streaming market data
//	2 -> frozen market data
//	3 -> delayed market data
//	4 -> delayed frozen market data
func (ib *IB) ReqMarketDataType(marketDataType int64) {
	log.Debug().Int64("marketDataType", marketDataType).Msg("<ReqMarketDataType>")
	ib.eClient.ReqMarketDataType(marketDataType)
}

// ReqSmartComponents requests the smartComponents.
//
// SmartComponents provide mapping from single letter codes to exchange names.
// Note: The exchanges must be open when using this request, otherwise an empty list is returned.
func (ib *IB) ReqSmartComponents(bboExchange string) ([]SmartComponent, error) {
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	reqID := ib.NextID()

	ch, unsubscribe := ib.pubSub.Subscribe(reqID)
	defer unsubscribe()

	ib.eClient.ReqSmartComponents(reqID, bboExchange)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case msg := <-ch:
		var scs []SmartComponent
		if isErrorMsg(msg) {
			return scs, msg2Error(msg)
		}
		if err := Decode(&scs, msg); err != nil {
			return nil, err
		}
		return scs, nil
	}
}

// ReqMarketRule requests the price increments rules.
//
// Note: market rule ids can be obtained by invoking reqContractDetails on a particular contract
func (ib *IB) ReqMarketRule(marketRuleID int64) ([]PriceIncrement, error) {
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	topic := Key("MarketRule", marketRuleID)

	ch, unsubscribe := ib.pubSub.Subscribe(topic)
	defer unsubscribe()

	ib.eClient.ReqMarketRule(marketRuleID)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case msg := <-ch:
		var pis []PriceIncrement
		if err := Decode(&pis, msg); err != nil {
			return nil, err
		}
		return pis, nil
	}
}

// ReqTickByTickData subscribe to tick-by-tick data stream and returns the *Ticker.
//
// contract is the *Contract you want subsciption for.
// tickType is one of "Last", "AllLast", "BidAsk" or "MidPoint".
// numberOfTicks is the number of ticks or 0 for unlimited.
// ignoreSize ignores bid/ask ticks that only update the size.
// No more than one request can be made for the same instrument within 15 seconds.
func (ib *IB) ReqTickByTickData(contract *Contract, tickType string, numberOfTicks int64, ignoreSize bool) *Ticker {
	reqID := ib.NextID()

	ib.state.mu.Lock()
	ticker := ib.state.startTicker(reqID, contract, tickType)
	ib.state.mu.Unlock()

	ib.eClient.ReqTickByTickData(reqID, contract, tickType, numberOfTicks, ignoreSize)

	return ticker
}

// CancelTickByTickData unsubscribes from tick-by-tick for given contract and tick type.
func (ib *IB) CancelTickByTickData(contract *Contract, tickType string) error {
	ib.state.mu.Lock()
	ticker := ib.state.tickers[contract]
	reqID, ok := ib.state.endTicker(ticker, tickType)
	ib.state.mu.Unlock()

	if !ok {
		log.Error().Err(errUnknowReqID).Int64("conID", contract.ConID).Msg("<CancelTickByTickData>")
		return errUnknowReqID
	}

	ib.eClient.CancelTickByTickData(reqID)
	log.Debug().Int64("reqID", reqID).Msg("<CancelTickByTickData>")
	return nil
}

// MidPoint requests and returns the last Mid Point
//
// No more than one request can be made for the same instrument within 15 seconds.
func (ib *IB) MidPoint(contract *Contract) (TickByTickMidPoint, error) {
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	reqID := ib.NextID()

	ch, unsubscribe := ib.pubSub.Subscribe(reqID)
	defer unsubscribe()

	ib.eClient.ReqTickByTickData(reqID, contract, "MidPoint", 0, true)
	defer ib.eClient.CancelTickByTickData(reqID)

	select {
	case <-ctx.Done():
		return TickByTickMidPoint{}, ctx.Err()
	case msg := <-ch:
		var tbtmp TickByTickMidPoint
		if isErrorMsg(msg) {
			return tbtmp, msg2Error(msg)
		}
		if err := Decode(&tbtmp, msg); err != nil {
			return TickByTickMidPoint{}, err
		}
		return tbtmp, nil
	}
}

// CalculateImpliedVolatility calculates the implied volatility given the option price.
func (ib *IB) CalculateImpliedVolatility(contract *Contract, optionPrice float64, underPrice float64, impVolOptions ...TagValue) (*TickOptionComputation, error) {
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	reqID := ib.NextID()

	ch, unsubscribe := ib.pubSub.Subscribe(reqID)
	defer unsubscribe()

	ib.eClient.CalculateImpliedVolatility(reqID, contract, optionPrice, underPrice, impVolOptions)

	select {
	case <-ctx.Done():
		ib.eClient.CancelCalculateImpliedVolatility(reqID)
		return nil, ctx.Err()
	case msg := <-ch:
		if isErrorMsg(msg) {
			return nil, msg2Error(msg)
		}
		items := Split(msg)
		switch items[0] {
		case "OptionComputation":
			var toc *TickOptionComputation
			if err := Decode(&toc, items[1]); err != nil {
				return toc, err
			}
			return toc, nil
		default:
			log.Error().Err(errUnknowItemType).Int64("reqID", reqID).Str("Type", items[0]).Msg("<CalculateImpliedVolatility>")
			return nil, errUnknowItemType
		}
	}
}

// CalculateOptionPrice calculates the price of the option given the volatility.
func (ib *IB) CalculateOptionPrice(contract *Contract, volatility float64, underPrice float64, optPrcOptions ...TagValue) (*TickOptionComputation, error) {
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	reqID := ib.NextID()

	ch, unsubscribe := ib.pubSub.Subscribe(reqID)
	defer unsubscribe()

	ib.eClient.CalculateOptionPrice(reqID, contract, volatility, underPrice, optPrcOptions)

	select {
	case <-ctx.Done():
		ib.eClient.CancelCalculateOptionPrice(reqID)
		return nil, ctx.Err()
	case msg := <-ch:
		if isErrorMsg(msg) {
			return nil, msg2Error(msg)
		}
		items := Split(msg)
		switch items[0] {
		case "OptionComputation":
			var toc *TickOptionComputation
			if err := Decode(&toc, items[1]); err != nil {
				return toc, err
			}
			return toc, nil
		default:
			log.Error().Err(errUnknowItemType).Int64("reqID", reqID).Str("Type", items[0]).Msg("<CalculateOptionPrice>")
			return nil, errUnknowItemType
		}
	}
}

// PlaceOrder places a new order or modify an existing order.
// It returns a *Trade that is kept live updated with status changes, fils, etc.
//
// contract is the *Contract to use for order.
// order contains the details of the order to be placed.
func (ib *IB) PlaceOrder(contract *Contract, order *Order) *Trade {
	if order.OrderID == 0 {
		order.OrderID = ib.NextID()
	}

	ib.eClient.PlaceOrder(order.OrderID, contract, order)

	key := orderKey(order.ClientID, order.OrderID, order.PermID)

	ib.state.mu.Lock()
	defer ib.state.mu.Unlock()

	trade, ok := ib.state.trades[key]
	if ok {
		// modification of an existing order
		if trade.IsDone() {
			panic("try to modify a done trade")
		}
		logEntry := TradeLogEntry{
			Time:    time.Now().UTC(),
			Status:  trade.OrderStatus.Status,
			Message: "Modify",
		}
		trade.addLog(logEntry)
		log.Debug().Int64("orderID", order.OrderID).Str("message", "modify order").Msg("<PlaceOrder>")
	} else {
		// new order
		order.ClientID = ib.config.ClientID
		trade = NewTrade(contract, order)
		key = orderKey(order.ClientID, order.OrderID, order.PermID) // clientID is updated
		ib.state.trades[key] = trade
		log.Debug().Int64("orderID", order.OrderID).Str("message", "open order").Msg("<PlaceOrder>")
	}
	return trade
}

// CancelOrder cancels the given order.
// orderCancel is an OrderCancel struct. You can pass NewOrderCancel()
func (ib *IB) CancelOrder(order *Order, orderCancel OrderCancel) {
	log.Debug().Int64("orderID", order.OrderID).Msg("<CancelOrder>")
	ib.eClient.CancelOrder(order.OrderID, orderCancel)
}

// ReqGlobalCancel cancels all open orders globally. It cancels both API and TWS open orders.
func (ib *IB) ReqGlobalCancel() {
	log.Debug().Msg("<ReqGlobalCancel>")
	ib.eClient.ReqGlobalCancel(NewOrderCancel())
}

// ReqOpenOrders requests the open orders that were placed from this client.
// The client with a clientId of 0 will also receive the TWS-owned open orders.
// These orders will be associated with the client and a new orderId will be generated.
// This association will persist over multiple API and TWS sessions.
func (ib *IB) ReqOpenOrders() {
	log.Debug().Msg("<ReqOpenOrders>")
	ib.eClient.ReqOpenOrders()
}

// ReqAutoOpenOrders requests that newly created TWS orders be implicitly associated with the client.
// This request can only be made from a client with clientId of 0.
// if autoBind is set to TRUE, newly created TWS orders will be implicitly associated with the client.
// If set to FALSE, no association will be made.
func (ib *IB) ReqAutoOpenOrders(autoBind bool) {
	log.Debug().Msg("<ReqAutoOpenOrders>")
	ib.eClient.ReqAutoOpenOrders(autoBind)
}

// ReqAllOpenOrders requests the open orders placed from all clients and also from TWS.
// Each open order will be fed back through the openOrder() and orderStatus() functions on the EWrapper.
// No association is made between the returned orders and the requesting client.
func (ib *IB) ReqAllOpenOrders() {
	log.Debug().Msg("<ReqAllOpenOrders>")
	ib.eClient.ReqAllOpenOrders()
}

// ReqAccountUpdates will start getting account values, portfolio, and last update time information.
func (ib *IB) ReqAccountUpdates(subscribe bool, accountName string) {
	ib.eClient.ReqAccountUpdates(subscribe, accountName)
}

// ReqAccountSummary requests and keep up to date the data that appears
// on the TWS Account Window Summary tab. The data is returned by
// accountSummary().
// This request is designed for an FA managed account but can be
// used for any multi-account structure.
// reqId is the ID of the data request. it Ensures that responses are matched
// to requests If several requests are in process.
// groupName sets to All to return account summary data for all
//
//	accounts, or set to a specific Advisor Account Group name that has
//	already been created in TWS Global Configuration.
//
// tags:str - A comma-separated list of account tags.  Available tags are:
//
//	accountountType
//	NetLiquidation,
//	TotalCashValue - Total cash including futures pnl
//	SettledCash - For cash accounts, this is the same as
//	TotalCashValue
//	AccruedCash - Net accrued interest
//	BuyingPower - The maximum amount of marginable US stocks the account can buy
//	EquityWithLoanValue - Cash + stocks + bonds + mutual funds
//	PreviousDayEquityWithLoanValue,
//	GrossPositionValue - The sum of the absolute value of all stock and equity option positions
//	RegTEquity,
//	RegTMargin,
//	SMA - Special Memorandum Account
//	InitMarginReq,
//	MaintMarginReq,
//	AvailableFunds,
//	ExcessLiquidity,
//	Cushion - Excess liquidity as a percentage of net liquidation value
//	FullInitMarginReq,
//	FullMaintMarginReq,
//	FullAvailableFunds,
//	FullExcessLiquidity,
//	LookAheadNextChange - Time when look-ahead values take effect
//	LookAheadInitMarginReq,
//	LookAheadMaintMarginReq,
//	LookAheadAvailableFunds,
//	LookAheadExcessLiquidity,
//	HighestSeverity - A measure of how close the account is to liquidation
//	DayTradesRemaining - The Number of Open/Close trades a user could put on before Pattern Day Trading is detected.
//		A value of "-1"	means that the user can put on unlimited day trades.
//	Leverage - GrossPositionValue / NetLiquidation
//	$LEDGER - Single flag to relay all cash balance tags*, only in base	currency.
//	$LEDGER:CURRENCY - Single flag to relay all cash balance tags*, only in	the specified currency.
//	$LEDGER:ALL - Single flag to relay all cash balance tags* in all currencies.
func (ib *IB) ReqAccountSummary(groupName string, tags string) (AccountSummary, error) {
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	reqID := ib.NextID()

	ch, unsubscribe := ib.pubSub.Subscribe(reqID)
	defer unsubscribe()

	ib.eClient.ReqAccountSummary(reqID, groupName, tags)
	var as AccountSummary
	for {
		select {
		case <-ctx.Done():
			ib.eClient.CancelAccountSummary(reqID)
			return as, ctx.Err()
		case msg := <-ch:
			switch msg {
			case "end":
				return as, nil
			default:
				var av AccountValue
				if err := Decode(&av, msg); err != nil {
					return as, err
				}
				as = append(as, av)
			}
		}
	}
}

// ReqPositionsMulti requests the positions for account and/or model.
// Results are delivered via EWrapper.positionMulti() and EWrapper.positionMultiEnd()
func (ib *IB) ReqPositionsMulti(reqID int64, account string, modelCode string) {
	ib.eClient.ReqPositionsMulti(reqID, account, modelCode)
	// TODO
}

// CancelPositionsMulti cancels the positions update of assigned account.
func (ib *IB) CancelPositionsMulti(reqID int64) {
	ib.eClient.CancelPositionsMulti(reqID)
	// TODO
}

// ReqAccountUpdatesMulti requests account updates for account and/or model.
func (ib *IB) ReqAccountUpdatesMulti(reqID int64, account string, modelCode string, ledgerAndNLV bool) {
	ib.eClient.ReqAccountUpdatesMulti(reqID, account, modelCode, ledgerAndNLV)
	// TODO
}

// CancelAccountUpdatesMulti cancels account update for reqID.
func (ib *IB) CancelAccountUpdatesMulti(reqID int64) {
	ib.eClient.CancelAccountUpdatesMulti(reqID)
	// TODO
}

// Executions returns a slice of all the executions from this session.
// To get executions from previous sessions of the day you must call reqExecutions.
func (ib *IB) Executions(execFilter ...*ExecutionFilter) []Execution {
	ib.state.mu.Lock()
	defer ib.state.mu.Unlock()
	var es []Execution
	for _, f := range ib.state.fills {
		if len(execFilter) == 0 {
			es = append(es, *f.Execution)
			continue
		}
		if f.Matches(execFilter[0]) {
			es = append(es, *f.Execution)
		}
	}
	return es
}

// ReqExecutions downloads the execution reports that meet the filter criteria to the client via the execDetails() function.
// To view executions beyond the past 24 hours, open the Trade Log in TWS and, while the Trade Log is displayed, request the executions again from the API.
// execFilter contains attributes that describe the filter criteria used to determine which execution reports are returned
// NOTE: Time format must be 'yyyymmdd-hh:mm:ss' Eg: '20030702-14:55'
func (ib *IB) ReqExecutions(execFilter ...*ExecutionFilter) ([]Execution, error) {
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	reqID := ib.NextID()

	ch, unsubscribe := ib.pubSub.Subscribe(reqID, 100)
	defer unsubscribe()

	ef := ibapi.NewExecutionFilter()
	if len(execFilter) > 0 {
		ef = execFilter[0]
	}

	ib.eClient.ReqExecutions(reqID, ef)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case msg := <-ch:
			if isErrorMsg(msg) {
				return nil, msg2Error(msg)
			}
			switch msg {
			case "end":
				return ib.Executions(execFilter...), nil
			default:
				/// Do nothing
			}
		}
	}
}

// Fills returns a slice of all the fills from this session.
// To get fills from previous sessions of the day you must call reqFills.
func (ib *IB) Fills(execFilter ...*ExecutionFilter) []Fill {
	ib.state.mu.Lock()
	defer ib.state.mu.Unlock()
	var fs []Fill
	for _, f := range ib.state.fills {
		if len(execFilter) == 0 {
			fs = append(fs, *f)
			continue
		}
		if f.Matches(execFilter[0]) {
			fs = append(fs, *f)
		}
	}
	return fs
}

func (ib *IB) ReqFills(execFilter ...*ExecutionFilter) ([]Fill, error) {
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	reqID := ib.NextID()

	ch, unsubscribe := ib.pubSub.Subscribe(reqID, 100)
	defer unsubscribe()

	ef := ibapi.NewExecutionFilter()
	if len(execFilter) > 0 {
		ef = execFilter[0]
	}

	ib.eClient.ReqExecutions(reqID, ef)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case msg := <-ch:
			if isErrorMsg(msg) {
				return nil, msg2Error(msg)
			}
			switch msg {
			case "end":
				return ib.Fills(execFilter...), nil
			default:
				/// Do nothing
			}
		}
	}
}

// ReqContractDetails downloads all details for a particular underlying.
// If the returned list is empty then the contract is not known.
// If the list has multiple values then the contract is ambiguous.
func (ib *IB) ReqContractDetails(contract *Contract) ([]ContractDetails, error) {
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	reqID := ib.NextID()

	ch, unsubscribe := ib.pubSub.Subscribe(reqID, 50)
	defer unsubscribe()

	ib.eClient.ReqContractDetails(reqID, contract)

	var cds []ContractDetails
	for {
		select {
		case <-ctx.Done():
			return cds, ctx.Err()
		case msg := <-ch:
			if isErrorMsg(msg) {
				return cds, msg2Error(msg)
			}
			switch msg {
			case "end":
				return cds, nil
			default:
				var cd ContractDetails
				if err := Decode(&cd, msg); err != nil {
					return cds, err
				}
				cds = append(cds, cd)
			}
		}
	}
}

// QualifyContract qualifies the given contracts by retrieving their details
// from the Interactive Brokers API. This method takes one or more contracts as input
// and checks each contract to determine if it is unambiguous.
//
// Parameters:
//   - contracts: One or more *Contract structs that need to be qualified.
//
// Returns:
//   - An error if any of the contracts cannot be qualified or if any are ambiguous (ErrAmbiguousContract).
//   - If successful, the contracts are updated in place with their corresponding details.
//
// QualifyContract fetches and qualifies contract details in parallel
func (ib *IB) QualifyContract(contracts ...*Contract) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(contracts)) // Channel for collecting errors
	doneChan := make(chan struct{})             // Channel to signal completion

	for _, contract := range contracts {
		wg.Add(1)
		go func(contract *Contract) {
			defer wg.Done()

			cds, err := ib.ReqContractDetails(contract)
			if err != nil {
				errChan <- err
				return
			}
			if len(cds) > 1 {
				log.Error().Err(ErrAmbiguousContract).Str("Symbol", contract.Symbol).Msg("QualifyContract")
				errChan <- ErrAmbiguousContract
				return
			}

			UpdateStruct(contract, cds[0].Contract)
		}(contract)
	}

	// Close the done channel once all goroutines finish
	go func() {
		wg.Wait()
		close(doneChan)
	}()

	// Wait for completion or an error
	select {
	case <-doneChan:
		close(errChan)
		// Check if any errors were reported
		if len(errChan) > 0 {
			return <-errChan // Return the first error
		}
	case err := <-errChan:
		return err // Return as soon as an error is encountered
	}

	return nil
}

// ReqMktDepthExchanges requests market depth exchanges.
func (ib *IB) ReqMktDepthExchanges() ([]DepthMktDataDescription, error) {
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	ch, unsubscribe := ib.pubSub.Subscribe("MktDepthExchanges")
	defer unsubscribe()

	ib.eClient.ReqMktDepthExchanges()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case msg := <-ch:
		var dmdds []DepthMktDataDescription
		if err := Decode(&dmdds, msg); err != nil {
			return nil, err
		}
		return dmdds, nil
	}
}

// ReqMktDepth requests the market depth for a specific contract.
//
// Requests the contract's market depth (order book). Note this request must be
// direct-routed to an exchange and not smart-routed. The number of simultaneous
// market depth requests allowed in an account is calculated based on a formula
// that looks at an accounts equity, commissions, and quote booster packs.
// contract contains a description of the contract for which market depth data is being requested.
// numRows specifies the numRowsumber of market depth rows to display.
// isSmartDepth	if true consolidates order book across exchanges.
// mktDepthOptions is for internal use only. Use default value XYZ.
func (ib *IB) ReqMktDepth(contract *Contract, numRows int, isSmartDepth bool, mktDepthOptions ...TagValue) (*Ticker, error) {
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	reqID := ib.NextID()

	ch, unsubscribe := ib.pubSub.Subscribe(reqID)
	defer unsubscribe()

	ib.state.mu.Lock()
	ticker := ib.state.startTicker(reqID, contract, "mktDepth")
	ib.state.mu.Unlock()

	ib.eClient.ReqMktDepth(reqID, contract, numRows, isSmartDepth, mktDepthOptions)

	cancelMktDepth := func() {
		ib.state.mu.Lock()
		ib.state.endTicker(ticker, "mktDepth")
		ib.state.mu.Unlock()
		ib.eClient.CancelMktDepth(reqID, isSmartDepth)
	}

	select {
	case <-ctx.Done():
		cancelMktDepth()
		return nil, ctx.Err()
	case msg := <-ch:
		if isErrorMsg(msg) {
			cancelMktDepth()
			return nil, msg2Error(msg)
		}
		return ticker, nil
	}

}

// CancelMktDepth cancels market depth updates.
func (ib *IB) CancelMktDepth(contract *Contract, isSmartDepth bool) error {
	ib.state.mu.Lock()
	ticker := ib.state.tickers[contract]
	reqID, ok := ib.state.endTicker(ticker, "mktDepth")
	ib.state.mu.Unlock()

	if !ok {
		log.Error().Err(errUnknowReqID).Int64("conID", contract.ConID).Msg("<CancelMktDepth>")
		return errUnknowReqID
	}

	ib.eClient.CancelMktDepth(reqID, isSmartDepth)
	log.Debug().Int64("reqID", reqID).Msg("<CancelMktDepth>")
	return nil
}

// ReqNewsBulletins requests and subscribes to news bulletins from the IB API.
//
// If allMsgs is set to true, it returns all existing bulletins for the current day as well as any new ones.
// If allMsgs is set to false, it only returns new bulletins.
// This method initiates the feed of news bulletins from the IB API.
func (ib *IB) ReqNewsBulletins(allMsgs bool) {
	ib.eClient.ReqNewsBulletins(allMsgs)
}

// CancelNewsBulletins cancels the subscription to news bulletins.
//
// It stops receiving updates for news bulletins that were previously requested with ReqNewsBulletins.
func (ib *IB) CancelNewsBulletins() {
	ib.eClient.CancelNewsBulletins()
}

// NewsBulletins returns a slice of all received NewsBulletin instances.
func (ib *IB) NewsBulletins() []NewsBulletin {
	ib.state.mu.Lock()
	defer ib.state.mu.Unlock()
	var nbs []NewsBulletin
	for _, nb := range ib.state.msgID2NewsBulletin {
		nbs = append(nbs, nb)
	}
	return nbs
}

// NewsBulletinsChan returns a channel that receives a continuous feed of NewsBulletin updates.
//
// Do not close the channel.
func (ib *IB) NewsBulletinsChan() chan NewsBulletin {
	ctx := ib.eClient.Ctx
	nbChan := make(chan NewsBulletin)
	ch, unsubscribe := ib.pubSub.Subscribe("NewsBulletin")
	var once sync.Once
	go func() {
		defer unsubscribe()
		defer func() { once.Do(func() { close(nbChan) }) }()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				var nb NewsBulletin
				if err := Decode(&nb, msg); err != nil {
					return
				}
				nbChan <- nb
			}
		}
	}()

	return nbChan
}

// RequestFA requests fa.
// The data returns in an XML string via wrapper.ReceiveFA().
// faData is 1->"GROUPS", 3->"ALIASES"
func (ib *IB) RequestFA(faDataType FaDataType) (cxml string, err error) {
	if !ib.IsFinancialAdvisorAccount() {
		return "", ErrNotFinancialAdvisor
	}
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	ch, unsubscribe := ib.pubSub.Subscribe("ReceiveFA")
	defer unsubscribe()

	ib.eClient.RequestFA(faDataType)

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case msg := <-ch:
		var rfa ReceiveFA
		if err := Decode(&rfa, msg); err != nil {
			return "", err
		}
		return rfa.Cxml, nil
	}
}

// ReplaceFA replaces the FA configuration information from the API.
// Note that this can also be done manually in TWS itself.
// faData specifies the type of Financial Advisor configuration data being requested.
// 1 = GROUPS
// 3 = ACCOUNT ALIASES
// cxml is the XML string containing the new FA configuration information.
func (ib *IB) ReplaceFA(faDataType FaDataType, cxml string) (string, error) {
	if !ib.IsFinancialAdvisorAccount() {
		return "", ErrNotFinancialAdvisor
	}
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	reqID := ib.NextID()

	ch, unsubscribe := ib.pubSub.Subscribe(reqID)
	defer unsubscribe()

	ib.eClient.ReplaceFA(reqID, faDataType, cxml)

	select {
	case <-ctx.Done():
		ib.eClient.CancelWshEventData(reqID)
		return "", ctx.Err()
	case text := <-ch:
		return text, nil
	}
}

// ReqHistoricalData requests historical data and returns a slice of []Bar.
//
// It requests historical data for a given contract by specifying an end time and date along with a duration string.
//
// Parameters:
//
// contract: A description of the contract for which historical data is being requested.
//
// endDateTime: Specifies the query end date and time in the format "yyyymmdd HH:mm:ss ttt", where "ttt" is an optional time zone.
// This value must be within the past six months. If no timezone is given then the TWS login timezone is used.
//
// durationStr: Sets the duration of the query. Valid values include any integer followed by a space and one of the following units. Examples:
//
//    "60 S", "30 D", "13 W", "6 M", "10 Y"...
//
// If no unit is specified, seconds are used by default.
//
// barSizeSetting: Specifies the size of the bars to be returned. Valid values are:
//
//     "1 secs", "5 secs", "10 secs", "15 secs", "30 secs"
//     "1 min", "2 mins", "3 mins", "5 mins", "10 mins", "15 mins", "20 mins", "30 mins"
//     "1 hour", "2 hours", "3 hours", "4 hours", "8 hours"
//     "1 day", "1 week", "1 month"
//
// whatToShow: Determines the type of data being requested. Valid values are:
//
//     "TRADES", "MIDPOINT", "BID", "ASK", "BID_ASK", "HISTORICAL_VOLATILITY",
//     "OPTION_IMPLIED_VOLATILITY", "REBATE_RATE", "FEE_RATE",
//     "YIELD_BID", "YIELD_ASK", "YIELD_BID_ASK", "YIELD_LAST".
//
//     For "SCHEDULE" use "ReqHistoricalSchedule()" instead.
//
// useRTH: Specifies whether to return all data available during the requested time span, or only data during regular trading hours (RTH).
// Valid values are:
//
//     false - All data is returned, including data outside regular trading hours.
//     true  - Only data within regular trading hours is returned, even if the requested time span falls outside of RTH.
//
// formatDate: Determines the date format used for the returned bars. Valid values are:
//
//     1 - Dates are formatted as "yyyymmdd HH:mm:ss".
//     2 - Dates are returned as a Unix timestamp (seconds since 1/1/1970 GMT).
//
// chartOptions: Reserved for internal use. Use the default value "XYZ".

func (ib *IB) ReqHistoricalData(contract *Contract, endDateTime string, duration string, barSize string, whatToShow string, useRTH bool, formatDate int, chartOptions ...TagValue) (chan Bar, CancelFunc) {
	barChan, cancel := ib.reqHistoricalData(contract, endDateTime, duration, barSize, whatToShow, useRTH, formatDate, false, chartOptions...)
	return barChan, cancel
}

// ReqHistoricalDataUpToDate requests historical data and returns a channel of `Bar` structs.
//
// This function continuously receives historical data updates for a given contract and returns a channel that streams `Bar` data updates.
//
// Parameters:
//   - contract: The contract for which historical data is being requested.
//   - duration: Defines the duration of the query. Examples of valid values include "60 S", "30 D", "13 W", "6 M", "10 Y". The unit must be specified (S for seconds, D for days, W for weeks, etc.).
//   - barSize: Specifies the size of the bars to be returned. Valid values include "1 sec", "5 secs", "10 secs", "15 secs", "30 secs", "1 min", "2 mins", "5 mins", etc.
//   - whatToShow: Determines the type of data being requested. Valid values include "TRADES", "BID", "ASK", "BID_ASK", "HISTORICAL_VOLATILITY", etc.
//   - useRTH: A boolean value to indicate if only data within regular trading hours (RTH) should be returned. `true` limits data to RTH, `false` includes all data.
//   - formatDate: Defines the date format for the returned bars. Valid values are `1` for the "yyyymmdd HH:mm:ss" format, or `2` for Unix timestamps.
//   - chartOptions: Reserved for internal use, generally set to the default value.
//
// Returns:
//
//	A `chan Bar` that streams incoming historical data. The caller must read from this channel until it is closed. The channel should be properly closed once the caller is done processing the data to prevent resource leaks.
//
// Do NOT close the channel.
func (ib *IB) ReqHistoricalDataUpToDate(contract *Contract, duration string, barSize string, whatToShow string, useRTH bool, formatDate int, chartOptions ...TagValue) (chan Bar, CancelFunc) {
	barChan, cancel := ib.reqHistoricalData(contract, "", duration, barSize, whatToShow, useRTH, formatDate, true, chartOptions...)
	return barChan, cancel
}

func (ib *IB) reqHistoricalData(contract *Contract, endDateTime string, duration string, barSize string, whatToShow string, useRTH bool, formatDate int, keepUpToDate bool, chartOptions ...TagValue) (chan Bar, CancelFunc) {
	ctx := ib.eClient.Ctx

	reqID := ib.NextID()

	ch, unsubscribe := ib.pubSub.Subscribe(reqID, 100)

	ib.eClient.ReqHistoricalData(reqID, contract, endDateTime, duration, barSize, whatToShow, useRTH, formatDate, keepUpToDate, chartOptions)

	barChan := make(chan Bar, 100)

	cancel := func() {
		ib.eClient.CancelHistoricalData(reqID)
		unsubscribe()
		// drains and safely closes the channel.
		for {
			select {
			case _, ok := <-barChan:
				if !ok {
					// Channel is already closed
					return
				}
				// Channel is not empty, drain it.
			default:
				// channel is empty, close it.
				close(barChan)
				return
			}
		}
	}

	go func() {
		defer cancel()
		for {
			select {
			case <-ctx.Done():
				log.Error().Err(ctx.Err()).Int64("reqID", reqID).Msg("<ReqHistoricalData>")
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				if isErrorMsg(msg) {
					log.Error().Err(msg2Error(msg)).Int64("reqID", reqID).Msg("<ReqHistoricalData>")
					return
				}
				items := Split(msg)
				switch items[0] {
				case "HistoricalDataEnd":
					if !keepUpToDate {
						log.Info().Str("symbol", contract.Symbol).Str("start date", items[1]).Str("end date", items[2]).Msg("<ReqHistoricalData> completed")
						return
					}
				case "HistoricalData", "HistoricalDataUpdate":
					var bar Bar
					if err := Decode(&bar, items[1]); err != nil {
						log.Error().Err(err).Int64("reqID", reqID).Msg("<ReqHistoricalData>")
						return
					}
					barChan <- bar
				default:
					log.Error().Err(errUnknowItemType).Int64("reqID", reqID).Str("Type", items[0]).Msg("<ReqHistoricalData>")
					return
				}
			}
		}
	}()

	return barChan, cancel
}

// ReqHistoricalSchedule requests historical schedule.
func (ib *IB) ReqHistoricalSchedule(contract *Contract, endDateTime string, duration string, useRTH bool) (HistoricalSchedule, error) {
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	reqID := ib.NextID()

	ch, unsubscribe := ib.pubSub.Subscribe(reqID)
	defer unsubscribe()

	ib.eClient.ReqHistoricalData(reqID, contract, endDateTime, duration, "1 day", "SCHEDULE", useRTH, 1, false, nil)

	select {
	case <-ctx.Done():
		ib.eClient.CancelHistoricalData(reqID)
		return HistoricalSchedule{}, ctx.Err()
	case msg := <-ch:
		var hs HistoricalSchedule
		if isErrorMsg(msg) {
			return hs, msg2Error(msg)
		}
		if err := Decode(&hs, msg); err != nil {
			return HistoricalSchedule{}, err
		}
		return hs, nil
	}
}

// ReqHeadTimeStamp retrieves the earliest available historical data timestamp for a contract.
//
// Parameters:
//   - contract: Contract specification containing symbol, type, expiry, etc.
//   - whatToShow: Type of data to retrieve (e.g., "TRADES", "MIDPOINT", "BID", "ASK")
//   - useRTH: When true, queries only Regular Trading Hours data
//   - formatDate: Determines the format of returned dates (1: utc, 2: local)
func (ib *IB) ReqHeadTimeStamp(contract *Contract, whatToShow string, useRTH bool, formatDate int) (time.Time, error) {
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	reqID := ib.NextID()

	ch, unsubscribe := ib.pubSub.Subscribe(reqID)
	defer unsubscribe()

	ib.eClient.ReqHeadTimeStamp(reqID, contract, whatToShow, useRTH, formatDate)

	select {
	case <-ctx.Done():
		ib.eClient.CancelHeadTimeStamp(reqID)
		return time.Time{}, ctx.Err()
	case msg := <-ch:
		if isErrorMsg(msg) {
			return time.Time{}, msg2Error(msg)
		}
		t, err := ParseIBTime(msg)
		if err != nil {
			return time.Time{}, err
		}
		return t, nil
	}
}

// ReqHistogramData requests histogram data.
// histograms return data as a function of price level
//
// contract: Contract to query.
// useRTH: If True then only show data from within Regular	Trading Hours, if False then show all data.
// period: Period of which data is being requested, for example "3 days".
func (ib *IB) ReqHistogramData(contract *Contract, useRTH bool, timePeriod string) ([]HistogramData, error) {
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	reqID := ib.NextID()

	ch, unsubscribe := ib.pubSub.Subscribe(reqID)
	defer unsubscribe()

	ib.eClient.ReqHistogramData(reqID, contract, useRTH, timePeriod)

	select {
	case <-ctx.Done():
		ib.eClient.CancelHistogramData(reqID)
		return nil, ctx.Err()
	case msg := <-ch:
		if isErrorMsg(msg) {
			return nil, msg2Error(msg)
		}
		var hds []HistogramData
		if err := Decode(&hds, msg); err != nil {
			return nil, err
		}
		return hds, nil
	}
}

// ReqHistoricalTicks requests historical ticks data.
// The time resolution of the ticks is one second.
//
// contract: Contract to query.
// startDateTime: the start time.
// endDateTime: One of “startDateTime“ or “endDateTime“ can	be given, the other must be blank.
// numberOfTicks: Number of ticks to request (1000 max). The actual result can contain a bit more to accommodate all ticks in the latest second.
// useRTH: If True then only show data from within Regular Trading Hours, if False then show all data.
// ignoreSize: Ignore bid/ask ticks that only update the size.
// miscOptions: Unknown.
func (ib *IB) ReqHistoricalTicks(contract *Contract, startDateTime, endDateTime time.Time, numberOfTicks int, useRTH bool, ignoreSize bool, miscOptions ...TagValue) ([]HistoricalTick, error, bool) {
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	reqID := ib.NextID()

	ch, unsubscribe := ib.pubSub.Subscribe(reqID)
	defer unsubscribe()

	ib.eClient.ReqHistoricalTicks(reqID, contract, FormatIBTimeUSEastern(startDateTime), FormatIBTimeUSEastern(endDateTime), numberOfTicks, "MIDPOINT", useRTH, ignoreSize, miscOptions)

	var ticks []HistoricalTick
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err(), false
		case msg := <-ch:
			if isErrorMsg(msg) {
				return nil, msg2Error(msg), false
			}
			items := Split(msg)
			if err := Decode(&ticks, items[0]); err != nil {
				return nil, err, false
			}
			done, err := strconv.ParseBool(items[1])
			if err != nil {
				log.Error().Err(err).Int64("reqID", reqID).Msg("<reqHistoricalTicks>")
			}
			return ticks, nil, done
		}
	}
}

// ReqHistoricalTickLast requests historical last ticks.
// The time resolution of the ticks is one second.
//
// contract: Contract to query.
// startDateTime: the start time.
// endDateTime: One of “startDateTime“ or “endDateTime“ can	be given, the other must be blank.
// numberOfTicks: Number of ticks to request (1000 max). The actual result can contain a bit more to accommodate all ticks in the latest second.
// useRTH: If True then only show data from within Regular Trading Hours, if False then show all data.
// ignoreSize: Ignore bid/ask ticks that only update the size.
// miscOptions: Unknown.
func (ib *IB) ReqHistoricalTickLast(contract *Contract, startDateTime, endDateTime time.Time, numberOfTicks int, useRTH bool, ignoreSize bool, miscOptions ...TagValue) ([]HistoricalTickLast, error, bool) {
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	reqID := ib.NextID()

	ch, unsubscribe := ib.pubSub.Subscribe(reqID)
	defer unsubscribe()

	ib.eClient.ReqHistoricalTicks(reqID, contract, FormatIBTimeUSEastern(startDateTime), FormatIBTimeUSEastern(endDateTime), numberOfTicks, "TRADES", useRTH, ignoreSize, miscOptions)

	var ticks []HistoricalTickLast
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err(), false
		case msg := <-ch:
			if isErrorMsg(msg) {
				return nil, msg2Error(msg), false
			}
			items := Split(msg)
			if err := Decode(&ticks, items[0]); err != nil {
				return nil, err, false
			}
			done, err := strconv.ParseBool(items[1])
			if err != nil {
				log.Error().Err(err).Int64("reqID", reqID).Msg("<ReqHistoricalTickLast>")
			}
			return ticks, nil, done
		}
	}
}

// ReqHistoricalTickBidAsk requests historical bid ask tick.
// The time resolution of the ticks is one second.
//
// contract: Contract to query.
// startDateTime: the start time.
// endDateTime: One of “startDateTime“ or “endDateTime“ can	be given, the other must be blank.
// numberOfTicks: Number of ticks to request (1000 max). The actual result can contain a bit more to accommodate all ticks in the latest second.
// useRTH: If True then only show data from within Regular Trading Hours, if False then show all data.
// ignoreSize: Ignore bid/ask ticks that only update the size.
// miscOptions: Unknown.
func (ib *IB) ReqHistoricalTickBidAsk(contract *Contract, startDateTime, endDateTime time.Time, numberOfTicks int, useRTH bool, ignoreSize bool, miscOptions ...TagValue) ([]HistoricalTickBidAsk, error, bool) {
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	reqID := ib.NextID()

	ch, unsubscribe := ib.pubSub.Subscribe(reqID)
	defer unsubscribe()

	ib.eClient.ReqHistoricalTicks(reqID, contract, FormatIBTimeUSEastern(startDateTime), FormatIBTimeUSEastern(endDateTime), numberOfTicks, "BID_ASK", useRTH, ignoreSize, miscOptions)

	var ticks []HistoricalTickBidAsk
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err(), false
		case msg := <-ch:
			if isErrorMsg(msg) {
				return nil, msg2Error(msg), false
			}
			items := Split(msg)
			if err := Decode(&ticks, items[0]); err != nil {
				return nil, err, false
			}
			done, err := strconv.ParseBool(items[1])
			if err != nil {
				log.Error().Err(err).Int64("reqID", reqID).Msg("<ReqHistoricalTickBidAsk>")
			}
			return ticks, nil, done
		}
	}
}

// ReqScannerParameters requests an XML string that describes all possible scanner queries.
func (ib *IB) ReqScannerParameters() (xml string, err error) {
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	ch, unsubscribe := ib.pubSub.Subscribe("ScannerParameters")
	defer unsubscribe()

	ib.eClient.ReqScannerParameters()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case xml := <-ch:
		return xml, nil
	}
}

type ScannerSubscriptionOptions struct {
	Options       []TagValue
	FilterOptions []TagValue
}

// ReqScannerSubscription subcribes a scanner that matched the subcription.
// scannerSubscription contains possible parameters used to filter results.
// scannerSubscriptionOptions, scannerSubscriptionOptions is for internal use only.Use default value XYZ.
func (ib *IB) ReqScannerSubscription(subscription *ScannerSubscription, scannerSubscriptionOptions ...ScannerSubscriptionOptions) ([]ScanData, error) {
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	reqID := ib.NextID()

	ch, unsubscribe := ib.pubSub.Subscribe(reqID, 50)
	defer unsubscribe()

	var options, filterOptions []TagValue
	if len(scannerSubscriptionOptions) > 0 {
		options = scannerSubscriptionOptions[0].Options
		filterOptions = scannerSubscriptionOptions[0].FilterOptions
	}

	ib.eClient.ReqScannerSubscription(reqID, subscription, options, filterOptions)

	var sds []ScanData
	for {
		select {
		case <-ctx.Done():
			ib.eClient.CancelScannerSubscription(reqID)
			return sds, ctx.Err()
		case msg := <-ch:
			if isErrorMsg(msg) {
				return sds, msg2Error(msg)
			}
			switch msg {
			case "end":
				return sds, nil
			default:
				var sd ScanData
				if err := Decode(&sd, msg); err != nil {
					return sds, err
				}
				sds = append(sds, sd)
			}
		}
	}
}

// ReqRealTimeBars requests realtime bars and returns a channel of Bar.
// reqId, the ticker ID, must be a unique value. When the data is received, it will be identified by this Id.
// This is also used when canceling the request.
// contract contains a description of the contract for which real time bars are being requested
// barSize, Currently only supports 5 second bars, if any other	value is used, an exception will be thrown.
// whatToShow determines the nature of the data extracted.
// Valid includes:
//
//	TRADES
//	BID
//	ASK
//	MIDPOINT
//
// useRTH sets regular Trading Hours only.
// Valid values include:
//
//	0 = all data available during the time span requested is returned,
//		including time intervals when the market in question was
//		outside of regular trading hours.
//	1 = only data within the regular trading hours for the product
//		requested is returned, even if the time time span falls
//		partially or completely outside.
//
// realTimeBarOptions is for internal use only. Use default value XYZ.
func (ib *IB) ReqRealTimeBars(contract *Contract, barSize int, whatToShow string, useRTH bool, realTimeBarsOptions ...TagValue) (chan RealTimeBar, CancelFunc) {
	ctx := ib.eClient.Ctx

	reqID := ib.NextID()

	ch, unsubscribe := ib.pubSub.Subscribe(reqID, 100)

	ib.eClient.ReqRealTimeBars(reqID, contract, barSize, whatToShow, useRTH, realTimeBarsOptions)

	rtBarChan := make(chan RealTimeBar, 100)

	cancel := func() {
		ib.eClient.CancelRealTimeBars(reqID)
		unsubscribe()
		// drains and safely closes the channel.
		for {
			select {
			case _, ok := <-rtBarChan:
				if !ok {
					// Channel is already closed
					return
				}
				// Channel is not empty, drain it.
			default:
				// channel is empty, close it.
				close(rtBarChan)
				return
			}
		}
	}

	go func() {
		defer cancel()
		for {
			select {
			case <-ctx.Done():
				log.Error().Err(ctx.Err()).Int64("reqID", reqID).Msg("<ReqRealTimeBars>")
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				if isErrorMsg(msg) {
					log.Error().Err(msg2Error(msg)).Int64("reqID", reqID).Msg("<ReqRealTimeBars>")
					return
				}
				var bar RealTimeBar
				if err := Decode(&bar, msg); err != nil {
					log.Error().Err(err).Int64("reqID", reqID).Msg("<ReqRealTimeBars>")
					return
				}
				rtBarChan <- bar
			}
		}
	}()

	return rtBarChan, cancel
}

// ReqFundamentalData requests fundamental data for stocks.
// The appropriate market data subscription must be set up in Account Management before you can receive this data.
// Result will be delivered via wrapper.FundamentalData().
// this func can handle conid specified in the Contract object,
// but not tradingClass or multiplier. This is because this func
// is used only for stocks and stocks do not have a multiplier and trading class.
// reqId is	the ID of the data request. Ensures that responses are matched to requests if several requests are in process.
// contract contains a description of the contract for which fundamental data is being requested.
// reportType is one of the following XML reports:
//
//	ReportSnapshot (company overview)
//	ReportsFinSummary (financial summary)
//	ReportRatios (financial ratios)
//	ReportsFinStatements (financial statements)
//	RESC (analyst estimates)
//	CalendarReport (company calendar)
func (ib *IB) ReqFundamentalData(contract *Contract, reportType string, fundamentalDataOptions ...TagValue) (data string, err error) {
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	reqID := ib.NextID()

	ch, unsubscribe := ib.pubSub.Subscribe(reqID)
	defer unsubscribe()

	ib.eClient.ReqFundamentalData(reqID, contract, reportType, fundamentalDataOptions)

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case msg := <-ch:
		if isErrorMsg(msg) {
			return "", msg2Error(msg)
		}
		return msg, nil
	}
}

// ReqNewsProviders requests a slice of news providers.
func (ib *IB) ReqNewsProviders() ([]NewsProvider, error) {
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	ch, unsubscribe := ib.pubSub.Subscribe("NewsProvider")
	defer unsubscribe()

	ib.eClient.ReqNewsProviders()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case msg := <-ch:
		var nps []NewsProvider
		if err := Decode(&nps, msg); err != nil {
			return nil, err
		}
		return nps, nil
	}
}

// ReqNewsArticle gets the body of a news article.
// providerCode: Code indicating news provider, like 'BZ' or 'FLY'.
// articleId: ID of the specific article. You can get it from ReqHistoricalNews
func (ib *IB) ReqNewsArticle(providerCode string, articleID string, newsArticleOptions ...TagValue) (*NewsArticle, error) {
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	reqID := ib.NextID()

	ch, unsubscribe := ib.pubSub.Subscribe(reqID)
	defer unsubscribe()

	ib.eClient.ReqNewsArticle(reqID, providerCode, articleID, newsArticleOptions)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case msg := <-ch:
		var na *NewsArticle
		if err := Decode(&na, msg); err != nil {
			return nil, err
		}
		return na, nil
	}
}

// ReqHistoricalNews requests historical news headlines.
//
// params:
// contractID: Search news articles for contract with this conId.
// providerCodes: A '+'-separated list of provider codes, like 'BZ+FLY'.
// startDateTime: The (exclusive) start of the date range.
// endDateTime: The (inclusive) end of the date range.
// totalResults: Maximum number of headlines to fetch (300 max).
// historicalNewsOptions: Unknown.
func (ib *IB) ReqHistoricalNews(contractID int64, providerCode string, startDateTime time.Time, endDateTime time.Time, totalResults int64, historicalNewsOptions ...TagValue) ([]HistoricalNews, error, bool) {
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	reqID := ib.NextID()

	ch, unsubscribe := ib.pubSub.Subscribe(reqID, 50)
	defer unsubscribe()

	ib.eClient.ReqHistoricalNews(reqID, contractID, providerCode, FormatIBTime(startDateTime), FormatIBTime(endDateTime), totalResults, historicalNewsOptions)

	var hns []HistoricalNews
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err(), false
		case msg := <-ch:
			if isErrorMsg(msg) {
				return nil, msg2Error(msg), false
			}
			items := Split(msg)
			switch items[0] {
			case "HistoricalNews": // "", "HistoricalNews"
				var hn HistoricalNews
				if err := Decode(&hn, items[1]); err != nil {
					return nil, err, false
				}
				hns = append(hns, hn)
			case "HistoricalNewsEnd":
				hasMore, err := strconv.ParseBool(items[1])
				if err != nil {
					return nil, err, false
				}
				return hns, nil, hasMore
			default:
				log.Error().Err(errUnknowItemType).Int64("reqID", reqID).Str("Type", items[0]).Msg("<ReqHistoricalNews>")
				return nil, errUnknowItemType, false
			}
		}
	}
}

// QueryDisplayGroups requests the display groups in TWS.
func (ib *IB) QueryDisplayGroups() (groups string, err error) {
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	reqID := ib.NextID()

	ch, unsubscribe := ib.pubSub.Subscribe(reqID)
	defer unsubscribe()

	ib.eClient.QueryDisplayGroups(reqID)

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case groups = <-ch:
		return groups, nil
	}
}

// UpdateDisplayGroup updates the display group in TWS.
// reqId is the requestId specified in subscribeToGroupEvents().
// contractInfo is the encoded value that uniquely represents the contract in IB.
// Possible values include:
//
//	none = empty selection
//	contractID@exchange - any non-combination contract.
//		Examples: 8314@SMART for IBM SMART; 8314@ARCA for IBM @ARCA.
//	combo = if any combo is selected.
func (ib *IB) UpdateDisplayGroup(reqID int64, contractInfo string) {
	ib.eClient.UpdateDisplayGroup(reqID, contractInfo)
}

// SubscribeToGroupEvents subcribes the group events.
// reqId is the unique number associated with the notification.
// groupId is the ID of the group, currently it is a number from 1 to 7.
func (ib *IB) SubscribeToGroupEvents(reqID int64, groupID int) {
	ib.eClient.SubscribeToGroupEvents(reqID, groupID)
}

// UnsubscribeFromGroupEvents unsubcribes the display group events.
func (ib *IB) UnsubscribeFromGroupEvents(reqID int64) {
	ib.eClient.UnsubscribeFromGroupEvents(reqID)
}

// // VerifyRequest is just for IB's internal use.
// // Allows to provide means of verification between the TWS and third party programs.
// func (ib *IB) VerifyRequest(apiName string, apiVersion string) {
// 	log.Debug().Str("apiName", apiName).Str("apiVersion", apiVersion).Msg("<VerifyRequest>")
// 	ib.eClient.VerifyRequest(apiName, apiVersion)
// }

// // VerifyMessage is just for IB's internal use.
// // Allows to provide means of verification between the TWS and third party programs.
// func (ib *IB) VerifyMessage(apiData string) {
// 	log.Debug().Str("apiData", apiData).Msg("<VerifyMessage>")
// 	ib.eClient.VerifyMessage(apiData)
// }

// // VerifyAndAuthRequest is just for IB's internal use.
// // Allows to provide means of verification between the TWS and third party programs.
// func (ib *IB) VerifyAndAuthRequest(apiName string, apiVersion string, opaqueIsvKey string) {
// 	log.Debug().Str("apiName", apiName).Str("apiVersion", apiVersion).Str("opaqueIsvKey", opaqueIsvKey).Msg("<VerifyAndAuthRequest>")
// 	ib.eClient.VerifyAndAuthRequest(apiName, apiVersion, opaqueIsvKey)
// }

// // VerifyAndAuthMessage is just for IB's internal use.
// // Allows to provide means of verification between the TWS and third party programs.
// func (ib *IB) VerifyAndAuthMessage(apiData string, xyzResponse string) {
// 	log.Debug().Str("apiData", apiData).Str("xyzResponse", xyzResponse).Msg("<VerifyAndAuthMessage>")
// 	ib.eClient.VerifyAndAuthMessage(apiData, xyzResponse)
// }

// ReqSecDefOptParams requests security definition option parameters.
// futFopExchange is the exchange on which the returned options are trading. Can be set to the empty string "" for all exchanges.
// underlyingSecType is the type of the underlying security, i.e. STK.
// underlyingConId is the contract ID of the underlying security.
func (ib *IB) ReqSecDefOptParams(underlyingSymbol string, futFopExchange string, underlyingSecurityType string, underlyingContractID int64) ([]OptionChain, error) {
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	reqID := ib.NextID()

	ch, unsubscribe := ib.pubSub.Subscribe(reqID, 50)
	defer unsubscribe()

	ib.eClient.ReqSecDefOptParams(reqID, underlyingSymbol, futFopExchange, underlyingSecurityType, underlyingContractID)

	var ocs []OptionChain
	for {
		select {
		case <-ctx.Done():
			return ocs, ctx.Err()
		case msg := <-ch:
			switch msg {
			case "end":
				return ocs, nil
			default:
				var oc OptionChain
				if err := Decode(&oc, msg); err != nil {
					return ocs, err
				}
				ocs = append(ocs, oc)
			}
		}
	}
}

// ReqSoftDollarTiers requests pre-defined Soft Dollar Tiers.
// This is only supported for registered professional advisors and hedge and mutual funds
// who have configured Soft Dollar Tiers in Account Management.
func (ib *IB) ReqSoftDollarTiers() ([]SoftDollarTier, error) {
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	reqID := ib.NextID()

	ch, unsubscribe := ib.pubSub.Subscribe(reqID)
	defer unsubscribe()

	ib.eClient.ReqSoftDollarTiers(reqID)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case msg := <-ch:
		var sdt []SoftDollarTier
		if err := Decode(&sdt, msg); err != nil {
			return nil, err
		}
		return sdt, nil
	}
}

// ReqFamilyCodes requests family codes.
func (ib *IB) ReqFamilyCodes() ([]FamilyCode, error) {
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	ch, unsubscribe := ib.pubSub.Subscribe("FamilyCodes")
	defer unsubscribe()

	ib.eClient.ReqFamilyCodes()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case msg := <-ch:
		var fcs []FamilyCode
		if err := Decode(&fcs, msg); err != nil {
			return nil, err
		}
		return fcs, nil
	}
}

// ReqMatchingSymbols requests matching symbols.
func (ib *IB) ReqMatchingSymbols(pattern string) ([]ContractDescription, error) {
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	reqID := ib.NextID()

	ch, unsubscribe := ib.pubSub.Subscribe(reqID)
	defer unsubscribe()

	ib.eClient.ReqMatchingSymbols(reqID, pattern)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case msg := <-ch:
		var cds []ContractDescription
		if err := Decode(&cds, msg); err != nil {
			return nil, err
		}
		return cds, nil
	}
}

// ReqCompletedOrders requests the completed orders
// If apiOnly parameter is true, then only completed orders placed from API are requested.
func (ib *IB) ReqCompletedOrders(apiOnly bool) {
	log.Debug().Bool("apiOnly", apiOnly).Msg("<ReqCompletedOrders>")
	ib.eClient.ReqCompletedOrders(apiOnly)
}

// ReqWshMetaData requests Wall Street Horizon Meta data
func (ib *IB) ReqWshMetaData() (dataJson string, err error) {
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	reqID := ib.NextID()

	ch, unsubscribe := ib.pubSub.Subscribe(reqID)
	defer unsubscribe()

	ib.eClient.ReqWshMetaData(reqID)

	select {
	case <-ctx.Done():
		ib.eClient.CancelWshMetaData(reqID)
		return "", ctx.Err()
	case msg := <-ch:
		if isErrorMsg(msg) {
			return "", msg2Error(msg)
		}
		return msg, nil
	}
}

// ReqWshEventData requests Wall Street Horizon event data.
func (ib *IB) ReqWshEventData(wshEventData WshEventData) (dataJson string, err error) {
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	reqID := ib.NextID()

	ch, unsubscribe := ib.pubSub.Subscribe(reqID)
	defer unsubscribe()

	ib.eClient.ReqWshEventData(reqID, wshEventData)

	select {
	case <-ctx.Done():
		ib.eClient.CancelWshEventData(reqID)
		return "", ctx.Err()
	case msg := <-ch:
		if isErrorMsg(msg) {
			return "", msg2Error(msg)
		}
		return msg, nil
	}
}

// ReqUserInfo returns user white branding info with timeout and error handling.
func (ib *IB) ReqUserInfo() (whiteBrandingId string, err error) {
	ctx, cancel := context.WithTimeout(ib.eClient.Ctx, ib.config.Timeout)
	defer cancel()

	reqID := ib.NextID()

	ch, unsubscribe := ib.pubSub.Subscribe(reqID)
	defer unsubscribe()

	ib.eClient.ReqUserInfo(reqID)

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case whiteBrandingId = <-ch:
		return whiteBrandingId, nil
	}
}
