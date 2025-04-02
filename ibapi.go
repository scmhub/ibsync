/*
The primary aim of this file is to re-export all structs, constants, and functions from the
ibapi package, allowing developers to access them through a single cohesive package. This
simplifies the process of integrating with the Interactive Brokers API by reducing the need
to manage multiple imports and providing a more straightforward API surface for developers.
*/
package ibsync

import "github.com/scmhub/ibapi"

const (
	UNSET_INT       = ibapi.UNSET_INT
	UNSET_LONG      = ibapi.UNSET_LONG
	UNSET_FLOAT     = ibapi.UNSET_FLOAT
	INFINITY_STRING = ibapi.INFINITY_STRING
)

var (
	UNSET_DECIMAL    = ibapi.UNSET_DECIMAL
	ZERO             = ibapi.ZERO
	StringToDecimal  = ibapi.StringToDecimal
	DecimalToString  = ibapi.DecimalToString
	IntMaxString     = ibapi.IntMaxString
	LongMaxString    = ibapi.LongMaxString
	FloatMaxString   = ibapi.FloatMaxString
	DecimalMaxString = ibapi.DecimalMaxString
	Logger           = ibapi.Logger
	SetLogLevel      = ibapi.SetLogLevel
	SetConsoleWriter = ibapi.SetConsoleWriter
	TickName         = ibapi.TickName
)

type (
	Bar                             = ibapi.Bar
	RealTimeBar                     = ibapi.RealTimeBar
	CommissionAndFeesReport         = ibapi.CommissionAndFeesReport
	ComboLeg                        = ibapi.ComboLeg
	Contract                        = ibapi.Contract
	ContractDescription             = ibapi.ContractDescription
	ContractDetails                 = ibapi.ContractDetails
	DeltaNeutralContract            = ibapi.DeltaNeutralContract
	Decimal                         = ibapi.Decimal
	DepthMktDataDescription         = ibapi.DepthMktDataDescription
	Execution                       = ibapi.Execution
	ExecutionFilter                 = ibapi.ExecutionFilter
	FaDataType                      = ibapi.FaDataType
	FamilyCode                      = ibapi.FamilyCode
	FundDistributionPolicyIndicator = ibapi.FundDistributionPolicyIndicator
	HistogramData                   = ibapi.HistogramData
	HistoricalSession               = ibapi.HistoricalSession
	HistoricalTick                  = ibapi.HistoricalTick
	HistoricalTickBidAsk            = ibapi.HistoricalTickBidAsk
	HistoricalTickLast              = ibapi.HistoricalTickLast
	IneligibilityReason             = ibapi.IneligibilityReason
	NewsProvider                    = ibapi.NewsProvider
	Order                           = ibapi.Order
	OrderCancel                     = ibapi.OrderCancel
	OrderID                         = ibapi.OrderID
	OrderState                      = ibapi.OrderState
	PriceIncrement                  = ibapi.PriceIncrement
	ScanData                        = ibapi.ScanData
	ScannerSubscription             = ibapi.ScannerSubscription
	SmartComponent                  = ibapi.SmartComponent
	SoftDollarTier                  = ibapi.SoftDollarTier
	TagValue                        = ibapi.TagValue
	TickerID                        = ibapi.TickerID
	TickType                        = ibapi.TickType
	TickAttrib                      = ibapi.TickAttrib
	TickAttribLast                  = ibapi.TickAttribLast
	TickAttribBidAsk                = ibapi.TickAttribBidAsk
	WshEventData                    = ibapi.WshEventData
)

var (
	NewBar                     = ibapi.NewBar
	NewRealTimeBar             = ibapi.NewRealTimeBar
	NewCommissionAndFeesReport = ibapi.NewCommissionAndFeesReport
	NewComboLeg                = ibapi.NewComboLeg
	NewContract                = ibapi.NewContract
	NewContractDescription     = ibapi.NewContractDescription
	NewContractDetails         = ibapi.NewContractDetails
	NewDeltaNeutralContract    = ibapi.NewDeltaNeutralContract
	NewDepthMktDataDescription = ibapi.NewDepthMktDataDescription
	NewExecution               = ibapi.NewExecution
	NewExecutionFilter         = ibapi.NewExecutionFilter
	NewFamilyCode              = ibapi.NewFamilyCode
	NewHistogramData           = ibapi.NewHistogramData
	NewHistoricalSession       = ibapi.NewHistoricalSession
	NewHistoricalTick          = ibapi.NewHistoricalTick
	NewHistoricalTickBidAsk    = ibapi.NewHistoricalTickBidAsk
	NewHistoricalTickLast      = ibapi.NewHistoricalTickLast
	NewNewsProvider            = ibapi.NewNewsProvider
	NewOrder                   = ibapi.NewOrder
	NewOrderCancel             = ibapi.NewOrderCancel
	NewOrderState              = ibapi.NewOrderState
	NewPriceIncrement          = ibapi.NewPriceIncrement
	NewScannerSubscription     = ibapi.NewScannerSubscription
	NewSmartComponent          = ibapi.NewSmartComponent
	NewSoftDollarTier          = ibapi.NewSoftDollarTier
	NewTagValue                = ibapi.NewTagValue
	NewTickAttrib              = ibapi.NewTickAttrib
	NewTickAttribLast          = ibapi.NewTickAttribLast
	NewTickAttribBidAsk        = ibapi.NewTickAttribBidAsk
	NewWshEventData            = ibapi.NewWshEventData
)

