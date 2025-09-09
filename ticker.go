package ibsync

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Ticker represents real-time market data for a financial contract.
//
// The Ticker struct captures comprehensive market information including:
// - Current price data (bid, ask, last price)
// - Historical price metrics (open, high, low, close)
// - Trading volume and statistics
// - Volatility indicators
// - Options-specific data like greeks
//
// Thread-safe access is provided through mutex-protected getter methods.
//
// Market Data Types:
// - Level 1 streaming ticks stored in 'ticks'
// - Level 2 market depth ticks stored in 'domTicks'
// - Order book (DOM) available in 'domBids' and 'domAsks'
// - Tick-by-tick data stored in 'tickByTicks'
//
// Options Greeks:
// - Bid, ask, and last greeks stored in 'bidGreeks', 'askGreeks', and 'lastGreeks'
// - Model-calculated greeks available in 'modelGreeks'
type Ticker struct {
	mu                  sync.Mutex
	contract            *Contract
	time                time.Time
	marketDataType      int64
	minTick             float64
	bid                 float64
	bidSize             Decimal
	bidExchange         string
	ask                 float64
	askSize             Decimal
	askExchange         string
	last                float64
	lastSize            Decimal
	lastExchange        string
	lastTimestamp       string
	prevBid             float64
	prevBidSize         Decimal
	prevAsk             float64
	prevAskSize         Decimal
	prevLast            float64
	prevLastSize        Decimal
	volume              Decimal
	open                float64
	high                float64
	low                 float64
	close               float64
	vwap                float64
	low13Week           float64
	high13Week          float64
	low26Week           float64
	high26Week          float64
	low52Week           float64
	high52Week          float64
	bidYield            float64
	askYield            float64
	lastYield           float64
	markPrice           float64
	halted              float64
	rtHistVolatility    float64
	rtVolume            float64
	rtTradeVolume       float64
	rtTime              time.Time
	avVolume            Decimal
	tradeCount          float64
	tradeRate           float64
	volumeRate          float64
	shortableShares     Decimal
	indexFuturePremium  float64
	futuresOpenInterest Decimal
	putOpenInterest     Decimal
	callOpenInterest    Decimal
	putVolume           Decimal
	callVolume          Decimal
	avOptionVolume      Decimal
	histVolatility      float64
	impliedVolatility   float64
	dividends           Dividends
	fundamentalRatios   FundamentalRatios
	ticks               []TickData
	tickByTicks         []TickByTick
	domBids             map[int64]DOMLevel
	domAsks             map[int64]DOMLevel
	domTicks            []MktDepthData
	bidGreeks           TickOptionComputation
	askGreeks           TickOptionComputation
	lastGreeks          TickOptionComputation
	modelGreeks         TickOptionComputation
	auctionVolume       Decimal
	auctionPrice        float64
	auctionImbalance    Decimal
	regulatoryImbalance Decimal
	bboExchange         string
	snapshotPermissions int64
}

// NewTicker creates a new Ticker instance for the given contract.
func NewTicker(contract *Contract) *Ticker {
	return &Ticker{
		contract: contract,
		domBids:  make(map[int64]DOMLevel),
		domAsks:  make(map[int64]DOMLevel),
	}
}

// Contract returns the financial contract associated with this Ticker.
func (t *Ticker) Contract() *Contract {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.contract
}

func (t *Ticker) Time() time.Time {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.time
}

func (t *Ticker) MarketDataType() int64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.marketDataType
}

func (t *Ticker) setMarketDataType(marketDataType int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.marketDataType = marketDataType
}

func (t *Ticker) MinTick() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.minTick
}

func (t *Ticker) Bid() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.bid
}

func (t *Ticker) BidSize() Decimal {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.bidSize
}

func (t *Ticker) BidExchange() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.bidExchange
}

func (t *Ticker) Ask() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.ask
}

func (t *Ticker) AskSize() Decimal {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.askSize
}

func (t *Ticker) AskExchange() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.askExchange
}

func (t *Ticker) Last() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.last
}

// Getters pour les autres champs
func (t *Ticker) LastSize() Decimal {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.lastSize
}

