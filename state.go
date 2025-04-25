package ibsync

import (
	"sync"
	"time"
)

// ibState holds the data to keep in sync with IB server.
// Note: It is the responsibility of the user to lock and unlock this state!
type ibState struct {
	mu                  sync.Mutex
	accounts            []string
	nextValidID         int64
	updateAccountTime   time.Time
	updateAccountValues map[string]AccountValue            //  Key(account, tag, currency, modelCode) -> AccountValue
	accountSummary      map[string]AccountValue            // Key(account, tag, currency) -> AccountValue
	portfolio           map[string]map[int64]PortfolioItem // account -> conId -> PortfolioItem
	positions           map[string]map[int64]Position      // account -> conId -> Position
	trades              map[string]*Trade                  // permId -> Trade
	permID2Trade        map[int64]*Trade                   // permId -> Trade
	fills               map[string]*Fill                   // execID -> Fill
	msgID2NewsBulletin  map[int64]NewsBulletin             // msgID -> NewsBulletin
	tickers             map[*Contract]*Ticker              // *Contract -> Ticker
	reqID2Ticker        map[int64]*Ticker                  // reqId -> Ticker
	ticker2ReqID        map[string]map[*Ticker]int64       // Ticker -> reqId
	reqID2Pnl           map[int64]*Pnl                     // reqId -> Pnl
	reqID2PnlSingle     map[int64]*PnlSingle               // reqId -> PnlSingle
	pnlKey2ReqID        map[string]int64                   // Key(account, modelCode) -> reqID
	pnlSingleKey2ReqID  map[string]int64                   // Key(account, modelCode, conID) -> reqID
	newsTicks           []NewsTick
}

// NewState creates and initializes a new ibState instance.
func NewState() *ibState {
	s := &ibState{}
	s.reset()
	return s
}

// reset reinitializes all maps and slices in the state to their zero values.
func (s *ibState) reset() {
	s.accounts = nil
	s.nextValidID = -1
	s.updateAccountTime = time.Time{}
	s.updateAccountValues = make(map[string]AccountValue)
	s.accountSummary = make(map[string]AccountValue)
	s.portfolio = make(map[string]map[int64]PortfolioItem)
	s.positions = make(map[string]map[int64]Position)
	s.trades = make(map[string]*Trade)
	s.permID2Trade = make(map[int64]*Trade)
	s.fills = make(map[string]*Fill)
	s.msgID2NewsBulletin = make(map[int64]NewsBulletin)
	s.tickers = make(map[*Contract]*Ticker)
	s.reqID2Ticker = make(map[int64]*Ticker)
	s.ticker2ReqID = make(map[string]map[*Ticker]int64)
	s.reqID2Pnl = make(map[int64]*Pnl)
	s.reqID2PnlSingle = make(map[int64]*PnlSingle)
	s.pnlKey2ReqID = make(map[string]int64)
	s.pnlSingleKey2ReqID = make(map[string]int64)
	s.newsTicks = nil
}

// startTicker registers a new ticker with the state for a specific request ID and contract.
func (s *ibState) startTicker(reqID int64, contract *Contract, tickerType string) *Ticker {
	ticker := NewTicker(contract)
	s.tickers[contract] = ticker
	s.reqID2Ticker[reqID] = ticker
	_, ok := s.ticker2ReqID[tickerType]
	if !ok {
		s.ticker2ReqID[tickerType] = make(map[*Ticker]int64)
	}
	s.ticker2ReqID[tickerType][ticker] = reqID
	return ticker
}

// endTicker removes a ticker from the state for a specific ticker type.
func (s *ibState) endTicker(ticker *Ticker, tickerType string) (int64, bool) {
	reqID, ok := s.ticker2ReqID[tickerType][ticker]
	if !ok {
		return 0, false
	}
	delete(s.ticker2ReqID[tickerType], ticker)
	return reqID, true
}

// updateID updates the next requested ID to be at least the specified minimum ID.
func (s *ibState) updateID(minID int64) {
	s.nextValidID = max(s.nextValidID, minID)
}
