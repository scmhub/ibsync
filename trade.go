package ibsync

import (
	"fmt"
	"sync"
	"time"
)

// Status represents the current state of an order in the trading system.
type Status string

// Order status constants define all possible states an order can be in.
const (
	PendingSubmit Status = "PendingSubmit" // indicates that you have transmitted the order, but have not yet received confirmation that it has been accepted by the order destination.
	PendingCancel Status = "PendingCancel" // PendingCancel	indicates that you have sent a request to cancel the order but have not yet received cancel confirmation from the order destination. At this point, your order is not confirmed canceled. It is not guaranteed that the cancellation will be successful.
	PreSubmitted  Status = "PreSubmitted"  // indicates that a simulated order type has been accepted by the IB system and that this order has yet to be elected. The order is held in the IB system until the election criteria are met. At that time the order is transmitted to the order destination as specified.
	Submitted     Status = "Submitted"     // indicates that your order has been accepted by the system.
	ApiPending    Status = "ApiPending"    // Order is pending processing by the API
	ApiCancelled  Status = "ApiCancelled"  // after an order has been submitted and before it has been acknowledged, an API client client can request its cancelation, producing this state.
	Cancelled     Status = "Cancelled"     // indicates that the balance of your order has been confirmed canceled by the IB system. This could occur unexpectedly when IB or the destination has rejected your order.
	Filled        Status = "Filled"        // 	indicates that the order has been completely filled. Market orders executions will not always trigger a Filled status.
	Inactive      Status = "Inactive"      // indicates that the order was received by the system but is no longer active because it was rejected or canceled.
)

// IsActive returns true if the status indicates the order is still active in the market.
func (s Status) IsActive() bool {
	switch s {
	case PendingSubmit, ApiPending, PreSubmitted, Submitted:
		return true
	default:
		return false
	}
}

// IsDone returns true if the status indicates the order has reached a terminal state.
func (s Status) IsDone() bool {
	switch s {
	case Filled, Cancelled, ApiCancelled:
		return true
	default:
		return false
	}
}

// OrderStatus represents the current state and details of an order.
type OrderStatus struct {
	OrderID       int64   // Unique identifier for the order
	Status        Status  // Current status of the order
	Filled        Decimal // Amount of order that has been filled
	Remaining     Decimal // Amount of order remaining to be filled
	AvgFillPrice  float64 // Average price of filled portions
	PermID        int64   // Permanent ID assigned by IB
	ParentID      int64   // ID of parent order if this is a child order
	LastFillPrice float64 // Price of the last fill
	ClientID      int64   // Client identifier
	WhyHeld       string  // Reason why order is being held
	MktCapPrice   float64 // Market cap price
}

// IsActive returns true if the order status indicates an active order.
func (os OrderStatus) IsActive() bool {
	return os.Status.IsActive()
}

// IsDone returns true if the order status indicates the order has reached a terminal state.
func (os OrderStatus) IsDone() bool {
	return os.Status.IsDone()
}

// Fill represents a single execution fill of an order, including contract details,
// execution information, and commission data.
type Fill struct {
	Contract                *Contract               // Contract details for the filled order
	Execution               *Execution              // Execution details of the fill
	CommissionAndFeesReport CommissionAndFeesReport // Commission and fees information for the fill
	Time                    time.Time               // Timestamp of the fill
}

// PassesExecutionFilter checks if the fill matches the specified execution filter criteria.
func (f *Fill) Matches(filter *ExecutionFilter) bool {
	if f == nil {
		return false
	}
	if filter.AcctCode != "" && filter.AcctCode != f.Execution.AcctNumber {
		return false
	}
	if filter.ClientID != 0 && filter.ClientID != f.Execution.ClientID {
		return false
	}
	if filter.Exchange != "" && filter.Exchange != f.Execution.Exchange {
		return false
	}
	if filter.SecType != "" && filter.SecType != f.Contract.SecType {
		return false
	}
	if filter.Side != "" && filter.Side != f.Execution.Side {
		return false
	}
	if filter.Symbol != "" && filter.Symbol != f.Contract.Symbol {
		return false
	}
	if filter.Time != "" {
		filterTime, err := ParseIBTime(filter.Time)
		if err != nil {
			log.Error().Err(err).Msg("PassesExecutionFilter")
			return false
		}
		if f.Time.Before(filterTime) {
			return false
		}
	}
	return true
}

