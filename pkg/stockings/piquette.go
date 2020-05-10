package stockings

import (
	"time"

	"github.com/piquette/finance-go"

	"github.com/piquette/finance-go/chart"
	"github.com/piquette/finance-go/datetime"
	"github.com/piquette/finance-go/quote"
	"github.com/shopspring/decimal"
)

// "github.com/shopspring/decimal"

// "net/http"
// "time"

// honestly is piquette even an API? its more of a libary/wrapper
type PiquetteAPI struct {
}

type piqHistoricalStocks []finance.ChartBar

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

func (piq PiquetteAPI) GetHistoricalPrice(ticker string, date time.Time) (*HistoricalStock, error) {

	historical := new(piqHistoricalStocks)

	year := date.Year()
	month := int(date.Month())
	day := date.Day()

	endDate := datetime.Datetime{
		Year:  year,
		Month: month,
		Day:   day,
	}

	if month == 1 {
		month = 12
		year--
	}

	// here we need to create a default startdate just one month earlier - perhaps make a seperate function here

	startDate := datetime.Datetime{
		Year:  year,
		Month: month,
		Day:   day,
	}

	// for some reason Piquette requires a start and end, so we're going to have to fabricate one
	// the end date is just whatever date they entered, but if the date entered is a weekend then theres a problem

	// a sus inefficient way is subtract a year from end date and get the start date

	params := &chart.Params{
		Symbol:   "meli",
		Interval: datetime.OneDay,
		Start:    &startDate,
		End:      &endDate,
	}

	// need to write in the values of our iterator into some structure, which is a list? or just create a list variable
	iter := chart.Get(params)
	for iter.Next() {
		*historical = append(*historical, *iter.Bar())

	}
	err := iter.Err()
	if err != nil {
		return nil, err
	}

	// now I need to find the stock price on the specific date
	// to do this i just need to return the last element in the array, this is very hacky
	// now I need to convert into the interface format

	quoteAtDate := (*historical)[len(*historical)-1]

	return &(HistoricalStock{time.Unix(int64(quoteAtDate.Timestamp), 0), quoteAtDate.AdjClose}), nil

}
