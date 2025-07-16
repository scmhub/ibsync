package ibsync

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/scmhub/ibapi"
	"github.com/scmhub/ibapi/protobuf"
)

var _ ibapi.EWrapper = (*WrapperSync)(nil)

type WrapperSync struct {
	state  *ibState
	pubSub *PubSub
}

// NewWrapperSync implements the ibapi EWrapper
func NewWrapperSync(state *ibState, pubSub *PubSub) *WrapperSync {
	return &WrapperSync{
		state:  state,
		pubSub: pubSub,
	}
}

func (w *WrapperSync) TickPrice(reqID TickerID, tickType TickType, price float64, attrib TickAttrib) {
	log.Debug().Int64("reqID", reqID).Int64("tickType", tickType).Str("tickName", TickName(tickType)).Str("price", FloatMaxString(price)).Bool("CanAutoExecute", attrib.CanAutoExecute).Bool("PastLimit", attrib.PastLimit).Bool("PreOpen", attrib.PreOpen).Msg("<TickPrice>")
	tickPrice := TickPrice{TickType: tickType, Price: price, Attrib: attrib}

	w.state.mu.Lock()
	ticker := w.state.reqID2Ticker[reqID]
	w.state.mu.Unlock()

	ticker.SetTickPrice(tickPrice)
	//w.pubSub.Publish(reqID, Join("price", Encode(tickPrice)))
}

func (w *WrapperSync) TickSize(reqID TickerID, tickType TickType, size Decimal) {
	log.Debug().Int64("reqID", reqID).Int64("tickType", tickType).Str("tickName", TickName(tickType)).Str("size", DecimalMaxString(size)).Msg("<TickSize>")
	tickSize := TickSize{TickType: tickType, Size: size}

	w.state.mu.Lock()
	ticker := w.state.reqID2Ticker[reqID]
	w.state.mu.Unlock()

	ticker.SetTickSize(tickSize)
	//w.pubSub.Publish(reqID, Join("size", Encode(tickSize)))
}

func (w *WrapperSync) TickOptionComputation(reqID TickerID, tickType TickType, tickAttrib int64, impliedVol float64, delta float64, optPrice float64, pvDividend float64, gamma float64, vega float64, theta float64, undPrice float64) {
	log.Debug().Int64("reqID", reqID).Int64("tickType", tickType).Str("tickName", TickName(tickType)).Str("tickAttrib", IntMaxString(tickAttrib)).Str("impliedVol", FloatMaxString(impliedVol)).Str("delta", FloatMaxString(delta)).Str("optPrice", FloatMaxString(optPrice)).Str("pvDividend", FloatMaxString(pvDividend)).Str("gamma", FloatMaxString(gamma)).Str("vega", FloatMaxString(vega)).Str("theta", FloatMaxString(theta)).Str("undPrice", FloatMaxString(undPrice)).Msg("<TickOptionComputation>")
	tickOptionComputation := TickOptionComputation{TickType: tickType, TickAttrib: tickAttrib, ImpliedVol: impliedVol, Delta: delta, OptPrice: optPrice, PvDividend: pvDividend, Gamma: gamma, Vega: vega, Theta: theta, UndPrice: undPrice}

	w.state.mu.Lock()
	ticker, ok := w.state.reqID2Ticker[reqID]
	w.state.mu.Unlock()

	if ok {
		ticker.SetTickOptionComputation(tickOptionComputation)
		return
	}

	w.pubSub.Publish(reqID, Join("OptionComputation", Encode(tickOptionComputation)))
}

func (w *WrapperSync) TickGeneric(reqID TickerID, tickType TickType, value float64) {
	log.Debug().Int64("reqID", reqID).Int64("tickType", tickType).Str("value", FloatMaxString(value)).Msg("<TickGeneric>")
	tickGeneric := TickGeneric{TickType: tickType, Value: value}

	w.state.mu.Lock()
	ticker := w.state.reqID2Ticker[reqID]
	w.state.mu.Unlock()

	ticker.SetTickGeneric(tickGeneric)
	//w.pubSub.Publish(reqID, Join("generic", Encode(tickGeneric)))
}

func (w *WrapperSync) TickString(reqID TickerID, tickType TickType, value string) {
	log.Debug().Int64("reqID", reqID).Int64("tickType", tickType).Str("value", value).Msg("<TickString>")
	tickString := TickString{TickType: tickType, Value: value}

	w.state.mu.Lock()
	ticker := w.state.reqID2Ticker[reqID]
	w.state.mu.Unlock()

	ticker.SetTickString(tickString)
	//w.pubSub.Publish(reqID, Join("string", Encode(tickString)))
}

func (w *WrapperSync) TickEFP(reqID TickerID, tickType TickType, basisPoints float64, formattedBasisPoints string, totalDividends float64, holdDays int64, futureLastTradeDate string, dividendImpact float64, dividendsToLastTradeDate float64) {
	log.Debug().Int64("reqID", reqID).Int64("tickType", tickType).Float64("basisPoints", basisPoints).Str("formattedBasisPoints", formattedBasisPoints).Float64("totalDividends", totalDividends).Int64("holdDays", holdDays).Str("futureLastTradeDate", futureLastTradeDate).Float64("dividendImpact", dividendImpact).Float64("dividendsToLastTradeDate", dividendsToLastTradeDate).Msg("<TickEFP>")
	tickEFP := TickEFP{TickType: tickType, BasisPoints: basisPoints, FormattedBasisPoints: formattedBasisPoints, TotalDividends: totalDividends, HoldDays: holdDays, FutureLastTradeDate: futureLastTradeDate, DividendImpact: dividendImpact, DividendsToLastTradeDate: dividendsToLastTradeDate}

	w.state.mu.Lock()
	ticker := w.state.reqID2Ticker[reqID]
	w.state.mu.Unlock()

	ticker.SetTickEFP(tickEFP)
	//w.pubSub.Publish(reqID, Join("efp", Encode(tickEFP)))
}

func (w *WrapperSync) OrderStatus(orderID OrderID, status string, filled Decimal, remaining Decimal, avgFillPrice float64, permID int64, parentID int64, lastFillPrice float64, clientID int64, whyHeld string, mktCapPrice float64) {
	log.Debug().Int64("orderID", orderID).Str("status", status).Stringer("filled", filled).Stringer("remaining", remaining).Float64("avgFillPrice", avgFillPrice).Int64("permID", permID).Int64("parentID", parentID).Float64("lastFillPrice", lastFillPrice).Int64("clientID", clientID).Str("whyHeld", whyHeld).Float64("mktCapPrice", mktCapPrice).Msg("<OrderStatus>")
	orderStatus := OrderStatus{OrderID: orderID, Status: Status(status), Filled: filled, Remaining: remaining, AvgFillPrice: avgFillPrice, PermID: permID, ParentID: parentID, LastFillPrice: lastFillPrice, ClientID: clientID, WhyHeld: whyHeld, MktCapPrice: mktCapPrice}
	key := orderKey(clientID, orderID, permID)

	w.state.mu.Lock()
	trade, ok := w.state.trades[key]
	w.state.mu.Unlock()

	if ok {
		trade.mu.Lock()
		oldStatus := trade.OrderStatus.Status
		trade.OrderStatus = orderStatus

		var msg string
		if Status(status) == Submitted && len(trade.logs) > 0 && trade.logs[len(trade.logs)-1].Message == "Modify" {
			msg = "Modified"
		}
		if msg != "" || orderStatus.Status != oldStatus {
			logEntry := TradeLogEntry{Time: time.Now().UTC().Truncate(time.Minute), Status: Status(status), Message: msg}
			trade.addLog(logEntry)
		}
		if Status(status).IsDone() {
			trade.markDone()
		}
		trade.mu.Unlock()
	} else {
		log.Error().Err(errUnknowOrder).Int64("OrderID", orderID).Int64("ClientID", clientID).Msg("<OrderStatus>")
	}
}