func (t *Ticker) LastExchange() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.lastExchange
}

func (t *Ticker) LastTimestamp() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.lastTimestamp
}

func (t *Ticker) PrevBid() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.prevBid
}

func (t *Ticker) PrevBidSize() Decimal {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.prevBidSize
}

func (t *Ticker) PrevAsk() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.prevAsk
}

func (t *Ticker) PrevAskSize() Decimal {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.prevAskSize
}

func (t *Ticker) PrevLast() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.prevLast
}

func (t *Ticker) PrevLastSize() Decimal {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.prevLastSize
}

func (t *Ticker) Volume() Decimal {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.volume
}

func (t *Ticker) Open() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.open
}

func (t *Ticker) High() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.high
}

func (t *Ticker) Low() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.low
}

func (t *Ticker) Close() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.close
}

func (t *Ticker) Vwap() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.vwap
}

func (t *Ticker) Low13Week() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.low13Week
}

func (t *Ticker) High13Week() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.high13Week
}

func (t *Ticker) Low26Week() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.low26Week
}

func (t *Ticker) High26Week() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.high26Week
}

func (t *Ticker) Low52Week() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.low52Week
}

func (t *Ticker) High52Week() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.high52Week
}

func (t *Ticker) BidYield() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.bidYield
}

func (t *Ticker) AskYield() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.askYield
}

func (t *Ticker) LastYield() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.lastYield
}

func (t *Ticker) MarkPrice() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.markPrice
}

func (t *Ticker) Halted() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.halted
}

func (t *Ticker) RtHistVolatility() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.rtHistVolatility
}

func (t *Ticker) RtVolume() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.rtVolume
}

func (t *Ticker) RtTradeVolume() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.rtTradeVolume
}

func (t *Ticker) RtTime() time.Time {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.rtTime
}

func (t *Ticker) AvVolume() Decimal {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.avVolume
}

func (t *Ticker) TradeCount() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.tradeCount
}

func (t *Ticker) TradeRate() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.tradeRate
}

func (t *Ticker) VolumeRate() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.volumeRate
}

func (t *Ticker) ShortableShares() Decimal {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.shortableShares
}

func (t *Ticker) IndexFuturePremium() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.indexFuturePremium
}

func (t *Ticker) FuturesOpenInterest() Decimal {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.futuresOpenInterest
}

func (t *Ticker) PutOpenInterest() Decimal {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.putOpenInterest
}

func (t *Ticker) CallOpenInterest() Decimal {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.callOpenInterest
}

func (t *Ticker) PutVolume() Decimal {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.putVolume
}

func (t *Ticker) CallVolume() Decimal {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.callVolume
}

func (t *Ticker) AvOptionVolume() Decimal {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.avOptionVolume
}

func (t *Ticker) HistVolatility() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.histVolatility
}

func (t *Ticker) ImpliedVolatility() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.impliedVolatility
}

func (t *Ticker) Dividends() Dividends {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.dividends
}

func (t *Ticker) FundamentalRatios() FundamentalRatios {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.fundamentalRatios
}

func (t *Ticker) Ticks() []TickData {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.ticks
}

func (t *Ticker) TickByTicks() []TickByTick {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.tickByTicks
}

func (t *Ticker) DomBids() []DOMLevel {
	t.mu.Lock()
	defer t.mu.Unlock()
	var dls []DOMLevel
	for _, dl := range t.domBids {
		dls = append(dls, dl)
	}
	return dls
}

func (t *Ticker) DomAsks() []DOMLevel {
	t.mu.Lock()
	defer t.mu.Unlock()
	var dls []DOMLevel
	for _, dl := range t.domAsks {
		dls = append(dls, dl)
	}
	return dls
}

func (t *Ticker) DomTicks() []MktDepthData {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.domTicks
}

func (t *Ticker) BidGreeks() TickOptionComputation {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.bidGreeks
}

func (t *Ticker) AskGreeks() TickOptionComputation {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.askGreeks
}

