package stockings

import (
	"fmt"
	"time"

	"github.com/bluedresscapital/coattails/pkg/util"

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

func (piq FingoPack) GetHistoricalPrice(ticker string, date time.Time) (*HistoricalStock, error) {
	// Assume GetHistoricalRange (ticker, start, end) will return prices from [start, end] (inclusive)
	historicalQuotes, err := piq.GetHistoricalRange(ticker, date, date)
	if err != nil {
		return nil, err
	}
	if len(*historicalQuotes) == 0 {
		return nil, fmt.Errorf("couldn't find a valid price for date %s", date)
	}
	// only interested in the last item which is the price at the date the user requested
	return &(*historicalQuotes)[len(*historicalQuotes)-1], nil
}

func (piq FingoPack) GetHistoricalRange(ticker string, start time.Time, end time.Time) (*HistoricalStocks, error) {
	if end.Before(start) {
		return nil, fmt.Errorf("invalid date range. start (%s) is after end (%s)", start, end)
	}
	historicalStocks, err := getHistoricalStocks(ticker, start, end)
	if err != nil {
		return nil, err
	}
	return historicalStocks, nil
}

func getHistoricalStocks(ticker string, start time.Time, end time.Time) (*HistoricalStocks, error) {
	start = util.GetTimelessDate(start)
	end = util.GetTimelessDate(end)
	// Just subtract 5 days (for now) to try and guarantee we can get some valid price for our date range
	// For example, if start was on a weekend, we'd want to get the most recent valid price prior to start.
	startRange := start.AddDate(0, 0, -5)
	// We need to set the end to + 1 day because fingo does exclusive end date for range
	endRange := end.AddDate(0, 0, 1)
	params := &chart.Params{
		Symbol:   ticker,
		Interval: datetime.OneDay,
		Start:    datetime.New(&startRange),
		End:      datetime.New(&(endRange)),
	}
	iter := chart.Get(params)
	if iter.Count() == 0 {
		return nil, fmt.Errorf("fingo empty iter returned for %s from [%s to %s)", ticker, startRange, endRange)
	}
	historicalRange := new(piqHistoricalStocks)
	for iter.Next() {
		*historicalRange = append(*historicalRange, *iter.Bar())
	}
	err := iter.Err()
	if err != nil {
		return nil, err
	}
	// a little bit of blackboxing here, but this iter contains all the information we need
	return piqConvertToHistoricalRange(historicalRange, start, end), nil
}

// piqConvertToHistoricalRange converts to accepted interface struct
func piqConvertToHistoricalRange(stocks *piqHistoricalStocks, start time.Time, end time.Time) *HistoricalStocks {
	ret := new(HistoricalStocks)
	for i := 0; i < len(*stocks); i++ {
		date := util.GetTimelessDate(time.Unix(int64((*stocks)[i].Timestamp), 0))
		if date.Before(start) || date.After(end) {
			continue
		}
		*ret = append(*ret, HistoricalStock{
			Date:  date,
			Price: (*stocks)[i].AdjClose,
		})
	}
	return ret
}