func (w *WrapperSync) OpenOrder(orderID OrderID, contract *Contract, order *Order, orderState *OrderState) {
	log.Debug().Int64("orderID", orderID).Stringer("contract", contract).Stringer("order", order).Stringer("orderState", orderState).Msg("<OpenOrder>")
	key := orderKey(order.ClientID, order.OrderID, order.PermID)
	status := Status(orderState.Status)

	w.state.mu.Lock()
	defer w.state.mu.Unlock()
	trade, ok := w.state.trades[key]

	if ok {
		// Update the existing trade object fields
		trade.mu.Lock()
		trade.Order.PermID = order.PermID
		trade.Order.TotalQuantity = order.TotalQuantity
		trade.Order.LmtPrice = order.LmtPrice
		trade.Order.AuxPrice = order.AuxPrice
		trade.Order.OrderType = order.OrderType
		trade.Order.OrderRef = order.OrderRef
		trade.OrderStatus.Status = status
		trade.mu.Unlock()
	} else {
		// Create a new trade if not found
		orderStatus := OrderStatus{
			OrderID: orderID,
			Status:  status,
		}
		trade = NewTrade(contract, order, orderStatus)
		w.state.trades[key] = trade
		w.state.permID2Trade[trade.Order.PermID] = trade
	}
	if status.IsDone() {
		trade.markDoneSafe()
	}
	// make sure that the client issues order ids larger than any
	// order id encountered (even from other clients) to avoid
	// "Duplicate order id" error
	w.state.updateID(orderID + 1)
}

func (w *WrapperSync) OpenOrderEnd() {
	log.Debug().Msg("<OpenOrderEnd>")
	w.pubSub.Publish("OpenOrdersEnd", "")
}

func (w *WrapperSync) WinError(text string, lastError int64) {
	log.Warn().Str("text", text).Int64("lastError", lastError).Msg("<WinError>")
}

func (w *WrapperSync) ConnectionClosed() {
	log.Warn().Msg("<ConnectionClosed>...")
}

func (w *WrapperSync) UpdateAccountValue(tag string, value string, currency string, accountName string) {
	log.Debug().Str("tag", tag).Str("value", value).Str("currency", currency).Str("accountName", accountName).Msg("<UpdateAccountValue>")
	av := AccountValue{Account: accountName, Tag: tag, Value: value, Currency: currency}
	w.state.mu.Lock()
	defer w.state.mu.Unlock()
	w.state.updateAccountValues[Key(accountName, tag, currency)] = av
}

func (w *WrapperSync) UpdatePortfolio(contract *Contract, position Decimal, marketPrice float64, marketValue float64, averageCost float64, unrealizedPNL float64, realizedPNL float64, accountName string) {
	log.Debug().Str("Symbol", contract.Symbol).Str("secType", contract.SecType).Str("exchange", contract.Exchange).Discard().Str("position", DecimalMaxString(position)).Str("marketPrice", FloatMaxString(marketPrice)).Str("marketValue", FloatMaxString(marketValue)).Str("averageCost", FloatMaxString(averageCost)).Str("unrealizedPNL", FloatMaxString(unrealizedPNL)).Str("realizedPNL", FloatMaxString(realizedPNL)).Str("accountName", accountName).Msg("<UpdatePortfolio>")
	pi := PortfolioItem{Contract: contract, Position: position, MarketPrice: marketPrice, MarketValue: marketValue, AverageCost: averageCost, UnrealizedPNL: unrealizedPNL, RealizedPNL: realizedPNL, Account: accountName}
	w.state.mu.Lock()
	defer w.state.mu.Unlock()
	portfolioItems, ok := w.state.portfolio[accountName]
	if !ok {
		portfolioItems = make(map[int64]PortfolioItem)
		w.state.portfolio[accountName] = portfolioItems
	}
	if pi.Position == ZERO {
		delete(portfolioItems, pi.Contract.ConID)
	} else {
		portfolioItems[pi.Contract.ConID] = pi
	}
}

func (w *WrapperSync) UpdateAccountTime(timeStamp string) {
	log.Debug().Str("timeStamp", timeStamp).Msg("<UpdateAccountTime>")
	t, err := time.Parse("15:04", timeStamp)
	if err != nil {
		log.Error().Err(err).Msg("<UpdateAccountTime>")
	}
	w.state.mu.Lock()
	defer w.state.mu.Unlock()
	w.state.updateAccountTime = t
	w.pubSub.Publish("UpdateAccountTime", timeStamp)
}

func (w *WrapperSync) AccountDownloadEnd(accountName string) {
	log.Debug().Str("accountName", accountName).Msg("<AccountDownloadEnd>")
	w.pubSub.Publish("AccountDownloadEnd", accountName)
}

func (w *WrapperSync) NextValidID(reqID int64) {
	log.Debug().Int64("reqID", reqID).Msg("<NextValidID>")
	w.state.mu.Lock()
	defer w.state.mu.Unlock()
	if reqID > w.state.nextValidID {
		w.state.nextValidID = reqID
	}
	w.pubSub.Publish("NextValidID", Encode(reqID))
}

func (w *WrapperSync) ContractDetails(reqID int64, contractDetails *ContractDetails) {
	log.Debug().Int64("reqID", reqID).Stringer("contractDetails", contractDetails).Msg("<ContractDetails>")
	w.pubSub.Publish(reqID, Encode(contractDetails))
}

func (w *WrapperSync) BondContractDetails(reqID int64, contractDetails *ContractDetails) {
	log.Debug().Int64("reqID", reqID).Stringer("contractDetails", contractDetails).Msg("<BondContractDetails>")
	w.pubSub.Publish(reqID, Encode(contractDetails))
}

func (w *WrapperSync) ContractDetailsEnd(reqID int64) {
	log.Debug().Int64("reqID", reqID).Msg("<ContractDetailsEnd>")
	w.pubSub.Publish(reqID, "end")
}

