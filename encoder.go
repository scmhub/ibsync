package ibsync

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strings"

	"github.com/scmhub/ibapi"
)

const sep = "::"

// init registers various common structs with gob for encoding/decoding.
func init() {
	// Common
	gob.Register(Fill{})
	gob.Register(OptionChain{})
	gob.Register(TickPrice{})
	gob.Register(TickSize{})
	gob.Register(TickOptionComputation{})
	gob.Register(TickGeneric{})
	gob.Register(TickString{})
	gob.Register(TickEFP{})

	// ibapi
	gob.Register(ibapi.Bar{})
	gob.Register(ibapi.CodeMsgPair{})
	gob.Register(ibapi.ComboLeg{})
	gob.Register(ibapi.CommissionReport{})
	gob.Register(ibapi.Contract{})
	gob.Register(ibapi.ContractDetails{})
	gob.Register(ibapi.ContractDescription{})
	gob.Register(ibapi.Decimal{})
	gob.Register(ibapi.DeltaNeutralContract{})
	gob.Register(ibapi.DepthMktDataDescription{})
	gob.Register(ibapi.Execution{})
	gob.Register(ibapi.FamilyCode{})
	gob.Register(ibapi.HistogramData{})
	gob.Register(ibapi.HistoricalTick{})
	gob.Register(ibapi.HistoricalTickBidAsk{})
	gob.Register(ibapi.HistoricalTickLast{})
	gob.Register(ibapi.HistoricalSession{})
	gob.Register(ibapi.IneligibilityReason{})
	gob.Register(ibapi.NewsProvider{})
	gob.Register(ibapi.Order{})
	gob.Register(ibapi.OrderState{})
	gob.Register(ibapi.PriceIncrement{})
	gob.Register(ibapi.RealTimeBar{})
	gob.Register(ibapi.SmartComponent{})
	gob.Register(ibapi.SoftDollarTier{})
	gob.Register(ibapi.TagValue{})
	gob.Register(ibapi.TickAttrib{})
	gob.Register(ibapi.TickAttribBidAsk{})
	gob.Register(ibapi.TickAttribLast{})
	gob.Register(ibapi.WshEventData{})
	// gob.Register(IneligibilityReason(""))
	// gob.Register(FundDistributionPolicyIndicator(""))
}

// isErrorMsg returns true if provided msg contains "error" string.
func isErrorMsg(msg string) bool {
	return strings.Contains(msg, "error")
}

// msg2Error decodes msg to ibapi CodeMsgPair
func msg2Error(msg string) error {
	var cmp ibapi.CodeMsgPair
	if err := Decode(&cmp, Split(msg)[1]); err != nil {
		return err
	}
	return normaliseCodeMsgPair(cmp)
}

// orderKey generates a unique key for an order based on client ID, order ID, or permanent ID.
func orderKey(clientID int64, orderID OrderID, permID int64) string {
	if orderID <= 0 {
		return Key(permID)
	}
	return Key(clientID, orderID)
}

// Key constructs a unique string key from a variadic number of values, separated by a specified delimiter.
// default delimiter is "::"
func Key(keys ...any) string {
	if len(keys) == 0 {
		return ""
	}
	var sb strings.Builder
	fmt.Fprint(&sb, keys[0])

	for i := 1; i < len(keys); i++ {
		fmt.Fprintf(&sb, "%s%v", sep, keys[i])
	}

	return sb.String()
}

// Encode serializes an input value to a gob-encoded string.
func Encode(e any) string {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	if err := enc.Encode(e); err != nil {
		log.Panic().Err(err).Any("e", e).Msg("internal encoding error")
	}
	return b.String()
}

// Decode deserializes a gob-encoded string back into a Go value.
//
// Parameters:
//   - e: A pointer to the target value where decoded data will be stored
//   - data: The gob-encoded string to be decoded
func Decode(e any, data string) error {
	buf := bytes.NewBufferString(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(e)
	return err
}

// Join concatenates multiple strings using the package's default separator ("::").
func Join(strs ...string) string {
	return strings.Join(strs, sep)
}

// Split divides a string into substrings using the package's default separator ("::").
func Split(str string) []string {
	return strings.Split(str, sep)
}