func (t *Ticker) midGreeks() TickOptionComputation {
	if t.bidGreeks.Type() == 0 || t.askGreeks.Type() == 0 {
		return TickOptionComputation{}
	}
	mg := TickOptionComputation{
		ImpliedVol: (t.bidGreeks.ImpliedVol + t.askGreeks.ImpliedVol) / 2,
		Delta:      (t.bidGreeks.Delta + t.askGreeks.Delta) / 2,
		OptPrice:   (t.bidGreeks.OptPrice + t.askGreeks.OptPrice) / 2,
		PvDividend: (t.bidGreeks.PvDividend + t.askGreeks.PvDividend) / 2,
		Gamma:      (t.bidGreeks.Gamma + t.askGreeks.Gamma) / 2,
		Vega:       (t.bidGreeks.Vega + t.askGreeks.Vega) / 2,
		Theta:      (t.bidGreeks.Theta + t.askGreeks.Theta) / 2,
		UndPrice:   (t.bidGreeks.UndPrice + t.askGreeks.UndPrice) / 2,
	}
	return mg
}

func (t *Ticker) MidGreeks() TickOptionComputation {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.midGreeks()
}

func (t *Ticker) LastGreeks() TickOptionComputation {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.lastGreeks
}

func (t *Ticker) ModelGreeks() TickOptionComputation {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.modelGreeks
}

// Greeks returns the most representative option Greeks.
//
// Selection priority:
// 1. Midpoint Greeks (average of bid and ask Greeks)
// 2. Last trade Greeks
// 3. Model-calculated Greeks
func (t *Ticker) Greeks() TickOptionComputation {
	t.mu.Lock()
	defer t.mu.Unlock()
	greeks := t.midGreeks()
	if greeks != (TickOptionComputation{}) {
		return greeks
	}
	if t.lastGreeks != (TickOptionComputation{}) {
		return t.lastGreeks
	}
	if t.modelGreeks != (TickOptionComputation{}) {
		return t.modelGreeks
	}
	return TickOptionComputation{}
}

func (t *Ticker) AuctionVolume() Decimal {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.auctionVolume
}

func (t *Ticker) AuctionPrice() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.auctionPrice
}

func (t *Ticker) AuctionImbalance() Decimal {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.auctionImbalance
}

func (t *Ticker) RegulatoryImbalance() Decimal {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.regulatoryImbalance
}

func (t *Ticker) BboExchange() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.bboExchange
}

func (t *Ticker) SnapshotPermissions() int64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.snapshotPermissions
}

func (t *Ticker) hasBidAsk() bool {
	return t.bid > 0 && t.ask > 0
}

func (t *Ticker) HasBidAsk() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.hasBidAsk()
}

// MidPoint calculates the average of the current bid and ask prices.
func (t *Ticker) MidPoint() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.hasBidAsk() {
		return (t.bid + t.ask) * 0.5
	}
	return math.NaN()
}

// MarketPrice determines the most appropriate current market price.
//
// Price selection priority:
// 1. If last price is within the bid-ask spread, use last price.
// 2. If no last price fits the spread, use midpoint (average of bid and ask).
// 3. If no bid-ask available, return last price.
func (t *Ticker) MarketPrice() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.hasBidAsk() {
		if t.bid <= t.last && t.last <= t.ask {
			return t.last
		}
		return (t.bid + t.ask) * 0.5
	}
	return t.last
}

