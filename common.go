package ibsync

import (
	"fmt"
	"strings"
	"time"
)

type AccountValue struct {
	Account  string
	Tag      string
	Value    string
	Currency string
}

type AccountValues []AccountValue

func (avs AccountValues) String() string {
	ss := make(map[string]string)
	var dots string
	for _, av := range avs {
		repeat := 40 - len(av.Tag) - len(av.Value)
		if repeat < 0 {
			repeat = 40 - len(av.Tag)
		}
		if repeat < 0 {
			repeat = 40
		}
		dots = strings.Repeat(".", repeat)
		ss[av.Account] = ss[av.Account] + fmt.Sprintf("\t%v%v%v %v\n", av.Tag, dots, av.Value, av.Currency)
	}
	s := "\n"
	for k, v := range ss {
		s = s + fmt.Sprintf("Account: %v\n", k)
		s = s + v
	}
	return s
}

type AccountSummary []AccountValue

func (as AccountSummary) String() string {
	return fmt.Sprint(AccountValues(as))
}

type ReceiveFA struct {
	FaDataType FaDataType
	Cxml       string
}

type TickData struct {
	Time     time.Time
	TickType TickType
	Price    float64
	Size     Decimal
}

type Tick interface {
	Type() TickType
	String() string
}

type TickPrice struct {
	TickType TickType
	Price    float64
	Attrib   TickAttrib
}

func (t TickPrice) Type() TickType {
	return t.TickType
}

func (t TickPrice) String() string {
	return fmt.Sprintf("<%v> price:%v, attrib:%v", TickName(t.TickType), t.Price, t.Attrib)
}

type TickSize struct {
	TickType TickType
	Size     Decimal
}

func (t TickSize) Type() TickType {
	return t.TickType
}

func (t TickSize) String() string {
	return fmt.Sprintf("<%v> size:%v", TickName(t.TickType), t.Size)
}

type TickOptionComputation struct {
	TickType   TickType
	TickAttrib int64
	ImpliedVol float64
	Delta      float64
	OptPrice   float64
	PvDividend float64
	Gamma      float64
	Vega       float64
	Theta      float64
	UndPrice   float64
}

func (t TickOptionComputation) Type() TickType {
	return t.TickType
}

func (t TickOptionComputation) String() string {
	return fmt.Sprintf("<%v> tickAttrib:%v, impliedVol:%v, delta:%v, optPrice:%v, pvDividend: %v, gamma:%v, vega:%v, theta:%v, undPrice:%v",
		TickName(t.TickType), t.TickAttrib, t.ImpliedVol, t.Delta, t.OptPrice, t.PvDividend, t.Gamma, t.Vega, t.Theta, t.UndPrice)
}

type TickGeneric struct {
	TickType TickType
	Value    float64
}

func (t TickGeneric) Type() TickType {
	return t.TickType
}

func (t TickGeneric) String() string {
	return fmt.Sprintf("<%v> value:%v", TickName(t.TickType), t.Value)
}

type TickString struct {
	TickType TickType
	Value    string
}

func (t TickString) Type() TickType {
	return t.TickType
}

func (t TickString) String() string {
	return fmt.Sprintf("<%v> value:%v", TickName(t.TickType), t.Value)
}

type TickEFP struct {
	TickType                 TickType
	BasisPoints              float64
	FormattedBasisPoints     string
	TotalDividends           float64
	HoldDays                 int64
	FutureLastTradeDate      string
	DividendImpact           float64
	DividendsToLastTradeDate float64
}

func (t TickEFP) Type() TickType {
	return t.TickType
}

func (t TickEFP) String() string {
	return fmt.Sprintf("<%v> basisPoints:%v, formattedBasisPoints:%v, totalDividends:%v, holdDays:%v, futureLastTradeDate:%v, dividendImpact:%v, dividendsToLastTradeDate:%v",
		TickName(t.TickType), t.BasisPoints, t.FormattedBasisPoints, t.TotalDividends, t.HoldDays, t.FutureLastTradeDate, t.DividendImpact, t.DividendsToLastTradeDate)
}