// TradeLogEntry represents a single entry in the trade's log, recording status changes
// and other significant events.
type TradeLogEntry struct {
	Time      time.Time // Timestamp of the log entry
	Status    Status    // Status at the time of the log entry
	Message   string    // Descriptive message about the event
	ErrorCode int64     // Error code if applicable
}

// Trade represents a complete trading operation, including the contract, order details,
// current status, and execution fills.
type Trade struct {
	Contract    *Contract
	Order       *Order
	OrderStatus OrderStatus
	mu          sync.RWMutex
	fills       []*Fill
	logs        []TradeLogEntry
	done        chan struct{}
}

/* func (t* Trade) Equal(other Trade) bool{
	t.Order
} */

// NewTrade creates a new Trade instance with the specified contract and order details.
// Optional initial order status can be provided.
func NewTrade(contract *Contract, order *Order, orderStatus ...OrderStatus) *Trade {
	var os OrderStatus
	if len(orderStatus) > 0 {
		os = orderStatus[0]
	} else {
		os = OrderStatus{
			OrderID: order.OrderID,
			Status:  PendingSubmit,
		}
	}
	return &Trade{
		Contract:    contract,
		Order:       order,
		OrderStatus: os,
		fills:       make([]*Fill, 0),
		logs:        []TradeLogEntry{{Time: time.Now().UTC(), Status: PendingSubmit}},
		done:        make(chan struct{}),
	}
}

// IsActive returns true if the trade is currently active in the market.
func (t *Trade) IsActive() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.isActive()
}

// isActive is an internal helper that checks if the trade is active without locking.
func (t *Trade) isActive() bool {
	return t.OrderStatus.IsActive()
}

// IsDone returns true if the trade has reached a terminal state.
func (t *Trade) IsDone() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.isDone()
}

// isDone is an internal helper that checks if the trade is done without locking.
func (t *Trade) isDone() bool {
	return t.OrderStatus.IsDone()
}

// Done returns a channel that will be closed when the trade reaches a terminal state.
func (t *Trade) Done() <-chan struct{} {
	return t.done
}

// markDone closes the done channel to signal trade completion.
// This is an internal method and should be called with appropriate locking.
func (t *Trade) markDone() {
	// Ensure that the done channel is closed only once
	select {
	case <-t.done:
		// Channel already closed
	default:
		close(t.done)
	}
}

// markDoneSafe safely marks the trade as done with proper locking.
func (t *Trade) markDoneSafe() {
	t.mu.RLock()
	defer t.mu.RUnlock()
	t.markDone()
}

// Fills returns a copy of all fills for this trade
func (t *Trade) Fills() []*Fill {
	t.mu.RLock()
	defer t.mu.RUnlock()

	fills := make([]*Fill, len(t.fills))
	copy(fills, t.fills)
	return fills
}

// addFill adds a new fill to the trade's fill history
func (t *Trade) addFill(fill *Fill) {
	t.fills = append(t.fills, fill)
}

// Logs returns a copy of all log entries for this trade
func (t *Trade) Logs() []TradeLogEntry {
	t.mu.RLock()
	defer t.mu.RUnlock()

	logs := make([]TradeLogEntry, len(t.logs))
	copy(logs, t.logs)
	return logs
}

// addLog adds a new log entry to the trade's history
func (t *Trade) addLog(tradeLogEntry TradeLogEntry) {
	t.logs = append(t.logs, tradeLogEntry)
}

func (t *Trade) addLogSafe(tradeLogEntry TradeLogEntry) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	t.addLog(tradeLogEntry)
}

func (t *Trade) Equal(other *Trade) bool {
	return t.Order.HasSameID(other.Order)
}

func (t *Trade) String() string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return fmt.Sprintf("Trade{Contract: %v, Order: %v, Status: %v, Fills: %d}",
		t.Contract, t.Order, t.OrderStatus, len(t.fills))
}