func (t *Ticker) SetTickPrice(tp TickPrice) {
	t.mu.Lock()
	defer t.mu.Unlock()
	var size Decimal
	switch tp.TickType {
	case BID, DELAYED_BID:
		if tp.Price == t.bid {
			return
		}
		t.prevBid = t.bid
		t.bid = tp.Price
	case ASK, DELAYED_ASK:
		if tp.Price == t.ask {
			return
		}
		t.prevAsk = t.ask
		t.ask = tp.Price
	case LAST, DELAYED_LAST:
		if tp.Price == t.last {
			return
		}
		t.prevLast = t.last
		t.last = tp.Price
	case HIGH, DELAYED_HIGH:
		t.high = tp.Price
	case LOW, DELAYED_LOW:
		t.low = tp.Price
	case CLOSE, DELAYED_CLOSE:
		t.close = tp.Price
	case OPEN, DELAYED_OPEN:
		t.open = tp.Price
	case LOW_13_WEEK:
		t.low13Week = tp.Price
	case HIGH_13_WEEK:
		t.high13Week = tp.Price
	case LOW_26_WEEK:
		t.low26Week = tp.Price
	case HIGH_26_WEEK:
		t.high26Week = tp.Price
	case LOW_52_WEEK:
		t.low52Week = tp.Price
	case HIGH_52_WEEK:
		t.high52Week = tp.Price
	case AUCTION_PRICE:
		t.auctionPrice = tp.Price
	case MARK_PRICE:
		t.markPrice = tp.Price
	case BID_YIELD, DELAYED_YIELD_BID:
		t.bidYield = tp.Price
	case ASK_YIELD, DELAYED_YIELD_ASK:
		t.askYield = tp.Price
	case LAST_YIELD:
		t.lastYield = tp.Price
	default:
		log.Warn().Err(errUnknownTickType).Int64("TickType", tp.TickType).Msg("SetTickPrice")
	}
	td := TickData{
		Time:     time.Now().UTC(),
		TickType: tp.TickType,
		Price:    tp.Price,
		Size:     size,
	}
	t.ticks = append(t.ticks, td)
}

func (t *Ticker) SetTickSize(ts TickSize) {
	t.mu.Lock()
	defer t.mu.Unlock()
	var price float64
	switch ts.TickType {
	case BID_SIZE, DELAYED_BID_SIZE:
		if ts.Size == t.bidSize {
			return
		}
		price = t.bid
		t.prevBidSize = t.bidSize
		t.bidSize = ts.Size
	case ASK_SIZE, DELAYED_ASK_SIZE:
		if ts.Size == t.askSize {
			return
		}
		price = t.ask
		t.prevAskSize = t.askSize
		t.askSize = ts.Size
	case LAST_SIZE, DELAYED_LAST_SIZE:
		price = t.last
		if price == 0 {
			return
		}
		if ts.Size != t.lastSize {
			t.prevLastSize = t.lastSize
			t.lastSize = ts.Size
		}
	case VOLUME, DELAYED_VOLUME:
		t.volume = ts.Size
	case AVG_VOLUME:
		t.avVolume = ts.Size
	case OPTION_CALL_OPEN_INTEREST:
		t.callOpenInterest = ts.Size
	case OPTION_PUT_OPEN_INTEREST:
		t.putOpenInterest = ts.Size
	case OPTION_CALL_VOLUME:
		t.callVolume = ts.Size
	case OPTION_PUT_VOLUME:
		t.putVolume = ts.Size
	case AUCTION_VOLUME:
		t.auctionVolume = ts.Size
	case AUCTION_IMBALANCE:
		t.auctionImbalance = ts.Size
	case REGULATORY_IMBALANCE:
		t.regulatoryImbalance = ts.Size
	case FUTURES_OPEN_INTEREST:
		t.futuresOpenInterest = ts.Size
	case AVG_OPT_VOLUME:
		t.avOptionVolume = ts.Size
	case SHORTABLE_SHARES:
		t.shortableShares = ts.Size
	default:
		log.Warn().Err(errUnknownTickType).Int64("TickType", ts.TickType).Msg("SetTickSize")
	}
	td := TickData{
		Time:     time.Now().UTC(),
		TickType: ts.TickType,
		Price:    price,
		Size:     ts.Size,
	}
	t.ticks = append(t.ticks, td)
}

func (t *Ticker) SetTickOptionComputation(toc TickOptionComputation) {
	t.mu.Lock()
	defer t.mu.Unlock()
	switch toc.TickType {
	case BID_OPTION_COMPUTATION, DELAYED_BID_OPTION:
		t.bidGreeks = toc
	case ASK_OPTION_COMPUTATION, DELAYED_ASK_OPTION:
		t.askGreeks = toc
	case LAST_OPTION_COMPUTATION, DELAYED_LAST_OPTION:
		t.lastGreeks = toc
	case MODEL_OPTION, DELAYED_MODEL_OPTION:
		t.modelGreeks = toc
	default:
		log.Warn().Err(errUnknownTickType).Int64("TickType", toc.TickType).Msg("SetTickOptionComputation")
	}
}

