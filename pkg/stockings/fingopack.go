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

	// need to convert our types to match our universal Stock struct - should this be broken out into a seperate function?
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
// This function takes in the desired date as the end date, then sets the start date as one month prior
func (piq FingoPack) GetHistoricalPrice(ticker string, date time.Time) (*HistoricalStock, error) {

	quoteAtDate := new(finance.ChartBar)

	// converting the time.Time struct into usuable integers to be fed into Fingo
	year := date.Year()
	month := int(date.Month())
	day := date.Day()

	// converting into our accepted Fingo form
	endDate := datetime.Datetime{
		Year:  year,
		Month: month,
		Day:   day,
	}

	// this brings our month down by one
	if month == 1 {
		month = 12
		year--
	}

	// now we can set our start date which is just one month prior - perhaps this part should be put in a seperate function
	startDate := datetime.Datetime{
		Year:  year,
		Month: month,
		Day:   day,
	}

	// now we are using the Fingo steps to get historical stock info in a range
	// set our paratemters which are self explanatory
	params := &chart.Params{
		Symbol:   ticker,
		Interval: datetime.OneDay,
		Start:    &startDate,
		End:      &endDate,
	}

	// a little bit of blackboxing here, but this iter contains all the information we need
	iter := chart.Get(params)

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

// GetHistoricalRange is nearly identical to GetHistoricalPrice, but there are some nuances - might not be able to create a seperate helper function
func (piq FingoPack) GetHistoricalRange(ticker string, start time.Time, end time.Time) (*HistoricalStocks, error) {

	// need to create a piqHistoricalStocks type that is a list of finance.ChartBar
	historicalRange := new(piqHistoricalStocks)

	// just setting our start and end date based on what we were given
	startDate := datetime.Datetime{
		Year:  start.Year(),
		Month: int(start.Month()),
		Day:   start.Day(),
	}

	endDate := datetime.Datetime{
		Year:  end.Year(),
		Month: int(end.Month()),
		Day:   end.Day(),
	}

	// setting params, same as GetHistoricalPrice
	params := &chart.Params{
		Symbol:   ticker,
		Interval: datetime.OneDay,
		Start:    &startDate,
		End:      &endDate,
	}

	iter := chart.Get(params)

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
