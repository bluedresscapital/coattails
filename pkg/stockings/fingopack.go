package stockings

import (
	"errors"
	"time"

	"github.com/piquette/finance-go"
	"github.com/piquette/finance-go/chart"
	"github.com/piquette/finance-go/datetime"
	"github.com/piquette/finance-go/quote"

	"github.com/shopspring/decimal"
)

type FingoPack struct {
}

var _ StockAPI = (*FingoPack)(nil)

type piqHistoricalStocks []finance.ChartBar

func (piq FingoPack) GetCurrentPrice(ticker string) (*Stock, error) {

	quote, err := quote.Get(ticker)
	if err != nil {
		return nil, err
	}

	symbol := quote.Symbol
	name := quote.ShortName
	currentPrice := decimal.NewFromFloat(quote.RegularMarketPrice)
	priceChange := float32(quote.RegularMarketChange)
	percentChange := float32(quote.RegularMarketChangePercent)

	return &Stock{symbol, name, currentPrice, priceChange, percentChange}, nil

}

// we use the GetHistoricalRange function and set our start date to 5 days prior becuase there has to be a valid open market date in the range
func (piq FingoPack) GetHistoricalPrice(ticker string, date time.Time) (*HistoricalStock, error) {

	historicalQuotes, err := piq.GetHistoricalRange(ticker, date.AddDate(0, 0, -5), date)

	if err != nil {
		return nil, err
	}

	if len(*historicalQuotes) < 0 {
		return nil, errors.New("Invalid Date")
	}

	// only interested in the last item which is the price at the date the user requested
	return &(*historicalQuotes)[len(*historicalQuotes)-1], nil

}

func (piq FingoPack) GetHistoricalRange(ticker string, start time.Time, end time.Time) (*HistoricalStocks, error) {

	// adding a day to our end date so that our range includes the price on that date
	barChart, err := getBarChartIter(ticker, start, end.AddDate(0, 0, 1))
	if err != nil {
		return nil, err
	}

	return barChart, nil
}

// piqConvertToHistoricalRange converts to accepted interface struct
func piqConvertToHistoricalRange(stocks *piqHistoricalStocks) *HistoricalStocks {

	ret := new(HistoricalStocks)

	for i := 0; i < len(*stocks); i++ {
		*ret = append(*ret, HistoricalStock{time.Unix(int64((*stocks)[i].Timestamp), 0), (*stocks)[i].AdjClose})
	}

	return ret
}

func getBarChartIter(ticker string, start time.Time, end time.Time) (*HistoricalStocks, error) {

	historicalRange := new(piqHistoricalStocks)
	// converting the time.Time format to an accepted financego format
	fingoStart := datetime.New(&start)
	fingoEnd := datetime.New(&end)

	params := &chart.Params{
		Symbol:   ticker,
		Interval: datetime.OneDay,
		Start:    fingoStart,
		End:      fingoEnd,
	}

	iter := chart.Get(params)

	for iter.Next() {
		*historicalRange = append(*historicalRange, *iter.Bar())
	}

	err := iter.Err()
	if err != nil {
		return nil, err
	}

	// a little bit of blackboxing here, but this iter contains all the information we need
	return piqConvertToHistoricalRange(historicalRange), nil
}
