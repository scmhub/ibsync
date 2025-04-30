package ibsync

// NewStock creates a stock contract (STK) for the specified symbol, exchange, and currency.
// The symbol represents the stock ticker (e.g., "AAPL"), the exchange is where the stock is traded (e.g., "NASDAQ"),
// and the currency is the denomination of the stock (e.g., "USD").
func NewStock(symbol, exchange, currency string) *Contract {

	contract := NewContract()
	contract.Symbol = symbol
	contract.SecType = "STK"
	contract.Currency = currency
	contract.Exchange = exchange

	return contract
}

// NewOption creates an option contract (OPT) based on the symbol of the underlying asset, expiration date,
// strike price, option right, exchange, contract size, and currency.
// lastTradeDateOrContractMonth is the expiration date. "YYYYMM" format will specify the last trading month or "YYYYMMDD" format the last trading day.
// The right specifies if it's a call ("C" or "CALL") or a put ("P" or "PUT") option.
func NewOption(symbol, lastTradeDateOrContractMonth string, strike float64, right, exchange, multiplier, currency string) *Contract {

	contract := NewContract()
	contract.Symbol = symbol
	contract.SecType = "OPT"
	contract.Currency = currency
	contract.Exchange = exchange
	contract.LastTradeDateOrContractMonth = lastTradeDateOrContractMonth
	contract.Right = right
	contract.Strike = strike
	contract.Multiplier = multiplier

	return contract
}

// NewFuture creates a future contract (FUT) based on the symbol of the underlying asset, expiration date,
// exchange, contract size (multiplier), and currency.
// lastTradeDateOrContractMonth is the expiration date. "YYYYMM" format will specify the last trading month or "YYYYMMDD" format the last trading day.
func NewFuture(symbol, lastTradeDateOrContractMonth string, exchange, multiplier, currency string) *Contract {

	contract := NewContract()
	contract.Symbol = symbol
	contract.SecType = "FUT"
	contract.Currency = currency
	contract.Exchange = exchange
	contract.LastTradeDateOrContractMonth = lastTradeDateOrContractMonth
	contract.Multiplier = multiplier

	return contract
}

// NewContFuture creates a continuous future contract (CONFUT) for the given symbol, exchange, contract size, and currency.
// A continuous future contract represents a series of futures contracts for the same asset.
func NewContFuture(symbol, exchange, multiplier, currency string) *Contract {

	contract := NewContract()
	contract.Symbol = symbol
	contract.SecType = "CONTFUT"
	contract.Currency = currency
	contract.Exchange = exchange
	contract.Multiplier = multiplier

	return contract
}

// NewForex creates a forex contract (CASH) for a currency pair.
// symbol is the base currency, and currency is the quote currency.
// For a pair like "EURUSD", "EUR" is the symbol and "USD" the currency.
func NewForex(symbol, exchange, currency string) *Contract {

	contract := NewContract()
	contract.Symbol = symbol
	contract.SecType = "CASH"
	contract.Currency = currency
	contract.Exchange = exchange

	return contract
}

// NewIndex creates an index contract (IND) for the given index symbol, exchange, and currency.
// The symbol typically represents a stock index (e.g., "SPX").
func NewIndex(symbol, exchange, currency string) *Contract {

	contract := NewContract()
	contract.Symbol = symbol
	contract.SecType = "IND"
	contract.Currency = currency
	contract.Exchange = exchange

	return contract
}

// NewCFD creates a contract for difference (CFD) for the specified symbol, exchange, and currency.
func NewCFD(symbol, exchange, currency string) *Contract {

	contract := NewContract()
	contract.Symbol = symbol
	contract.SecType = "CFD"
	contract.Currency = currency
	contract.Exchange = exchange

	return contract
}

// NewCommodity creates a commodity contract (CMDTY) for the given symbol, exchange, and currency.
func NewCommodity(symbol, exchange, currency string) *Contract {

	contract := NewContract()
	contract.Symbol = symbol
	contract.SecType = "CMDTY"
	contract.Currency = currency
	contract.Exchange = exchange

	return contract
}

// NewBond creates a bond contract (Bond).
func NewBond(symbol, exchange, currency string) *Contract {

	contract := NewContract()
	contract.Symbol = symbol
	contract.SecType = "Bond"
	contract.Currency = currency
	contract.Exchange = exchange

	return contract
}

// NewFutureOption creates a future option contract (FOP) based on the symbol of the underlying asset, expiration date,
// strike price, option right, exchange, contract size, and currency.
// lastTradeDateOrContractMonth is the expiration date. "YYYYMM" format will specify the last trading month or "YYYYMMDD" format the last trading day.
// The right specifies if it's a call ("C" or "CALL") or a put ("P" or "PUT") option.
func NewFutureOption(symbol, lastTradeDateOrContractMonth string, strike float64, right, exchange, multiplier, currency string) *Contract {

	contract := NewContract()
	contract.Symbol = symbol
	contract.SecType = "FOP"
	contract.Currency = currency
	contract.Exchange = exchange
	contract.LastTradeDateOrContractMonth = lastTradeDateOrContractMonth
	contract.Right = right
	contract.Strike = strike
	contract.Multiplier = multiplier

	return contract
}

// NewMutualFund creates a mutual fund contract (FUND).
func NewMutualFund() *Contract {

	contract := NewContract()
	contract.SecType = "FUND"

	return contract
}

// NewWarrant creates a warrant contract (WAR).
func NewWarrant() *Contract {

	contract := NewContract()
	contract.SecType = "WAR"

	return contract
}

// NewBag creates a bag contract (BAG), which may represent a collection of contracts bundled together.
func NewBag() *Contract {

	contract := NewContract()
	contract.SecType = "BAG"

	return contract
}

// NewCrypto creates a cryptocurrency contract (CRYPTO) for the specified symbol, exchange, and currency.
// The symbol represents the cryptocurrency being traded (e.g., "BTC").
func NewCrypto(symbol, exchange, currency string) *Contract {

	contract := NewContract()
	contract.Symbol = symbol
	contract.SecType = "CRYPTO"
	contract.Currency = currency
	contract.Exchange = exchange

	return contract
}
