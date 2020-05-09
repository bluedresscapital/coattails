package stockings

import (
	"github.com/piquette/finance-go/quote"
	"github.com/shopspring/decimal"
)

// "github.com/shopspring/decimal"

// "net/http"
// "time"

// honestly is piquette even an API? its more of a libary/wrapper
type PiquetteAPI struct{}

// why exactly do we need to use piq here? is that just not being used, but could be used
func (piq PiquetteAPI) GetCurrentPrice(ticker string) (*Stock, error) {
	// don't need any http or url shit -> pretty sure this isn't an API
	// fullQuote is a struct finance.Quote
	quote, err := quote.Get(ticker)
	if err != nil {
		return nil, err
	}

	// theres a problem with types not being consistent between Stock struct and finance.Quote struct
	// just list out the items here for clarity
	symbol := quote.Symbol
	name := quote.ShortName
	currentPrice := decimal.NewFromFloat(quote.RegularMarketPrice)
	priceChange := float32(quote.RegularMarketChange)
	percentChange := float32(quote.RegularMarketChangePercent)

	return &Stock{symbol, name, currentPrice, priceChange, percentChange}, nil

}