func (w *WrapperSync) ExecDetails(reqID int64, contract *Contract, execution *Execution) {
	log.Debug().Int64("reqID", reqID).Stringer("contract", contract).Stringer("executioncontract", execution).Msg("<ExecDetails>")
	if execution.OrderID == UNSET_INT {
		execution.OrderID = 0
	}
	w.state.mu.Lock()
	trade, ok := w.state.permID2Trade[execution.PermID]
	if !ok {
		key := orderKey(execution.ClientID, execution.OrderID, execution.PermID)
		trade, ok = w.state.trades[key]
		if !ok {
			w.state.trades[strconv.FormatInt(execution.PermID, 10)] = trade
			trade = NewTrade(contract, nil, OrderStatus{OrderID: execution.OrderID})
			w.state.permID2Trade[execution.PermID] = trade
		}
	}
	executionTime, err := ParseIBTime(execution.Time)
	if err != nil {
		log.Error().Err(err).Int64("reqID", reqID).Msg("<ExecDetails>")
		return
	}
	fill := &Fill{
		Contract:                contract,
		Execution:               execution,
		CommissionAndFeesReport: NewCommissionAndFeesReport(),
		Time:                    executionTime,
	}
	_, ok = w.state.fills[execution.ExecID]
	if !ok {
		w.state.fills[execution.ExecID] = fill
		trade.addFill(fill)
		logEntry := TradeLogEntry{
			Time:    executionTime,
			Status:  trade.OrderStatus.Status,
			Message: fmt.Sprintf("Fill %v@%v", execution.Shares, execution.Price),
		}
		trade.addLog(logEntry)
	}
	w.state.mu.Unlock()

	w.pubSub.Publish(reqID, (Encode(fill)))
}

func (w *WrapperSync) ExecDetailsEnd(reqID int64) {
	log.Debug().Int64("reqID", reqID).Msg("<ExecDetailsEnd>")
	w.pubSub.Publish(reqID, "end")
}

func (w *WrapperSync) Error(reqID TickerID, errorTime int64, errCode int64, errString string, advancedOrderRejectJson string) {
	logger := log.Error()
	if slices.Contains(warningCodes, errCode) { //|| (2100 <= errCode && errCode < 2200) {
		logger = log.Warn()
	}
	logger.Int64("reqID", reqID).Int64("errorTime", errorTime).Int64("errCode", errCode).Str("errString", errString)
	if advancedOrderRejectJson != "" {
		logger = logger.Str("advancedOrderRejectJson", advancedOrderRejectJson)
		errString = errString + " " + advancedOrderRejectJson
	}
	logger.Msg("<Error>")

	w.pubSub.Publish(reqID, Join("error", Encode(ibapi.CodeMsgPair{Code: errCode, Msg: errString})))
}

func (w *WrapperSync) UpdateMktDepth(TickerID TickerID, position int64, operation int64, side int64, price float64, size Decimal) {
	log.Debug().Int64("TickerID", TickerID).Int64("position", position).Int64("operation", operation).Int64("side", side).Str("price", FloatMaxString(price)).Str("size", DecimalMaxString(size)).Msg("<UpdateMktDepth>")
	w.updateMktDepth(TickerID, position, "", operation, side, price, size, false)
}

func (w *WrapperSync) UpdateMktDepthL2(TickerID TickerID, position int64, marketMaker string, operation int64, side int64, price float64, size Decimal, isSmartDepth bool) {
	log.Debug().Int64("TickerID", TickerID).Int64("position", position).Str("marketMaker", marketMaker).Int64("operation", operation).Int64("side", side).Str("price", FloatMaxString(price)).Str("size", DecimalMaxString(size)).Bool("isSmartDepth", isSmartDepth).Msg("<UpdateMktDepthL2>")
	w.updateMktDepth(TickerID, position, marketMaker, operation, side, price, size, isSmartDepth)
}
func (w *WrapperSync) updateMktDepth(TickerID TickerID, position int64, marketMaker string, operation int64, side int64, price float64, size Decimal, isSmartDepth bool) {
	w.state.mu.Lock()
	defer w.state.mu.Unlock()
	ticker := w.state.reqID2Ticker[TickerID]

	// side: 0 = ask, 1 = bid
	var dom map[int64]DOMLevel
	switch side {
	case 0:
		dom = ticker.domAsks
	case 1:
		dom = ticker.domBids
	default:
		log.Error().Err(errors.New("unknown DOM side")).Msg("updateMktDepth")
		return
	}
	// operation: 0 = insert, 1 = update, 2 = delete
	switch operation {
	case 0, 1:
		dom[position] = DOMLevel{Price: price, Size: size, MarketMaker: marketMaker}
	case 2:
		delete(dom, position)
	default:
		log.Error().Err(errors.New("unknown DOM operation")).Msg("updateMktDepth>")
		return
	}

	tick := MktDepthData{Time: time.Now(), Position: position, MarketMaker: marketMaker, Operation: operation, Side: side, Price: price, Size: size, IsSmartDepth: isSmartDepth}
	ticker.domTicks = append(ticker.domTicks, tick)
	w.pubSub.Publish(TickerID, "ok")
}

func (w *WrapperSync) UpdateNewsBulletin(msgID int64, msgType int64, newsMessage string, originExch string) {
	log.Debug().Int64("msgID", msgID).Int64("msgType", msgType).Str("newsMessage", newsMessage).Str("originExch", originExch).Msg("<UpdateNewsBulletin>")
	newsBulletin := NewsBulletin{MsgID: msgID, MsgType: msgType, NewsMessage: newsMessage, OriginExch: originExch}
	w.state.mu.Lock()
	w.state.msgID2NewsBulletin[msgID] = newsBulletin
	w.state.mu.Unlock()
	w.pubSub.Publish("NewsBulletin", Encode(newsBulletin))
}

func (w *WrapperSync) ManagedAccounts(accountsList []string) {
	log.Debug().Strs("accountsList", accountsList).Msg("<ManagedAccounts>")
	w.state.mu.Lock()
	defer w.state.mu.Unlock()
	w.state.accounts = accountsList
	w.pubSub.Publish("ManagedAccounts", Join(accountsList...))
}

func (w *WrapperSync) ReceiveFA(faDataType FaDataType, cxml string) {
	log.Debug().Stringer("faDataType", faDataType).Str("cxml", cxml).Msg("<ReceiveFA>")
	receiveFA := ReceiveFA{FaDataType: faDataType, Cxml: cxml}
	w.pubSub.Publish("ReceiveFA", Encode(receiveFA))
}

func (w *WrapperSync) HistoricalData(reqID int64, bar *Bar) {
	log.Debug().Int64("reqID", reqID).Stringer("bar", bar).Msg("<HistoricalData>")
	w.pubSub.Publish(reqID, Join("HistoricalData", Encode(bar)))
}

func (w *WrapperSync) HistoricalDataEnd(reqID int64, startDateStr string, endDateStr string) {
	log.Debug().Int64("reqID", reqID).Str("startDateStr", startDateStr).Str("endDateStr", endDateStr).Msg("<HistoricalDataEnd>")
	w.pubSub.Publish(reqID, Join("HistoricalDataEnd", startDateStr, endDateStr))
}

func (w *WrapperSync) ScannerParameters(xml string) {
	log.Debug().Str("xml", xml[:50]).Msg("<ScannerParameters>")
	w.pubSub.Publish("ScannerParameters", xml)
}