func (t *Ticker) SetTickGeneric(tg TickGeneric) {
	t.mu.Lock()
	defer t.mu.Unlock()
	switch tg.TickType {
	case OPTION_HISTORICAL_VOL:
		t.histVolatility = tg.Value
	case OPTION_IMPLIED_VOL:
		t.impliedVolatility = tg.Value
	case INDEX_FUTURE_PREMIUM:
		t.indexFuturePremium = tg.Value
	case HALTED, DELAYED_HALTED:
		t.halted = tg.Value
	case TRADE_COUNT:
		t.tradeCount = tg.Value
	case TRADE_RATE:
		t.tradeRate = tg.Value
	case VOLUME_RATE:
		t.volumeRate = tg.Value
	case RT_HISTORICAL_VOL:
		t.rtHistVolatility = tg.Value
	default:
		log.Warn().Err(errUnknownTickType).Int64("TickType", tg.TickType).Float64("TickValue", tg.Value).Msg("SetTickGeneric")
	}
	td := TickData{
		Time:     time.Now().UTC(),
		TickType: tg.TickType,
		Price:    tg.Value,
		Size:     ZERO,
	}
	t.ticks = append(t.ticks, td)
}

func (t *Ticker) SetTickString(ts TickString) {
	t.mu.Lock()
	defer t.mu.Unlock()
	switch ts.TickType {
	case BID_EXCH:
		t.bidExchange = ts.Value
	case ASK_EXCH:
		t.askExchange = ts.Value
	case LAST_EXCH:
		t.lastExchange = ts.Value
	case LAST_TIMESTAMP, DELAYED_LAST_TIMESTAMP:
		t.lastTimestamp = ts.Value
	case FUNDAMENTAL_RATIOS:
		d := make(FundamentalRatios)
		for _, t := range strings.Split(ts.Value, ";") {
			if t == "" {
				continue
			}
			kv := strings.Split(t, "=")
			if len(kv) == 2 {
				k, v := kv[0], kv[1]
				if v == "-99999.99" {
					d[k] = UNSET_FLOAT
					continue
				}
				f, err := strconv.ParseFloat(v, 64)
				if err != nil {
					log.Warn().Err(errors.New("fundamental ratio error")).Str("key", k).Str("value", v).Msg("SetTickString")
					continue
				}
				d[k] = f
			}
		}
		t.fundamentalRatios = d
	case RT_VOLUME, RT_TRD_VOLUME:
		// RT Volume or RT Trade Volume value: " price;size;ms since epoch;total volume;VWAP;single trade"
		split := strings.Split(ts.Value, ";")
		if split[3] != "" {
			f, err := strconv.ParseFloat(split[3], 64)
			if err != nil {
				log.Error().Err(err).Msg("<SetTickString>")
			}
			if ts.TickType == RT_VOLUME {
				t.rtVolume = f
			} else {
				t.rtTradeVolume = f
			}
		}
		if split[4] != "" {
			f, err := strconv.ParseFloat(split[4], 64)
			if err != nil {
				log.Error().Err(err).Msg("<SetTickString>")
			}
			t.vwap = f
		}
		if split[2] != "" {
			d, err := ParseIBTime(split[2])
			if err != nil {
				log.Error().Err(err).Msg("<SetTickString>")
			}
			t.rtTime = d
		}
		if split[0] != "" {
			return
		}
		price, err := strconv.ParseFloat(split[0], 64)
		if err != nil {
			log.Error().Err(err).Msg("<SetTickString>")
		}
		if split[1] != "" {
			size := StringToDecimal(split[1])
			if err != nil {
				log.Error().Err(err).Msg("<SetTickString>")
			}
			if t.prevLast != t.last {
				t.prevLast = t.last
				t.last = price
			}
			if t.prevLastSize != t.lastSize {
				t.prevLastSize = t.lastSize
				t.lastSize = size
			}
			td := TickData{
				Time:     time.Now().UTC(),
				TickType: ts.TickType,
				Price:    price,
				Size:     size,
			}
			t.ticks = append(t.ticks, td)
		}
	case IB_DIVIDENDS:
		// Dividend Value: "past12,next12,nextDate,nextAmount"
		split := strings.Split(ts.Value, ",")
		ds := Dividends{}
		if split[0] != "" {
			f, err := strconv.ParseFloat(split[0], 64)
			if err != nil {
				log.Error().Err(err).Msg("<SetTickString>")
			}
			ds.Past12Months = f
		}
		if split[1] != "" {
			f, err := strconv.ParseFloat(split[1], 64)
			if err != nil {
				log.Error().Err(err).Msg("<SetTickString>")
			}
			ds.Next12Months = f
		}
		if split[2] != "" {
			d, err := ParseIBTime(split[2])
			if err != nil {
				log.Error().Err(err).Msg("<SetTickString>")
			}
			ds.NextDate = d
		}
		if split[3] != "" {
			f, err := strconv.ParseFloat(split[3], 64)
			if err != nil {
				log.Error().Err(err).Msg("<SetTickString>")
			}
			ds.NextAmount = f
		}
		t.dividends = ds
	default:
		log.Warn().Err(errUnknownTickType).Int64("TickType", ts.TickType).Msg("SetTickString")
	}
}