var (
	// ibapi custom contracts
	AtAuction                       = ibapi.AtAuction
	Discretionary                   = ibapi.Discretionary
	MarketOrder                     = ibapi.MarketOrder
	MarketIfTouched                 = ibapi.MarketIfTouched
	MarketOnClose                   = ibapi.MarketOnClose
	MarketOnOpen                    = ibapi.MarketOnOpen
	MidpointMatch                   = ibapi.MidpointMatch
	Midprice                        = ibapi.Midprice
	PeggedToMarket                  = ibapi.PeggedToMarket
	PeggedToStock                   = ibapi.PeggedToStock
	RelativePeggedToPrimary         = ibapi.RelativePeggedToPrimary
	SweepToFill                     = ibapi.SweepToFill
	AuctionLimit                    = ibapi.AuctionLimit
	AuctionPeggedToStock            = ibapi.AuctionPeggedToStock
	AuctionRelative                 = ibapi.AuctionRelative
	Block                           = ibapi.Block
	BoxTop                          = ibapi.BoxTop
	LimitOrder                      = ibapi.LimitOrder
	LimitOrderWithCashQty           = ibapi.LimitOrderWithCashQty
	LimitIfTouched                  = ibapi.LimitIfTouched
	LimitOnClose                    = ibapi.LimitOnClose
	LimitOnOpen                     = ibapi.LimitOnOpen
	PassiveRelative                 = ibapi.PassiveRelative
	PeggedToMidpoint                = ibapi.PeggedToMidpoint
	BracketOrder                    = ibapi.BracketOrder
	MarketToLimit                   = ibapi.MarketToLimit
	MarketWithProtection            = ibapi.MarketWithProtection
	Stop                            = ibapi.Stop
	StopLimit                       = ibapi.StopLimit
	StopWithProtection              = ibapi.StopWithProtection
	TrailingStop                    = ibapi.TrailingStop
	TrailingStopLimit               = ibapi.TrailingStopLimit
	ComboLimitOrder                 = ibapi.ComboLimitOrder
	ComboMarketOrder                = ibapi.ComboMarketOrder
	LimitOrderForComboWithLegPrices = ibapi.LimitOrderForComboWithLegPrices
	RelativeLimitCombo              = ibapi.RelativeLimitCombo
	RelativeMarketCombo             = ibapi.RelativeMarketCombo
	OneCancelsAll                   = ibapi.OneCancelsAll
	Volatility                      = ibapi.Volatility
	MarketFHedge                    = ibapi.MarketFHedge
	PeggedToBenchmark               = ibapi.PeggedToBenchmark
	AttachAdjustableToStop          = ibapi.AttachAdjustableToStop
	AttachAdjustableToStopLimit     = ibapi.AttachAdjustableToStopLimit
	AttachAdjustableToTrail         = ibapi.AttachAdjustableToTrail
	WhatIfLimitOrder                = ibapi.WhatIfLimitOrder
	NewPriceCondition               = ibapi.NewPriceCondition
	NewExecutionCondition           = ibapi.NewExecutionCondition
	NewMarginCondition              = ibapi.NewMarginCondition
	NewPercentageChangeCondition    = ibapi.NewPercentageChangeCondition
	NewTimeCondition                = ibapi.NewTimeCondition
	NewVolumeCondition              = ibapi.NewVolumeCondition
	LimitIBKRATS                    = ibapi.LimitIBKRATS
	LimitOrderWithManualOrderTime   = ibapi.LimitOrderWithManualOrderTime
	PegBestUpToMidOrder             = ibapi.PegBestUpToMidOrder
	PegBestOrder                    = ibapi.PegBestOrder
	PegMidOrder                     = ibapi.PegMidOrder
	LimitOrderWithCustomerAccount   = ibapi.LimitOrderWithCustomerAccount
	LimitOrderWithIncludeOvernight  = ibapi.LimitOrderWithIncludeOvernight
	CancelOrderEmpty                = ibapi.CancelOrderEmpty
	CancelOrderWithManualTime       = ibapi.CancelOrderWithManualTime
	LimitOrderWithCmeTaggingFields  = ibapi.LimitOrderWithCmeTaggingFields
	OrderCancelWithCmeTaggingFields = ibapi.OrderCancelWithCmeTaggingFields
)