func (w *WrapperSync) ScannerData(reqID int64, rank int64, contractDetails *ContractDetails, distance string, benchmark string, projection string, legsStr string) {
	log.Debug().Int64("reqID", reqID).Int64("rank", rank).Stringer("contractDetails", contractDetails).Str("distance", distance).Str("benchmark", benchmark).Str("projection", projection).Str("legsStr", legsStr).Msg("<ScannerData>")
	sd := ScanData{Rank: rank, ContractDetails: contractDetails, Distance: distance, Benchmark: benchmark, Projection: projection, LegsStr: legsStr}
	w.pubSub.Publish(reqID, Encode(sd))
}

func (w *WrapperSync) ScannerDataEnd(reqID int64) {
	log.Debug().Int64("reqID", reqID).Msg("<ScannerDataEnd>")
	w.pubSub.Publish(reqID, "end")
}

func (w *WrapperSync) RealtimeBar(reqID int64, time int64, open float64, high float64, low float64, close float64, volume Decimal, wap Decimal, count int64) {
	log.Debug().Int64("reqID", reqID).Int64("bar time", time).Float64("open", open).Float64("high", high).Float64("low", low).Float64("close", close).Stringer("volume", volume).Stringer("wap", wap).Int64("count", count).Msg("<RealtimeBar>")
	rtb := RealTimeBar{Time: time, Open: open, High: high, Low: low, Close: close, Volume: volume, Wap: wap, Count: count}
	w.pubSub.Publish(reqID, Encode(rtb))
}

func (w *WrapperSync) CurrentTime(t int64) {
	currentTime := time.Unix(t, 0)
	log.Debug().Time("Server Time", currentTime).Msg("<CurrentTime>")
	w.pubSub.Publish("CurrentTime", Encode(currentTime))
}

func (w *WrapperSync) FundamentalData(reqID int64, data string) {
	log.Debug().Int64("reqID", reqID).Str("data", data).Msg("<FundamentalData>")
	w.pubSub.Publish(reqID, data)
}

func (w *WrapperSync) DeltaNeutralValidation(reqID int64, deltaNeutralContract DeltaNeutralContract) {
	log.Debug().Int64("reqID", reqID).Stringer("deltaNeutralContract", deltaNeutralContract).Msg("<DeltaNeutralValidation>")
	w.pubSub.Publish(reqID, Encode(deltaNeutralContract))
}

func (w *WrapperSync) TickSnapshotEnd(reqID int64) {
	log.Debug().Int64("reqID", reqID).Msg("<TickSnapshotEnd>")
	w.pubSub.Publish(reqID, "TickSnapshotEnd")
}

func (w *WrapperSync) MarketDataType(reqID int64, marketDataType int64) {
	log.Debug().Int64("reqID", reqID).Int64("marketDataType", marketDataType).Msg("<MarketDataType>")
	w.state.mu.Lock()
	defer w.state.mu.Unlock()
	ticker, ok := w.state.reqID2Ticker[reqID]
	if ok {
		ticker.setMarketDataType(marketDataType)
	}
}

func (w *WrapperSync) CommissionAndFeesReport(commissionAndFeesReport CommissionAndFeesReport) {
	if commissionAndFeesReport.Yield == UNSET_FLOAT {
		commissionAndFeesReport.Yield = 0.0
	}
	if commissionAndFeesReport.RealizedPNL == UNSET_FLOAT {
		commissionAndFeesReport.RealizedPNL = 0.0
	}
	log.Debug().Stringer("commissionAndFeesReport", commissionAndFeesReport).Msg("<CommissionAndFeesReport>")

	w.state.mu.Lock()
	fill, ok := w.state.fills[commissionAndFeesReport.ExecID]
	if !ok {
		log.Error().Err(errUnknowExecution).Stringer("commissionReportAndFees", commissionAndFeesReport).Msg("<CommissionReportAndFeesœ		>")
		return
	}
	w.state.mu.Unlock()

	fill.CommissionAndFeesReport = commissionAndFeesReport

}

func (w *WrapperSync) Position(account string, contract *Contract, position Decimal, avgCost float64) {
	log.Debug().Str("account", account).Stringer("contract", contract).Str("position", DecimalMaxString(position)).Str("avgCost", FloatMaxString(avgCost)).Msg("<Position>")
	p := Position{Account: account, Contract: contract, Position: position, AvgCost: avgCost}
	w.state.mu.Lock()
	defer w.state.mu.Unlock()
	positions, ok := w.state.positions[p.Account]
	if !ok {
		positions = make(map[int64]Position)
		w.state.positions[p.Account] = positions
	}
	if p.Position == ZERO {
		delete(positions, p.Contract.ConID)
	} else {
		positions[p.Contract.ConID] = p
	}
	w.pubSub.Publish("Position", Encode(p))
}

func (w *WrapperSync) PositionEnd() {
	log.Debug().Msg("<PositionEnd>")
	w.pubSub.Publish("PositionEnd", "")
}

func (w *WrapperSync) AccountSummary(reqID int64, account string, tag string, value string, currency string) {
	log.Debug().Int64("reqID", reqID).Str("account", account).Str("tag", tag).Str("value", value).Str("currency", currency).Msg("<AccountSummary>")
	av := AccountValue{Account: account, Tag: tag, Value: value, Currency: currency}

	w.state.mu.Lock()
	w.state.accountSummary[Key(account, tag, currency)] = av
	w.state.mu.Unlock()

	w.pubSub.Publish(reqID, Encode(av))
}

func (w *WrapperSync) AccountSummaryEnd(reqID int64) {
	log.Debug().Int64("reqID", reqID).Msg("<AccountSummaryEnd>")
	w.pubSub.Publish(reqID, "end")
}

func (w *WrapperSync) VerifyMessageAPI(apiData string) {
	log.Warn().Str("apiData", apiData).Msg("<VerifyMessageAPI>")
}

func (w *WrapperSync) VerifyCompleted(isSuccessful bool, errorText string) {
	log.Warn().Bool("isSuccessful", isSuccessful).Str("errorText", errorText).Msg("<VerifyCompleted>")
}

func (w *WrapperSync) DisplayGroupList(reqID int64, groups string) {
	log.Debug().Int64("reqID", reqID).Str("groups", groups).Msg("<DisplayGroupList>")
	w.pubSub.Publish(reqID, groups)
}

func (w *WrapperSync) DisplayGroupUpdated(reqID int64, contractInfo string) {
	log.Debug().Int64("reqID", reqID).Str("contractInfo", contractInfo).Msg("<DisplayGroupUpdated>")
	w.pubSub.Publish(reqID, contractInfo)
}

func (w *WrapperSync) VerifyAndAuthMessageAPI(apiData string, xyzChallange string) {
	log.Warn().Str("apiData", apiData).Str("xyzChallange", xyzChallange).Msg("<VerifyAndAuthMessageAPI>")
}

func (w *WrapperSync) VerifyAndAuthCompleted(isSuccessful bool, errorText string) {
	log.Warn().Bool("isSuccessful", isSuccessful).Str("errorText", errorText).Msg("<VerifyAndAuthCompleted>")
}