func (t *Ticker) SetTickEFP(te TickEFP) {
	t.mu.Lock()
	defer t.mu.Unlock()
	// TODO
}

func (t *Ticker) SetTickByTickAllLast(tbt TickByTickAllLast) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if tbt.Price != t.last {
		t.prevLast = t.last
		t.last = tbt.Price
	}
	if tbt.Size != t.lastSize {
		t.prevLastSize = t.lastSize
		t.lastSize = tbt.Size
	}
	t.tickByTicks = append(t.tickByTicks, tbt)
}

func (t *Ticker) SetTickByTickBidAsk(tbt TickByTickBidAsk) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if tbt.BidPrice != t.bid {
		t.prevBid = t.bid
		t.bid = tbt.BidPrice
	}
	if tbt.BidSize != t.bidSize {
		t.prevBidSize = t.bidSize
		t.bidSize = tbt.BidSize
	}
	if tbt.AskPrice != t.ask {
		t.prevAsk = t.ask
		t.ask = tbt.AskPrice
	}
	if tbt.AskSize != t.askSize {
		t.prevAskSize = t.askSize
		t.askSize = tbt.AskSize
	}
	t.tickByTicks = append(t.tickByTicks, tbt)
}

func (t *Ticker) SetTickByTickMidPoint(tbt TickByTickMidPoint) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.tickByTicks = append(t.tickByTicks, tbt)
}

