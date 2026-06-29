package ibsync

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// OrderStatusData represents the current state and details of an order.
type OrderStatusData struct {
	OrderID       int64       // Unique identifier for the order
	Status        OrderStatus // Current status of the order
	Filled        Decimal     // Amount of order that has been filled
	Remaining     Decimal     // Amount of order remaining to be filled
	AvgFillPrice  float64     // Average price of filled portions
	PermID        int64       // Permanent ID assigned by IB
	ParentID      int64       // ID of parent order if this is a child order
	LastFillPrice float64     // Price of the last fill
	ClientID      int64       // Client identifier
	WhyHeld       string      // Reason why order is being held
	MktCapPrice   float64     // Market cap price
}

// IsActive returns true if the order status indicates an active order.
func (os OrderStatusData) IsActive() bool {
	return os.Status.IsActive()
}

// IsDone returns true if the order status indicates the order has reached a terminal state.
func (os OrderStatusData) IsDone() bool {
	return os.Status.IsTerminal()
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

func (f *Fill) String() string {
	return fmt.Sprintf("Fill{Contract: %v, Execution: %v, CommissionAndFeesReport: %v, Time: %v}",
		f.Contract, f.Execution, f.CommissionAndFeesReport, f.Time)
}

// TradeLogEntry represents a single entry in the trade's log, recording status changes
// and other significant events.
type TradeLogEntry struct {
	Time      time.Time   // Timestamp of the log entry
	Status    OrderStatus // Status at the time of the log entry
	Message   string      // Descriptive message about the event
	ErrorCode int64       // Error code if applicable
}

// Trade represents a complete trading operation, including the contract, order details,
// current status, and execution fills.
type Trade struct {
	Contract    *Contract
	Order       *Order
	OrderStatus OrderStatusData
	mu          sync.RWMutex
	fills       []*Fill
	logs        []TradeLogEntry
	done        chan struct{}
	ack         chan struct{}
}

/* func (t* Trade) Equal(other Trade) bool{
	t.Order
} */

// NewTrade creates a new Trade instance with the specified contract and order details.
// Optional initial order status can be provided.
func NewTrade(contract *Contract, order *Order, orderStatus ...OrderStatusData) *Trade {
	var os OrderStatusData
	if len(orderStatus) > 0 {
		os = orderStatus[0]
	} else {
		os = OrderStatusData{
			OrderID: order.OrderID,
			Status:  OrderStatusPendingSubmit,
		}
	}
	return &Trade{
		Contract:    contract,
		Order:       order,
		OrderStatus: os,
		fills:       make([]*Fill, 0),
		logs:        []TradeLogEntry{{Time: time.Now().UTC(), Status: OrderStatusPendingSubmit}},
		done:        make(chan struct{}),
		ack:         make(chan struct{}),
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

// Ack returns a channel that will be closed when the trade is acknowledged by IB.
func (t *Trade) Ack() <-chan struct{} {
	return t.ack
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

// markAck closes the ack channel to signal trade acknowledge.
// This is an internal method and should be called with appropriate locking.
func (t *Trade) markAck() {
	// Ensure that the done channel is closed only once
	select {
	case <-t.ack:
		// Channel already closed
	default:
		close(t.ack)
	}
}

// reetAck resets the ack channel to a new open channel.
func (t *Trade) resetAck() {
	t.ack = make(chan struct{})
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

// Logs returns a copy of all log entries for this trade.
// Logs are sorted by timestamp.
func (t *Trade) Logs() []TradeLogEntry {
	t.mu.RLock()
	defer t.mu.RUnlock()

	logs := make([]TradeLogEntry, len(t.logs))
	copy(logs, t.logs)

	// Sort by timestamp on read
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Time.Before(logs[j].Time)
	})

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

	return fmt.Sprintf("Trade{Contract: %v, Order: %v, Status: %v, Fills: %v}",
		t.Contract, t.Order, t.OrderStatus, t.fills)
}