func (w *WrapperSync) ConnectAck() {
	log.Debug().Msg("<ConnectAck>...")
	w.pubSub.Publish("ConnectAck", "")
}

func (w *WrapperSync) PositionMulti(reqID int64, account string, modelCode string, contract *Contract, pos Decimal, avgCost float64) {
	log.Debug().Int64("reqID", reqID).Str("account", account).Str("modelCode", modelCode).Stringer("contract", contract).Str("position", DecimalMaxString(pos)).Str("avgCost", FloatMaxString(avgCost)).Msg("<PositionMulti>")
	w.pubSub.Publish(reqID, Join(account, modelCode, Encode(contract), pos.String(), Encode(avgCost)))
}

func (w *WrapperSync) PositionMultiEnd(reqID int64) {
	log.Debug().Int64("reqID", reqID).Msg("<PositionMultiEnd>")
	w.pubSub.Publish(reqID, "end")
}

func (w *WrapperSync) AccountUpdateMulti(reqID int64, account string, modelCode string, key string, value string, currency string) {
	log.Debug().Int64("reqID", reqID).Str("account", account).Str("modelCode", modelCode).Str("key", key).Str("value", value).Str("currency", currency).Msg("<AccountUpdateMulti>")
	w.pubSub.Publish(reqID, Join(account, modelCode, key, value, currency))
}

func (w *WrapperSync) AccountUpdateMultiEnd(reqID int64) {
	log.Debug().Int64("reqID", reqID).Msg("<AccountUpdateMultiEnd>")
	w.pubSub.Publish(reqID, "end")
}

func (w *WrapperSync) SecurityDefinitionOptionParameter(reqID int64, exchange string, underlyingConID int64, tradingClass string, multiplier string, expirations []string, strikes []float64) {
	log.Debug().Int64("reqID", reqID).Str("exchange", exchange).Str("underlyingConID", IntMaxString(underlyingConID)).Str("tradingClass", tradingClass).Str("multiplier", multiplier).Strs("expirations", expirations).Floats64("strikes", strikes).Msg("<SecurityDefinitionOptionParameter>")
	optionChain := OptionChain{Exchange: exchange, UnderlyingConId: underlyingConID, TradingClass: tradingClass, Multiplier: multiplier, Expirations: expirations, Strikes: strikes}
	w.pubSub.Publish(reqID, Encode(optionChain))
}

func (w *WrapperSync) SecurityDefinitionOptionParameterEnd(reqID int64) {
	log.Debug().Int64("reqID", reqID).Msg("<SecurityDefinitionOptionParameterEnd>")
	w.pubSub.Publish(reqID, "end")
}

func (w *WrapperSync) SoftDollarTiers(reqID int64, tiers []SoftDollarTier) {
	for _, sdt := range tiers {
		log.Debug().Int64("reqID", reqID).Stringer("softDollarTier", sdt).Msg("<SoftDollarTiers>")
	}
	w.pubSub.Publish(reqID, Encode(tiers))
}

func (w *WrapperSync) FamilyCodes(familyCodes []FamilyCode) {
	for _, fc := range familyCodes {
		log.Debug().Stringer("familyCode", fc).Msg("<FamilyCodes>")
	}
	w.pubSub.Publish("FamilyCodes", Encode(familyCodes))
}

func (w *WrapperSync) SymbolSamples(reqID int64, contractDescriptions []ContractDescription) {
	log.Debug().Int("nb_samples", len(contractDescriptions)).Int64("reqID", reqID).Msg("<SymbolSamples>")
	for i, cd := range contractDescriptions {
		log.Debug().Stringer("contract", cd.Contract).Msgf("<Sample %v>", i)
	}
	w.pubSub.Publish(reqID, Encode(contractDescriptions))
}

func (w *WrapperSync) MktDepthExchanges(depthMktDataDescriptions []DepthMktDataDescription) {
	log.Debug().Any("depthMktDataDescriptions", depthMktDataDescriptions).Msg("<MktDepthExchanges>")
	w.pubSub.Publish("MktDepthExchanges", Encode(depthMktDataDescriptions))
}

func (w *WrapperSync) TickNews(TickerID TickerID, timeStamp int64, providerCode string, articleID string, headline string, extraData string) {
	log.Debug().Int64("TickerID", TickerID).Str("timeStamp", IntMaxString(timeStamp)).Str("providerCode", providerCode).Str("articleID", articleID).Str("headline", headline).Str("extraData", extraData).Msg("<TickNews>")
	newsTick := NewsTick{TimeStamp: timeStamp, ProviderCode: providerCode, ArticleId: articleID, Headline: headline, ExtraData: extraData}

	w.state.mu.Lock()
	w.state.newsTicks = append(w.state.newsTicks, newsTick)
	w.state.mu.Unlock()

	w.pubSub.Publish(TickerID, Encode(newsTick))
}

func (w *WrapperSync) SmartComponents(reqID int64, smartComponents []SmartComponent) {
	log.Debug().Int64("reqID", reqID).Msg("<SmartComponents>")
	for i, sc := range smartComponents {
		log.Debug().Stringer("smartComponent", sc).Msgf("<Sample %v>", i)
	}
	w.pubSub.Publish(reqID, Encode(smartComponents))
}

func (w *WrapperSync) TickReqParams(tickerID TickerID, minTick float64, bboExchange string, snapshotPermissions int64) {
	log.Debug().Int64("TickerID", tickerID).Str("minTick", FloatMaxString(minTick)).Str("bboExchange", bboExchange).Str("snapshotPermissions", IntMaxString(snapshotPermissions)).Msg("<TickReqParams>")

	w.state.mu.Lock()
	defer w.state.mu.Unlock()

	ticker, ok := w.state.reqID2Ticker[tickerID]
	if !ok {
		log.Error().Err(errUnknowReqID).Msg("<TickReqParams>")
		return
	}

	ticker.mu.Lock()
	defer ticker.mu.Unlock()
	ticker.minTick = minTick
	ticker.bboExchange = bboExchange
	ticker.snapshotPermissions = snapshotPermissions
}

func (w *WrapperSync) NewsProviders(newsProviders []NewsProvider) {
	for _, np := range newsProviders {
		log.Debug().Stringer("newsProvider", np).Msg("<NewsProviders>")
	}
	w.pubSub.Publish("NewsProvider", Encode(newsProviders))
}

func (w *WrapperSync) NewsArticle(requestID int64, articleType int64, articleText string) {
	log.Debug().Int64("requestID", requestID).Int64("articleType", articleType).Str("articleText", articleText).Msg("<NewsArticle>")
	na := &NewsArticle{ArticleType: articleType, ArticleText: articleText}
	w.pubSub.Publish(requestID, Encode(na))
}

func (w *WrapperSync) HistoricalNews(requestID int64, time string, providerCode string, articleID string, headline string) {
	log.Debug().Int64("requestID", requestID).Str("news time", time).Str("providerCode", providerCode).Str("providerCode", providerCode).Str("headline", headline).Msg("<HistoricalNews>")
	t, err := ParseIBTime(time)
	if err != nil {
		log.Error().Err(err).Msg("<HistoricalNews>")
		return
	}
	hn := HistoricalNews{Time: t, ProviderCode: providerCode, ArticleID: articleID, Headline: headline}
	w.pubSub.Publish(requestID, Join("HistoricalNews", Encode(hn)))
}