// TickByTick
type TickByTick interface {
	Timestamp() time.Time
	String() string
}

type TickByTickAllLast struct {
	Time              int64
	TickType          int64
	Price             float64
	Size              Decimal
	TickAttribLast    TickAttribLast
	Exchange          string
	SpecialConditions string
}

func (t TickByTickAllLast) Timestamp() time.Time {
	return time.Unix(t.Time, 0)
}

func (t TickByTickAllLast) String() string {
	return fmt.Sprintf("<TickByTickAllLast> timestamp:%v, tickType:%v, price:%v, size:%v, tickAttribLast:%v, exchange:%v, specialConditions:%v",
		t.Timestamp(), t.TickType, t.Price, t.Size, t.TickAttribLast, t.Exchange, t.SpecialConditions)
}

type TickByTickBidAsk struct {
	Time             int64
	BidPrice         float64
	AskPrice         float64
	BidSize          Decimal
	AskSize          Decimal
	TickAttribBidAsk TickAttribBidAsk
}

func (t TickByTickBidAsk) Timestamp() time.Time {
	return time.Unix(t.Time, 0)
}

func (t TickByTickBidAsk) String() string {
	return fmt.Sprintf("<TickByTickBidAsk> timestamp:%v, bidPrice:%v, askPrice:%v, bidSize:%v, askSize:%v, tickAttribBidAsk:%v",
		t.Timestamp(), t.BidPrice, t.AskPrice, t.BidSize, t.AskSize, t.TickAttribBidAsk)
}

type TickByTickMidPoint struct {
	Time     int64
	MidPoint float64
}

func (t TickByTickMidPoint) Timestamp() time.Time {
	return time.Unix(t.Time, 0)
}

func (t TickByTickMidPoint) String() string {
	return fmt.Sprintf("<TickByTickMidPoint> timestamp:%v, midPoint:%v", t.Timestamp(), t.MidPoint)
}

type MktDepthData struct {
	Time         time.Time
	Position     int64
	MarketMaker  string
	Operation    int64
	Side         int64
	Price        float64
	Size         Decimal
	IsSmartDepth bool
}

// DOMLevel represents a single level in the order book
type DOMLevel struct {
	Price       float64
	Size        Decimal
	MarketMaker string
}

// String provides a readable representation of a DOM level
func (dl DOMLevel) String() string {
	return fmt.Sprintf("Price: %v, Size: %v, Market Maker: %s", dl.Price, dl.Size, dl.MarketMaker)
}

type Dividends struct {
	Past12Months float64
	Next12Months float64
	NextDate     time.Time
	NextAmount   float64
}

type NewsArticle struct {
	ArticleType int64
	ArticleText string
}

type HistoricalNews struct {
	Time         time.Time
	ProviderCode string
	ArticleID    string
	Headline     string
}

type NewsTick struct {
	TimeStamp    int64
	ProviderCode string
	ArticleId    string
	Headline     string
	ExtraData    string
}

type NewsBulletin struct {
	MsgID       int64
	MsgType     int64
	NewsMessage string
	OriginExch  string
}

type PortfolioItem struct {
	Contract      *Contract
	Position      Decimal
	MarketPrice   float64
	MarketValue   float64
	AverageCost   float64
	UnrealizedPNL float64
	RealizedPNL   float64
	Account       string
}

type Position struct {
	Account  string
	Contract *Contract
	Position Decimal
	AvgCost  float64
}

type Pnl struct {
	Account       string
	ModelCode     string
	DailyPNL      float64
	UnrealizedPnl float64
	RealizedPNL   float64
}

type PnlSingle struct {
	Account       string
	ModelCode     string
	ConID         int64
	Position      Decimal
	DailyPNL      float64
	UnrealizedPnl float64
	RealizedPNL   float64
	Value         float64
}

type HistoricalSchedule struct {
	StartDateTime string
	EndDateTime   string
	TimeZone      string
	Sessions      []HistoricalSession
}

type OptionChain struct {
	Exchange        string
	UnderlyingConId int64
	TradingClass    string
	Multiplier      string
	Expirations     []string
	Strikes         []float64
}

type FundamentalRatios map[string]float64