const (
	// TickType
	BID_SIZE                  = ibapi.BID_SIZE
	BID                       = ibapi.BID
	ASK                       = ibapi.ASK
	ASK_SIZE                  = ibapi.ASK_SIZE
	LAST                      = ibapi.LAST
	LAST_SIZE                 = ibapi.LAST_SIZE
	HIGH                      = ibapi.HIGH
	LOW                       = ibapi.LOW
	VOLUME                    = ibapi.VOLUME
	CLOSE                     = ibapi.CLOSE
	BID_OPTION_COMPUTATION    = ibapi.BID_OPTION_COMPUTATION
	ASK_OPTION_COMPUTATION    = ibapi.ASK_OPTION_COMPUTATION
	LAST_OPTION_COMPUTATION   = ibapi.LAST_OPTION_COMPUTATION
	MODEL_OPTION              = ibapi.MODEL_OPTION
	OPEN                      = ibapi.OPEN
	LOW_13_WEEK               = ibapi.LOW_13_WEEK
	HIGH_13_WEEK              = ibapi.HIGH_13_WEEK
	LOW_26_WEEK               = ibapi.LOW_26_WEEK
	HIGH_26_WEEK              = ibapi.HIGH_26_WEEK
	LOW_52_WEEK               = ibapi.LOW_52_WEEK
	HIGH_52_WEEK              = ibapi.HIGH_52_WEEK
	AVG_VOLUME                = ibapi.AVG_VOLUME
	OPEN_INTEREST             = ibapi.OPEN_INTEREST
	OPTION_HISTORICAL_VOL     = ibapi.OPTION_HISTORICAL_VOL
	OPTION_IMPLIED_VOL        = ibapi.OPTION_IMPLIED_VOL
	OPTION_BID_EXCH           = ibapi.OPTION_BID_EXCH
	OPTION_ASK_EXCH           = ibapi.OPTION_ASK_EXCH
	OPTION_CALL_OPEN_INTEREST = ibapi.OPTION_CALL_OPEN_INTEREST
	OPTION_PUT_OPEN_INTEREST  = ibapi.OPTION_PUT_OPEN_INTEREST
	OPTION_CALL_VOLUME        = ibapi.OPTION_CALL_VOLUME
	OPTION_PUT_VOLUME         = ibapi.OPTION_PUT_VOLUME
	INDEX_FUTURE_PREMIUM      = ibapi.INDEX_FUTURE_PREMIUM
	BID_EXCH                  = ibapi.BID_EXCH
	ASK_EXCH                  = ibapi.ASK_EXCH
	AUCTION_VOLUME            = ibapi.AUCTION_VOLUME
	AUCTION_PRICE             = ibapi.AUCTION_PRICE
	AUCTION_IMBALANCE         = ibapi.AUCTION_IMBALANCE
	MARK_PRICE                = ibapi.MARK_PRICE
	BID_EFP_COMPUTATION       = ibapi.BID_EFP_COMPUTATION
	ASK_EFP_COMPUTATION       = ibapi.ASK_EFP_COMPUTATION
	LAST_EFP_COMPUTATION      = ibapi.LAST_EFP_COMPUTATION
	OPEN_EFP_COMPUTATION      = ibapi.OPEN_EFP_COMPUTATION
	HIGH_EFP_COMPUTATION      = ibapi.HIGH_EFP_COMPUTATION
	LOW_EFP_COMPUTATION       = ibapi.LOW_EFP_COMPUTATION
	CLOSE_EFP_COMPUTATION     = ibapi.CLOSE_EFP_COMPUTATION
	LAST_TIMESTAMP            = ibapi.LAST_TIMESTAMP
	SHORTABLE                 = ibapi.SHORTABLE
	FUNDAMENTAL_RATIOS        = ibapi.FUNDAMENTAL_RATIOS
	RT_VOLUME                 = ibapi.RT_VOLUME
	HALTED                    = ibapi.HALTED
	BID_YIELD                 = ibapi.BID_YIELD
	ASK_YIELD                 = ibapi.ASK_YIELD
	LAST_YIELD                = ibapi.LAST_YIELD
	CUST_OPTION_COMPUTATION   = ibapi.CUST_OPTION_COMPUTATION
	TRADE_COUNT               = ibapi.TRADE_COUNT
	TRADE_RATE                = ibapi.TRADE_RATE
	VOLUME_RATE               = ibapi.VOLUME_RATE
	LAST_RTH_TRADE            = ibapi.LAST_RTH_TRADE
	RT_HISTORICAL_VOL         = ibapi.RT_HISTORICAL_VOL
	IB_DIVIDENDS              = ibapi.IB_DIVIDENDS
	BOND_FACTOR_MULTIPLIER    = ibapi.BOND_FACTOR_MULTIPLIER
	REGULATORY_IMBALANCE      = ibapi.REGULATORY_IMBALANCE
	NEWS_TICK                 = ibapi.NEWS_TICK
	SHORT_TERM_VOLUME_3_MIN   = ibapi.SHORT_TERM_VOLUME_3_MIN
	SHORT_TERM_VOLUME_5_MIN   = ibapi.SHORT_TERM_VOLUME_5_MIN
	SHORT_TERM_VOLUME_10_MIN  = ibapi.SHORT_TERM_VOLUME_10_MIN
	DELAYED_BID               = ibapi.DELAYED_BID
	DELAYED_ASK               = ibapi.DELAYED_ASK
	DELAYED_LAST              = ibapi.DELAYED_LAST
	DELAYED_BID_SIZE          = ibapi.DELAYED_BID_SIZE
	DELAYED_ASK_SIZE          = ibapi.DELAYED_ASK_SIZE
	DELAYED_LAST_SIZE         = ibapi.DELAYED_LAST_SIZE
	DELAYED_HIGH              = ibapi.DELAYED_HIGH
	DELAYED_LOW               = ibapi.DELAYED_LOW
	DELAYED_VOLUME            = ibapi.DELAYED_VOLUME
	DELAYED_CLOSE             = ibapi.DELAYED_CLOSE
	DELAYED_OPEN              = ibapi.DELAYED_OPEN
	RT_TRD_VOLUME             = ibapi.RT_TRD_VOLUME
	CREDITMAN_MARK_PRICE      = ibapi.CREDITMAN_MARK_PRICE
	CREDITMAN_SLOW_MARK_PRICE = ibapi.CREDITMAN_SLOW_MARK_PRICE
	DELAYED_BID_OPTION        = ibapi.DELAYED_BID_OPTION
	DELAYED_ASK_OPTION        = ibapi.DELAYED_ASK_OPTION
	DELAYED_LAST_OPTION       = ibapi.DELAYED_LAST_OPTION
	DELAYED_MODEL_OPTION      = ibapi.DELAYED_MODEL_OPTION
	LAST_EXCH                 = ibapi.LAST_EXCH
	LAST_REG_TIME             = ibapi.LAST_REG_TIME
	FUTURES_OPEN_INTEREST     = ibapi.FUTURES_OPEN_INTEREST
	AVG_OPT_VOLUME            = ibapi.AVG_OPT_VOLUME
	DELAYED_LAST_TIMESTAMP    = ibapi.DELAYED_LAST_TIMESTAMP
	SHORTABLE_SHARES          = ibapi.SHORTABLE_SHARES
	DELAYED_HALTED            = ibapi.DELAYED_HALTED
	REUTERS_2_MUTUAL_FUNDS    = ibapi.REUTERS_2_MUTUAL_FUNDS
	ETF_NAV_CLOSE             = ibapi.ETF_NAV_CLOSE
	ETF_NAV_PRIOR_CLOSE       = ibapi.ETF_NAV_PRIOR_CLOSE
	ETF_NAV_BID               = ibapi.ETF_NAV_BID
	ETF_NAV_ASK               = ibapi.ETF_NAV_ASK
	ETF_NAV_LAST              = ibapi.ETF_NAV_LAST
	ETF_FROZEN_NAV_LAST       = ibapi.ETF_FROZEN_NAV_LAST
	ETF_NAV_HIGH              = ibapi.ETF_NAV_HIGH
	ETF_NAV_LOW               = ibapi.ETF_NAV_LOW
	SOCIAL_MARKET_ANALYTICS   = ibapi.SOCIAL_MARKET_ANALYTICS
	ESTIMATED_IPO_MIDPOINT    = ibapi.ESTIMATED_IPO_MIDPOINT
	FINAL_IPO_LAST            = ibapi.FINAL_IPO_LAST
	DELAYED_YIELD_BID         = ibapi.DELAYED_YIELD_BID
	DELAYED_YIELD_ASK         = ibapi.DELAYED_YIELD_ASK
	NOT_SET                   = ibapi.NOT_SET
)