func (w *WrapperSync) HistoricalNewsEnd(requestID int64, hasMore bool) {
	log.Debug().Int64("requestID", requestID).Bool("hasMore", hasMore).Msg("<HistoricalNewsEnd>")
	w.pubSub.Publish(requestID, Join("HistoricalNewsEnd", strconv.FormatBool(hasMore)))
}

func (w *WrapperSync) HeadTimestamp(reqID int64, headTimestamp string) {
	log.Debug().Int64("reqID", reqID).Str("headTimestamp", headTimestamp).Msg("<HeadTimestamp>")
	w.pubSub.Publish(reqID, headTimestamp)
}

func (w *WrapperSync) HistogramData(reqID int64, data []HistogramData) {
	log.Debug().Int64("reqID", reqID).Any("data", data).Msg("<HistogramData>")
	w.pubSub.Publish(reqID, Encode(data))
}

func (w *WrapperSync) HistoricalDataUpdate(reqID int64, bar *Bar) {
	log.Debug().Int64("reqID", reqID).Stringer("bar", bar).Msg("<HistoricalDataUpdate>")
	w.pubSub.Publish(reqID, Join("HistoricalDataUpdate", Encode(bar)))
}

func (w *WrapperSync) RerouteMktDataReq(reqID int64, conID int64, exchange string) {
	log.Debug().Int64("reqID", reqID).Int64("conID", conID).Str("exchange", exchange).Msg("<RerouteMktDataReq>")
	w.pubSub.Publish(reqID, Join(Encode(conID), exchange))
}

func (w *WrapperSync) RerouteMktDepthReq(reqID int64, conID int64, exchange string) {
	log.Debug().Int64("reqID", reqID).Int64("conID", conID).Str("exchange", exchange).Msg("<RerouteMktDepthReq>")
	w.pubSub.Publish(reqID, Join(Encode(conID), exchange))
}

func (w *WrapperSync) MarketRule(marketRuleID int64, priceIncrements []PriceIncrement) {
	log.Debug().Int64("marketRuleID", marketRuleID).Any("priceIncrements", priceIncrements).Msg("<MarketRule>")
	w.pubSub.Publish(Key("MarketRule", marketRuleID), Encode(priceIncrements))
}

func (w *WrapperSync) Pnl(reqID int64, dailyPnL float64, unrealizedPnL float64, realizedPnL float64) {
	log.Debug().Int64("reqID", reqID).Str("dailyPnL", FloatMaxString(dailyPnL)).Str("unrealizedPnL", FloatMaxString(unrealizedPnL)).Str("realizedPnL", FloatMaxString(realizedPnL)).Msg("<Pnl>")

	w.state.mu.Lock()
	pnl, ok := w.state.reqID2Pnl[reqID]
	if !ok {
		log.Error().Err(errUnknowReqID).Msg("<Pnl>")
		return
	}
	pnl.DailyPNL = dailyPnL
	pnl.UnrealizedPnl = unrealizedPnL
	pnl.RealizedPNL = realizedPnL
	w.state.mu.Unlock()

	w.pubSub.Publish("Pnl", Encode(pnl))
}

func (w *WrapperSync) PnlSingle(reqID int64, pos Decimal, dailyPnL float64, unrealizedPnL float64, realizedPnL float64, value float64) {
	log.Debug().Int64("reqID", reqID).Str("position", DecimalMaxString(pos)).Str("dailyPnL", FloatMaxString(dailyPnL)).Str("unrealizedPnL", FloatMaxString(unrealizedPnL)).Str("realizedPnL", FloatMaxString(realizedPnL)).Str("value", FloatMaxString(value)).Msg("<PnlSingle>")
	w.state.mu.Lock()
	pnlSingle, ok := w.state.reqID2PnlSingle[reqID]
	if !ok {
		log.Error().Err(errUnknowReqID).Msg("<PnlSingle>")
		return
	}
	pnlSingle.Position = pos
	pnlSingle.DailyPNL = dailyPnL
	pnlSingle.UnrealizedPnl = unrealizedPnL
	pnlSingle.RealizedPNL = realizedPnL
	pnlSingle.Value = value
	w.state.mu.Unlock()

	w.pubSub.Publish("PnlSingle", Encode(pnlSingle))
}

func (w *WrapperSync) HistoricalTicks(reqID int64, ticks []HistoricalTick, done bool) {
	log.Debug().Int64("reqID", reqID).Bool("done", done).Any("ticks", ticks).Msg("<HistoricalTicks>")
	w.pubSub.Publish(reqID, Join(Encode(ticks), strconv.FormatBool(done)))
}

func (w *WrapperSync) HistoricalTicksBidAsk(reqID int64, ticks []HistoricalTickBidAsk, done bool) {
	log.Debug().Int64("reqID", reqID).Bool("done", done).Any("ticks", ticks).Msg("<HistoricalTicksBidAsk>")
	w.pubSub.Publish(reqID, Join(Encode(ticks), strconv.FormatBool(done)))
}

func (w *WrapperSync) HistoricalTicksLast(reqID int64, ticks []HistoricalTickLast, done bool) {
	log.Debug().Int64("reqID", reqID).Bool("done", done).Any("ticks", ticks).Msg("<HistoricalTicksLast>")
	w.pubSub.Publish(reqID, Join(Encode(ticks), strconv.FormatBool(done)))
}

func (w *WrapperSync) TickByTickAllLast(reqID int64, tickType int64, time int64, price float64, size Decimal, tickAttribLast TickAttribLast, exchange string, specialConditions string) {
	log.Debug().Int64("reqID", reqID).Int64("tickType", tickType).Int64("tick time", time).Str("price", FloatMaxString(price)).Str("size", DecimalMaxString(size)).Bool("PastLimit", tickAttribLast.PastLimit).Bool("Unreported", tickAttribLast.Unreported).Str("exchange", exchange).Str("specialConditions", specialConditions).Msg("<TickByTickAllLast>")
	tbtal := TickByTickAllLast{Time: time, TickType: tickType, Price: price, Size: size, TickAttribLast: tickAttribLast, Exchange: exchange, SpecialConditions: specialConditions}

	w.state.mu.Lock()
	ticker := w.state.reqID2Ticker[reqID]
	w.state.mu.Unlock()

	ticker.SetTickByTickAllLast(tbtal)
	w.pubSub.Publish(reqID, Join("AllLast", Encode(tbtal)))
}