func (t *Ticker) String() string {
	return Stringify(struct {
		Contract            *Contract
		Time                time.Time
		MarketDataType      int64
		MinTick             float64
		Bid                 float64
		BidSize             Decimal
		BidExchange         string
		Ask                 float64
		AskSize             Decimal
		AskExchange         string
		Last                float64
		LastSize            Decimal
		LastExchange        string
		LastTimestamp       string
		PrevBid             float64
		PrevBidSize         Decimal
		PrevAsk             float64
		PrevAskSize         Decimal
		PrevLast            float64
		PrevLastSize        Decimal
		Volume              Decimal
		Open                float64
		High                float64
		Low                 float64
		Close               float64
		Vwap                float64
		Low13Week           float64
		High13Week          float64
		Low26Week           float64
		High26Week          float64
		Low52Week           float64
		High52Week          float64
		BidYield            float64
		AskYield            float64
		LastYield           float64
		MarkPrice           float64
		Halted              float64
		RtHistVolatility    float64
		RtVolume            float64
		RtTradeVolume       float64
		RtTime              time.Time
		AvVolume            Decimal
		TradeCount          float64
		TradeRate           float64
		VolumeRate          float64
		ShortableShares     Decimal
		IndexFuturePremium  float64
		FuturesOpenInterest Decimal
		PutOpenInterest     Decimal
		CallOpenInterest    Decimal
		PutVolume           Decimal
		CallVolume          Decimal
		AvOptionVolume      Decimal
		HistVolatility      float64
		ImpliedVolatility   float64
		Dividends           Dividends
		FundamentalRatios   FundamentalRatios
		Ticks               []TickData
		TickByTicks         []TickByTick
		DomBids             map[int64]DOMLevel
		DomAsks             map[int64]DOMLevel
		DomTicks            []MktDepthData
		BidGreeks           TickOptionComputation
		AskGreeks           TickOptionComputation
		LastGreeks          TickOptionComputation
		ModelGreeks         TickOptionComputation
		AuctionVolume       Decimal
		AuctionPrice        float64
		AuctionImbalance    Decimal
		RegulatoryImbalance Decimal
		BboExchange         string
		SnapshotPermissions int64
	}{
		Contract:            t.contract,
		Time:                t.time,
		MarketDataType:      t.marketDataType,
		MinTick:             t.minTick,
		Bid:                 t.bid,
		BidSize:             t.bidSize,
		BidExchange:         t.bidExchange,
		Ask:                 t.ask,
		AskSize:             t.askSize,
		AskExchange:         t.askExchange,
		Last:                t.last,
		LastSize:            t.lastSize,
		LastExchange:        t.lastExchange,
		LastTimestamp:       t.lastTimestamp,
		PrevBid:             t.prevBid,
		PrevBidSize:         t.prevBidSize,
		PrevAsk:             t.prevAsk,
		PrevAskSize:         t.prevAskSize,
		PrevLast:            t.prevLast,
		PrevLastSize:        t.prevLastSize,
		Volume:              t.volume,
		Open:                t.open,
		High:                t.high,
		Low:                 t.low,
		Close:               t.close,
		Vwap:                t.vwap,
		Low13Week:           t.low13Week,
		High13Week:          t.high13Week,
		Low26Week:           t.low26Week,
		High26Week:          t.high26Week,
		Low52Week:           t.low52Week,
		High52Week:          t.high52Week,
		BidYield:            t.bidYield,
		AskYield:            t.askYield,
		LastYield:           t.lastYield,
		MarkPrice:           t.markPrice,
		Halted:              t.halted,
		RtHistVolatility:    t.rtHistVolatility,
		RtVolume:            t.rtVolume,
		RtTradeVolume:       t.rtTradeVolume,
		RtTime:              t.rtTime,
		AvVolume:            t.avVolume,
		TradeCount:          t.tradeCount,
		TradeRate:           t.tradeRate,
		VolumeRate:          t.volumeRate,
		ShortableShares:     t.shortableShares,
		IndexFuturePremium:  t.indexFuturePremium,
		FuturesOpenInterest: t.futuresOpenInterest,
		PutOpenInterest:     t.putOpenInterest,
		CallOpenInterest:    t.callOpenInterest,
		PutVolume:           t.putVolume,
		CallVolume:          t.callVolume,
		AvOptionVolume:      t.avOptionVolume,
		HistVolatility:      t.histVolatility,
		ImpliedVolatility:   t.impliedVolatility,
		Dividends:           t.dividends,
		FundamentalRatios:   t.fundamentalRatios,
		Ticks:               t.ticks,
		TickByTicks:         t.tickByTicks,
		DomBids:             t.domBids,
		DomAsks:             t.domAsks,
		DomTicks:            t.domTicks,
		BidGreeks:           t.bidGreeks,
		AskGreeks:           t.askGreeks,
		LastGreeks:          t.lastGreeks,
		ModelGreeks:         t.modelGreeks,
		AuctionVolume:       t.auctionVolume,
		AuctionPrice:        t.auctionPrice,
		AuctionImbalance:    t.auctionImbalance,
		RegulatoryImbalance: t.regulatoryImbalance,
		BboExchange:         t.bboExchange,
		SnapshotPermissions: t.snapshotPermissions,
	})
}
