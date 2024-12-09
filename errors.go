package ibsync

import (
	"errors"
	"slices"

	"github.com/scmhub/ibapi"
)

// Exported Errors
var (
	ErrNotPaperTrading     = errors.New("this account is not a paper trading account")
	ErrNotFinancialAdvisor = errors.New("this account is not a financial advisor account")
	ErrNoConfigProvided    = errors.New("no config provided")
	ErrAmbiguousContract   = errors.New("ambiguous contract")
	ErrNoDataSubscription  = errors.New("no data subscription")
)

// Internal Errors
var (
	errUnknowReqID     = errors.New("unknown reqID")
	errUnknowOrder     = errors.New("unknown order")
	errUnknowExecution = errors.New("unknown execution")
	errUnknowItemType  = errors.New("unknown item type")
	errUnknownTickType = errors.New("unknown tick type")
)

// TWS Errors
// https://www.interactivebrokers.eu/campus/ibkr-api-page/tws-api-error-codes/

// warnigCodes are errors codes received from TWS that should be treated as warnings
var warningCodes = []int64{ //
	161,   // Cancel attempted when order is not in a cancellable state. Order permId = // An attempt was made to cancel an order not active at the time.
	202,   // Order cancelled – Reason:	An active order on the IB server was cancelled. // See Order Placement Considerations for additional information/considerations for these errors.
	2104,  // Market data farm connection is OK.
	2106,  // A historical data farm is connected.
	2107,  // HMDS data farm connection is inactive but should be available upon demand.
	2108,  // A market data farm connection has become inactive but should be available upon demand.
	2119,  // Market data farm is connecting.
	2158,  // Sec-def data farm connection is OK.
	10167, // Requested market data is not subscribed. Displaying delayed market data.
	10197, // No market data during competing live session.
}

func IsWarning(cmp ibapi.CodeMsgPair) bool {
	return slices.Contains(warningCodes, cmp.Code)
}

var (
	// Errors
	ErrMaxNbTickerReached             = ibapi.CodeMsgPair{Code: 101, Msg: "Max number of tickers has been reached."}
	ErrMissingReportType              = ibapi.CodeMsgPair{Code: 430, Msg: "The fundamentals data for the security specified is not available."}
	ErrNewsFeedNotAllowed             = ibapi.CodeMsgPair{Code: 10276, Msg: "News feed is not allowed."}
	ErrAdditionalSubscriptionRequired = ibapi.CodeMsgPair{Code: 10089, Msg: "Requested market data requires additional subscription for API."}
	ErrPartlyNotSubsribed             = ibapi.CodeMsgPair{Code: 10090, Msg: "Part of requested market data is not subscribed."}

	// Warnings
	WarnDelayedMarketData    = ibapi.CodeMsgPair{Code: 10167, Msg: "Requested market data is not subscribed. Displaying delayed market data."}
	WarnCompetingLiveSession = ibapi.CodeMsgPair{Code: 10197, Msg: "No market data during competing live session."}
)

// normaliseCodeMsgPair nomalise IB errors
// IB errors can have different error messages for a given code.
// We get rid of the original message in order to have consitent errors.
// Original message can be seen in the logs.
func normaliseCodeMsgPair(cmp ibapi.CodeMsgPair) error {
	switch cmp.Code {
	case 2104, 2106, 2107, 2108, 21019:
		return cmp
	case 10167:
		return WarnDelayedMarketData
	case 10197:
		return WarnCompetingLiveSession
	case 430:
		return ErrMissingReportType
	case 10089:
		return ErrAdditionalSubscriptionRequired
	case 10090:
		return ErrPartlyNotSubsribed
	default:
		return cmp
	}
}