func (w *WrapperSync) TickByTickBidAsk(reqID int64, time int64, bidPrice float64, askPrice float64, bidSize Decimal, askSize Decimal, tickAttribBidAsk TickAttribBidAsk) {
	log.Debug().Int64("reqID", reqID).Int64("tick time", time).Str("bidPrice", FloatMaxString(bidPrice)).Str("askPrice", FloatMaxString(askPrice)).Str("bidSize", DecimalMaxString(bidSize)).Str("askSize", DecimalMaxString(askSize)).Bool("AskPastHigh", tickAttribBidAsk.AskPastHigh).Bool("BidPastLow", tickAttribBidAsk.BidPastLow).Msg("<TickByTickBidAsk>")
	tbtba := TickByTickBidAsk{Time: time, BidPrice: bidPrice, AskPrice: askPrice, BidSize: bidSize, AskSize: askSize, TickAttribBidAsk: tickAttribBidAsk}

	w.state.mu.Lock()
	ticker, exists := w.state.reqID2Ticker[reqID]
	w.state.mu.Unlock()

	if exists {
		ticker.SetTickByTickBidAsk(tbtba)
	}

	w.pubSub.Publish(reqID, Join("BidAsk", Encode(tbtba)))
}

func (w *WrapperSync) TickByTickMidPoint(reqID int64, time int64, midPoint float64) {
	log.Debug().Int64("reqID", reqID).Int64("tick time", time).Str("midPoint", FloatMaxString(midPoint)).Msg("<TickByTickMidPoint>")
	tbtmp := TickByTickMidPoint{Time: time, MidPoint: midPoint}

	w.state.mu.Lock()
	ticker, exists := w.state.reqID2Ticker[reqID]
	w.state.mu.Unlock()
	if exists {
		ticker.SetTickByTickMidPoint(tbtmp)
	}
	w.pubSub.Publish(reqID, Encode(tbtmp))
}

func (w *WrapperSync) OrderBound(permID int64, clientID int64, orderID int64) {
	log.Debug().Str("permID", LongMaxString(permID)).Str("clientID", IntMaxString(clientID)).Str("OrderID", IntMaxString(orderID)).Msg("<OrderBound>")
}

func (w *WrapperSync) CompletedOrder(contract *Contract, order *Order, orderState *OrderState) {
	logger := log.Debug().Str("account", order.Account).Str("PermID", LongMaxString(order.PermID)).Str("symbol", contract.Symbol).Str("action", order.Action).Str("orderType", order.OrderType).Str("totalQuantity", DecimalMaxString(order.TotalQuantity)).Str("filledQuantity", DecimalMaxString(order.FilledQuantity))
	logger.Str("lmtPrice", FloatMaxString(order.LmtPrice)).Str("auxPrice", FloatMaxString(order.AuxPrice)).Str("Status", orderState.Status).Str("completedTime", orderState.CompletedTime).Str("CompletedStatus", orderState.CompletedStatus).Msg("<CompletedOrder>")

	orderStatus := OrderStatus{
		OrderID: order.OrderID,
		Status:  Status(orderState.Status),
	}
	trade := NewTrade(contract, order, orderStatus)
	trade.markDone()

	w.state.mu.Lock()
	_, ok := w.state.permID2Trade[order.PermID]
	if !ok {
		w.state.trades[strconv.FormatInt(order.PermID, 10)] = trade
		w.state.permID2Trade[order.PermID] = trade
	}
	w.state.mu.Unlock()

	// w.pubSub.Publish("CompletedOrder", Join(Encode(contract), Encode(order), Encode(orderState)))
}

func (w WrapperSync) CompletedOrdersEnd() {
	log.Info().Msg("<CompletedOrdersEnd>")
	w.pubSub.Publish("CompletedOrdersEnd", "")
}

func (w WrapperSync) ReplaceFAEnd(reqID int64, text string) {
	log.Info().Int64("reqID", reqID).Str("text", text).Msg("<ReplaceFAEnd>")
	w.pubSub.Publish(reqID, text)
}

func (w *WrapperSync) WshMetaData(reqID int64, dataJson string) {
	log.Info().Int64("reqID", reqID).Str("dataJson", dataJson).Msg("<WshMetaData>")
	w.pubSub.Publish(reqID, dataJson)
}

func (w *WrapperSync) WshEventData(reqID int64, dataJson string) {
	log.Debug().Int64("reqID", reqID).Str("dataJson", dataJson).Msg("<WshEventData>")
	w.pubSub.Publish(reqID, dataJson)
}

func (w *WrapperSync) HistoricalSchedule(reqID int64, startDarteTime, endDateTime, timeZone string, sessions []HistoricalSession) {
	log.Debug().Int64("reqID", reqID).Str("startDarteTime", startDarteTime).Str("endDateTime", endDateTime).Str("timeZone", timeZone).Msg("<HistoricalSchedule>")
	hs := HistoricalSchedule{StartDateTime: startDarteTime, EndDateTime: endDateTime, TimeZone: timeZone, Sessions: sessions}
	w.pubSub.Publish(reqID, Encode(hs))
}

func (w *WrapperSync) UserInfo(reqID int64, whiteBrandingId string) {
	log.Debug().Int64("reqID", reqID).Str("whiteBrandingId", whiteBrandingId).Msg("<UserInfo>")
	w.pubSub.Publish(reqID, whiteBrandingId)
}

func (w WrapperSync) CurrentTimeInMillis(timeInMillis int64) {
	log.Debug().Int64("TimeInMillis", timeInMillis).Msg("<CurrentTimeInMillis>")
	w.pubSub.Publish("CurrentTimeInMillis", Encode(timeInMillis))
}

// Protobuf

func (w WrapperSync) ExecDetailsProtoBuf(executionDetailsProto *protobuf.ExecutionDetails) {
	log.Debug().Stringer("ExecutionDetailsProto", executionDetailsProto).Msg("<ExecDetailsProtoBuf>")
}

func (w WrapperSync) ExecDetailsEndProtoBuf(executionDetailsEndProto *protobuf.ExecutionDetailsEnd) {
	log.Debug().Stringer("ExecutionDetailsEndProto", executionDetailsEndProto).Msg("<ExecDetailsEndProtoBuf>")
}

func (w WrapperSync) OrderStatusProtoBuf(orderStatusProto *protobuf.OrderStatus) {
	log.Debug().Stringer("OrderStatusProto", orderStatusProto).Msg("<OrderStatusProtoBuf>")
}

func (w WrapperSync) OpenOrderProtoBuf(openOrderProto *protobuf.OpenOrder) {
	log.Debug().Stringer("OpenOrderProto", openOrderProto).Msg("<OpenOrderProtoBuf>")
}

func (w WrapperSync) OpenOrdersEndProtoBuf(openOrdersEndProto *protobuf.OpenOrdersEnd) {
	log.Debug().Stringer("OpenOrdersEndProto", openOrdersEndProto).Msg("<OpenOrdersEndProtoBuf>")
}

func (w WrapperSync) ErrorProtoBuf(errorProto *protobuf.ErrorMessage) {
	log.Debug().Stringer("ErrorProto", errorProto).Msg("<ErrorProtoBuf>")
}

func (w WrapperSync) CompletedOrderProtoBuf(completedOrderProto *protobuf.CompletedOrder) {
	log.Debug().Stringer("completedOrderProto", completedOrderProto).Msg("<completedOrderProtoBuf>")
}

