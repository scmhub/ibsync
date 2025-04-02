package ibsync

import (
	"reflect"
	"testing"
	"time"
)

func TestStatus_IsActive(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{"PendingSubmit is active", PendingSubmit, true},
		{"ApiPending is active", ApiPending, true},
		{"PreSubmitted is active", PreSubmitted, true},
		{"Submitted is active", Submitted, true},
		{"Cancelled is not active", Cancelled, false},
		{"Filled is not active", Filled, false},
		{"ApiCancelled is not active", ApiCancelled, false},
		{"Inactive is not active", Inactive, false},
		{"Empty status is not active", Status(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsActive(); got != tt.want {
				t.Errorf("Status.IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatus_IsDone(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{"Filled is done", Filled, true},
		{"Cancelled is done", Cancelled, true},
		{"ApiCancelled is done", ApiCancelled, true},
		{"PendingSubmit is not done", PendingSubmit, false},
		{"Submitted is not done", Submitted, false},
		{"PreSubmitted is not done", PreSubmitted, false},
		{"ApiPending is not done", ApiPending, false},
		{"Inactive is not done", Inactive, false},
		{"Empty status is not done", Status(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsDone(); got != tt.want {
				t.Errorf("Status.IsDone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFill_Matches(t *testing.T) {
	timeStr := "20240102-15:04:05"
	testTime, _ := time.Parse("20060102-15:04:05", timeStr)

	tests := []struct {
		name   string
		fill   *Fill
		filter *ExecutionFilter
		want   bool
	}{
		{
			name:   "nil fill never matches",
			fill:   nil,
			filter: &ExecutionFilter{},
			want:   false,
		},
		{
			name: "empty filter matches valid fill",
			fill: &Fill{
				Contract: &Contract{Symbol: "AAPL", SecType: "STK"},
				Execution: &Execution{
					AcctNumber: "123",
					ClientID:   456,
					Exchange:   "NASDAQ",
					Side:       "BUY",
				},
				Time: testTime,
			},
			filter: NewExecutionFilter(),
			want:   true,
		},
		{
			name: "matching all criteria",
			fill: &Fill{
				Contract: &Contract{Symbol: "AAPL", SecType: "STK"},
				Execution: &Execution{
					AcctNumber: "123",
					ClientID:   456,
					Exchange:   "NASDAQ",
					Side:       "BUY",
				},
				Time: testTime,
			},
			filter: &ExecutionFilter{
				AcctCode: "123",
				ClientID: 456,
				Exchange: "NASDAQ",
				SecType:  "STK",
				Side:     "BUY",
				Symbol:   "AAPL",
				Time:     timeStr,
			},
			want: true,
		},
		{
			name: "non-matching account",
			fill: &Fill{
				Contract:  &Contract{Symbol: "AAPL", SecType: "STK"},
				Execution: &Execution{AcctNumber: "123"},
			},
			filter: func() *ExecutionFilter {
				f := NewExecutionFilter()
				f.AcctCode = "456"
				return f
			}(),
			want: false,
		},
		{
			name: "non-matching client ID",
			fill: &Fill{
				Contract: &Contract{Symbol: "AAPL", SecType: "STK"},
				Execution: &Execution{
					AcctNumber: "123",
					ClientID:   456,
				},
			},
			filter: func() *ExecutionFilter {
				f := NewExecutionFilter()
				f.ClientID = 789
				return f
			}(),
			want: false,
		},
		{
			name: "non-matching exchange",
			fill: &Fill{
				Contract: &Contract{Symbol: "AAPL", SecType: "STK"},
				Execution: &Execution{
					AcctNumber: "123",
					Exchange:   "NASDAQ",
				},
			},
			filter: func() *ExecutionFilter {
				f := NewExecutionFilter()
				f.Exchange = "NYSE"
				return f
			}(),
			want: false,
		},
		{
			name: "non-matching security type",
			fill: &Fill{
				Contract:  &Contract{Symbol: "AAPL", SecType: "STK"},
				Execution: &Execution{AcctNumber: "123"},
			},
			filter: func() *ExecutionFilter {
				f := NewExecutionFilter()
				f.SecType = "OPT"
				return f
			}(),
			want: false,
		},
		{
			name: "non-matching side",
			fill: &Fill{
				Contract: &Contract{Symbol: "AAPL", SecType: "STK"},
				Execution: &Execution{
					AcctNumber: "123",
					Side:       "BUY",
				},
			},
			filter: func() *ExecutionFilter {
				f := NewExecutionFilter()
				f.Side = "SELL"
				return f
			}(),
			want: false,
		},
		{
			name: "non-matching symbol",
			fill: &Fill{
				Contract:  &Contract{Symbol: "AAPL", SecType: "STK"},
				Execution: &Execution{AcctNumber: "123"},
			},
			filter: func() *ExecutionFilter {
				f := NewExecutionFilter()
				f.Symbol = "MSFT"
				return f
			}(),
			want: false,
		},
		{
			name: "non-matching time",
			fill: &Fill{
				Contract:  &Contract{Symbol: "AAPL", SecType: "STK"},
				Execution: &Execution{AcctNumber: "123"},
				Time:      testTime,
			},
			filter: func() *ExecutionFilter {
				f := NewExecutionFilter()
				f.Time = "20240102-16:04:05"
				return f
			}(),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fill.Matches(tt.filter); got != tt.want {
				t.Errorf("Fill.Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewTrade(t *testing.T) {
	contract := &Contract{Symbol: "AAPL"}
	order := &Order{OrderID: 123}
	customStatus := OrderStatus{
		OrderID: 123,
		Status:  Submitted,
	}

	tests := []struct {
		name     string
		contract *Contract
		order    *Order
		status   []OrderStatus
		wantErr  bool
	}{
		{
			name:     "basic creation",
			contract: contract,
			order:    order,
			status:   nil,
			wantErr:  false,
		},
		{
			name:     "with custom status",
			contract: contract,
			order:    order,
			status:   []OrderStatus{customStatus},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trade := NewTrade(tt.contract, tt.order, tt.status...)
			if (trade == nil) != tt.wantErr {
				t.Errorf("NewTrade() error = %v, wantErr %v", trade == nil, tt.wantErr)
				return
			}
			if trade != nil {
				if trade.Contract != tt.contract {
					t.Errorf("NewTrade() contract = %v, want %v", trade.Contract, tt.contract)
				}
				if trade.Order != tt.order {
					t.Errorf("NewTrade() order = %v, want %v", trade.Order, tt.order)
				}
			}
		})
	}
}

func TestTrade_Fills(t *testing.T) {
	trade := NewTrade(&Contract{}, &Order{})
	fill1 := &Fill{Time: time.Now()}
	fill2 := &Fill{Time: time.Now()}

	trade.addFill(fill1)
	trade.addFill(fill2)

	fills := trade.Fills()
	if len(fills) != 2 {
		t.Errorf("Trade.Fills() len = %v, want %v", len(fills), 2)
	}

	// Verify that modifying returned fills doesn't affect original
	fills[0] = &Fill{}
	if reflect.DeepEqual(fills[0], trade.fills[0]) {
		t.Error("Trade.Fills() returned slice should be a copy")
	}
}

func TestTrade_Logs(t *testing.T) {
	trade := NewTrade(&Contract{}, &Order{})

	// Verify initial log entry
	logs := trade.Logs()
	if len(logs) != 1 {
		t.Errorf("New trade should have 1 initial log entry, got %d", len(logs))
	}
	if logs[0].Status != PendingSubmit {
		t.Errorf("Initial log status = %v, want %v", logs[0].Status, PendingSubmit)
	}

	// Add new log entry
	newEntry := TradeLogEntry{
		Time:    time.Now(),
		Status:  Submitted,
		Message: "Test message",
	}
	trade.addLog(newEntry)

	// Verify log was added
	logs = trade.Logs()
	if len(logs) != 2 {
		t.Errorf("Trade.Logs() len = %v, want %v", len(logs), 2)
	}
	if !reflect.DeepEqual(logs[1], newEntry) {
		t.Errorf("Trade.Logs()[1] = %v, want %v", logs[1], newEntry)
	}
}

func TestTrade_IsActive(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{"active status", PendingSubmit, true},
		{"inactive status", Cancelled, false},
		{"done status", Filled, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trade := NewTrade(&Contract{}, &Order{}, OrderStatus{Status: tt.status})
			if got := trade.IsActive(); got != tt.want {
				t.Errorf("Trade.IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrade_IsDone(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{"done status", Filled, true},
		{"active status", PendingSubmit, false},
		{"inactive status", Inactive, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trade := NewTrade(&Contract{}, &Order{}, OrderStatus{Status: tt.status})
			if got := trade.IsDone(); got != tt.want {
				t.Errorf("Trade.IsDone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrade_Done(t *testing.T) {
	trade := NewTrade(&Contract{}, &Order{})

	// Test initial state
	select {
	case <-trade.Done():
		t.Error("Done channel should not be closed initially")
	default:
		// Expected behavior
	}

	// Mark as done
	trade.markDoneSafe()

	// Verify channel is closed
	select {
	case <-trade.Done():
		// Expected behavior
	default:
		t.Error("Done channel should be closed after markDoneSafe()")
	}

	// Verify multiple markDone calls don't panic
	trade.markDoneSafe()
}
