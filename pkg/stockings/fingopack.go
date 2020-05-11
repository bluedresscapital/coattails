package stockings

import (
	"time"

	"github.com/piquette/finance-go"
	"github.com/piquette/finance-go/chart"
	"github.com/piquette/finance-go/datetime"
	"github.com/piquette/finance-go/quote"

	"github.com/shopspring/decimal"
)

// FingoPack will be used to determine which interface we are using
type FingoPack struct {
}

// fingoHistoricalStocks is a list of struct ChartBar which is determined by finance-go
type piqHistoricalStocks []finance.ChartBar

func (piq FingoPack) GetCurrentPrice(ticker string) (*Stock, error) {

	// simple call to get the quote of a stock in the finance.quote struct
	quote, err := quote.Get(ticker)
	if err != nil {
		return nil, err
	}

	// need to convert our types to match our universal Stock struct
	// some of the types are already correct, but listed them all out for clarity
	// finance.quote has a HUGE variety of info in it that we can pull
	symbol := quote.Symbol
	name := quote.ShortName
	currentPrice := decimal.NewFromFloat(quote.RegularMarketPrice)
	priceChange := float32(quote.RegularMarketChange)
	percentChange := float32(quote.RegularMarketChangePercent)

	return &Stock{symbol, name, currentPrice, priceChange, percentChange}, nil

}

// GetHistoricalPrice is hacky because Fingo only has a function that returns quotes in a time span
// This function takes in the desired date as the end date, then sets the start date as five days prior
func (piq FingoPack) GetHistoricalPrice(ticker string, date time.Time) (*HistoricalStock, error) {

	quoteAtDate := new(finance.ChartBar)

	// creating a hypothetical start five days before (is this long enough)
	adjStart := date.AddDate(0, 0, -5)
	// ending date is set at desired date + 1 day because of how financego slices its range
	adjEnd := date.AddDate(0, 0, 1)

	// returns and Iter that holds all the information we need
	iter := getBarChartIter(ticker, adjStart, adjEnd)

	// iter.Next() eventually equals false when you've finished
	for iter.Next() {
		// we are only interested in saving the last value in our iterator because that is reflective of the correct price requested by user
		// note if the user requests a weekend, then we need to return the previous adj close price
		quoteAtDate = iter.Bar()
	}

	// some more blackboxing...somehow this just returns errors :)
	err := iter.Err()
	if err != nil {
		return nil, err
	}

	// at this point quoteAtDate holds the finance.ChartBar struct of information we need
	// just need to convert it into our universal HistoricalStock struct
	return &(HistoricalStock{time.Unix(int64(quoteAtDate.Timestamp), 0), quoteAtDate.AdjClose}), nil

}

// GetHistoricalRange is nearly identical to GetHistoricalPrice, but takes in two dates, so we don't need to create a hypothetical start date
func (piq FingoPack) GetHistoricalRange(ticker string, start time.Time, end time.Time) (*HistoricalStocks, error) {

	// need to create a piqHistoricalStocks type that is a list of finance.ChartBar
	historicalRange := new(piqHistoricalStocks)
	// similar reasoning as previous function, need to +1 day (we can adjust this depending on how we interpret "GetHistoricalRange")
	adjEnd := end.AddDate(0, 0, 1)
	iter := getBarChartIter(ticker, start, adjEnd)

	// now we need to append to our historicalRange becuase we are interested in the entire range
	for iter.Next() {
		*historicalRange = append(*historicalRange, *iter.Bar())
	}

	err := iter.Err()
	if err != nil {
		return nil, err
	}

	// we use our helper function to convert historicalRange into a return struct accepted by our interface
	return piqConvertToHistoricalRange(historicalRange), nil
}

// piqConvertToHistoricalRange converts to accepted interface struct
func piqConvertToHistoricalRange(stocks *piqHistoricalStocks) *HistoricalStocks {

	// ret is struct of what we want
	ret := new(HistoricalStocks)

	// iterate through and append to ret the proper format of what we want
	for i := 0; i < len(*stocks); i++ {
		*ret = append(*ret, HistoricalStock{time.Unix(int64((*stocks)[i].Timestamp), 0), (*stocks)[i].AdjClose})
	}

	return ret
}

// getBarChartIter is a private function that returns a financego chart.Iter type
func getBarChartIter(ticker string, start time.Time, end time.Time) *chart.Iter {

	// converting the time.Time format to an accepted financego format
	fingoStart := datetime.New(&start)
	fingoEnd := datetime.New(&end)

	params := &chart.Params{
		Symbol:   ticker,
		Interval: datetime.OneDay,
		Start:    fingoStart,
		End:      fingoEnd,
	}

	// a little bit of blackboxing here, but this iter contains all the information we need
	return chart.Get(params)
}