func (w WrapperSync) CompletedOrdersEndProtoBuf(completedOrdersEndProto *protobuf.CompletedOrdersEnd) {
	log.Debug().Stringer("completedOrdersEndProto", completedOrdersEndProto).Msg("<completedOrdersEndProtoBuf>")
}

func (w WrapperSync) OrderBoundProtoBuf(orderBoundProto *protobuf.OrderBound) {
	log.Debug().Stringer("orderBoundProto", orderBoundProto).Msg("<orderBoundProtoBuf>")
}

func (w WrapperSync) ContractDataProtoBuf(contractDataProto *protobuf.ContractData) {
	log.Debug().Stringer("contractDataProto", contractDataProto).Msg("<ContractDataProtoBuf>")
}

func (w WrapperSync) BondContractDataProtoBuf(contractDataProto *protobuf.ContractData) {
	log.Debug().Stringer("contractDataProto", contractDataProto).Msg("<BondContractDataProtoBuf>")
}

func (w WrapperSync) ContractDataEndProtoBuf(contractDataEndProto *protobuf.ContractDataEnd) {
	log.Debug().Stringer("contractDataEndProto", contractDataEndProto).Msg("<ContractDataEndProtoBuf>")
}

func (w WrapperSync) TickPriceProtoBuf(tickPriceProto *protobuf.TickPrice) {
	log.Debug().Stringer("tickPriceProto", tickPriceProto).Msg("<TickPriceProtoBuf>")
}

func (w WrapperSync) TickSizeProtoBuf(tickSizeProto *protobuf.TickSize) {
	log.Debug().Stringer("tickSizeProto", tickSizeProto).Msg("<TickSizeProtoBuf>")
}

func (w WrapperSync) TickOptionComputationProtoBuf(tickOptionComputationProto *protobuf.TickOptionComputation) {
	log.Debug().Stringer("tickOptionComputationProto", tickOptionComputationProto).Msg("<TickOptionComputationProtoBuf>")
}

func (w WrapperSync) TickGenericProtoBuf(tickGenericProto *protobuf.TickGeneric) {
	log.Debug().Stringer("tickGenericProto", tickGenericProto).Msg("<TickGenericProtoBuf>")
}

func (w WrapperSync) TickStringProtoBuf(tickStringProto *protobuf.TickString) {
	log.Debug().Stringer("tickStringProto", tickStringProto).Msg("<TickStringProtoBuf>")
}

func (w WrapperSync) TickSnapshotEndProtoBuf(tickSnapshotEndProto *protobuf.TickSnapshotEnd) {
	log.Debug().Stringer("tickSnapshotEndProto", tickSnapshotEndProto).Msg("<TickSnapshotEndProtoBuf>")
}

func (w WrapperSync) UpdateMarketDepthProtoBuf(marketDepthProto *protobuf.MarketDepth) {
	log.Debug().Stringer("marketDepthProto", marketDepthProto).Msg("<UpdateMarketDepthProtoBuf>")
}

func (w WrapperSync) UpdateMarketDepthL2ProtoBuf(marketDepthL2Proto *protobuf.MarketDepthL2) {
	log.Debug().Stringer("marketDepthL2Proto", marketDepthL2Proto).Msg("<UpdateMarketDepthL2ProtoBuf>")
}

func (w WrapperSync) MarketDataTypeProtoBuf(marketDataTypeProto *protobuf.MarketDataType) {
	log.Debug().Stringer("marketDataTypeProto", marketDataTypeProto).Msg("<MarketDataTypeProtoBuf>")
}

func (w WrapperSync) TickReqParamsProtoBuf(tickReqParamsProto *protobuf.TickReqParams) {
	log.Debug().Stringer("tickReqParamsProto", tickReqParamsProto).Msg("<TickReqParamsProtoBuf>")
}

func (w WrapperSync) UpdateAccountValueProtoBuf(accountValueProto *protobuf.AccountValue) {
	log.Debug().Stringer("accountValueProto", accountValueProto).Msg("<UpdateAccountValueProtoBuf>")
}

func (w WrapperSync) UpdatePortfolioProtoBuf(portfolioValueProto *protobuf.PortfolioValue) {
	log.Debug().Stringer("portfolioValueProto", portfolioValueProto).Msg("<UpdatePortfolioProtoBuf>")
}

func (w WrapperSync) UpdateAccountTimeProtoBuf(accountUpdateTimeProto *protobuf.AccountUpdateTime) {
	log.Debug().Stringer("accountUpdateTimeProto", accountUpdateTimeProto).Msg("<UpdateAccountTimeProtoBuf>")
}

func (w WrapperSync) AccountDataEndProtoBuf(accountDataEndProto *protobuf.AccountDataEnd) {
	log.Debug().Stringer("accountDataEndProto", accountDataEndProto).Msg("<AccountDataEndProtoBuf>")
}

func (w WrapperSync) ManagedAccountsProtoBuf(managedAccountsProto *protobuf.ManagedAccounts) {
	log.Debug().Stringer("managedAccountsProto", managedAccountsProto).Msg("<ManagedAccountsProtoBuf>")
}

func (w WrapperSync) PositionProtoBuf(positionProto *protobuf.Position) {
	log.Debug().Stringer("positionProto", positionProto).Msg("<PositionProtoBuf>")
}

func (w WrapperSync) PositionEndProtoBuf(positionEndProto *protobuf.PositionEnd) {
	log.Debug().Stringer("positionEndProto", positionEndProto).Msg("<PositionEndProtoBuf>")
}

func (w WrapperSync) AccountSummaryProtoBuf(accountSummaryProto *protobuf.AccountSummary) {
	log.Debug().Stringer("accountSummaryProto", accountSummaryProto).Msg("<AccountSummaryProtoBuf>")
}

func (w WrapperSync) AccountSummaryEndProtoBuf(accountSummaryEndProto *protobuf.AccountSummaryEnd) {
	log.Debug().Stringer("accountSummaryEndProto", accountSummaryEndProto).Msg("<AccountSummaryEndProtoBuf>")
}

func (w WrapperSync) PositionMultiProtoBuf(positionMultiProto *protobuf.PositionMulti) {
	log.Debug().Stringer("positionMultiProto", positionMultiProto).Msg("<PositionMultiProtoBuf>")
}

func (w WrapperSync) PositionMultiEndProtoBuf(positionMultiEndProto *protobuf.PositionMultiEnd) {
	log.Debug().Stringer("positionMultiEndProto", positionMultiEndProto).Msg("<PositionMultiEndProtoBuf>")
}

func (w WrapperSync) AccountUpdateMultiProtoBuf(accountUpdateMultiProto *protobuf.AccountUpdateMulti) {
	log.Debug().Stringer("accountUpdateMultiProto", accountUpdateMultiProto).Msg("<AccountUpdateMultiProtoBuf>")
}

func (w WrapperSync) AccountUpdateMultiEndProtoBuf(accountUpdateMultiEndProto *protobuf.AccountUpdateMultiEnd) {
	log.Debug().Stringer("accountUpdateMultiEndProto", accountUpdateMultiEndProto).Msg("<AccountUpdateMultiEndProtoBuf>")
}
