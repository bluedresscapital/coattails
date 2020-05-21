package stockings

import (
	"fmt"
	"log"
	"time"

	"github.com/bluedresscapital/coattails/pkg/util"
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"github.com/shopspring/decimal"
)

/*
	Could look to make this an interface as we move forward with different stock apis
	ie make different structs for td, iex, rh apis and implement each of these methods for the
	different apis via StockAPI interface.

*/

type StockAPI interface {
	GetCurrentPrice(ticker string) (*Stock, error)
	GetHistoricalPrice(ticker string, date time.Time) (*HistoricalStock, error)
	GetHistoricalRange(ticker string, start time.Time, end time.Time) (*HistoricalStocks, error)
}

type Stock struct {
	Symbol        string
	Name          string
	LatestPrice   decimal.Decimal
	Change        float32
	ChangePercent float32
}

type HistoricalStock struct {
	Date  time.Time
	Price decimal.Decimal
}

type HistoricalStocks []HistoricalStock

func GetHistoricalPrice(api StockAPI, ticker string, date time.Time) (*decimal.Decimal, error) {
	date = util.GetTimelessDate(date)
	hist, err := GetHistoricalRange(api, ticker, date, date)
	if err != nil {
		return nil, err
	}
	if len(*hist) != 1 || !(*hist)[0].Date.Equal(date) {
		return nil, fmt.Errorf("invalid history: %v for %s on %s", *hist, ticker, date)
	}
	price := (*hist)[0].Price
	return &price, nil
}

func GetCurrentPrice(api StockAPI, ticker string) (*decimal.Decimal, error) {
	return GetHistoricalPrice(api, ticker, util.GetTimelessDate(time.Now()))
}

// GetHistoricalRange will return prices for *EVERY DAY* from start to end
func GetHistoricalRange(api StockAPI, ticker string, start time.Time, end time.Time) (*HistoricalStocks, error) {
	if start.After(end) {
		return nil, fmt.Errorf("start date (%s) is after end (%s)", start, end)
	}
	now := time.Now()
	if start.After(now) || end.After(now) {
		return nil, fmt.Errorf("start %s or end %s is after now %s, which is invalid", start, end, now)
	}
	start = util.GetTimelessDate(start)
	end = util.GetTimelessDate(end)
	days := int(end.Sub(start).Hours()/24) + 1 // Add one to include end date
	count, err := wardrobe.FetchStockQuoteCount(ticker, start, end)
	if err != nil {
		return nil, err
	}
	if days == *count {
		log.Printf("Fetching stock quotes from db...")
		sq, err := wardrobe.FetchStockQuotes(ticker, start, end)
		if err != nil {
			return nil, err
		}
		ret := new(HistoricalStocks)
		for _, q := range sq {
			*ret = append(*ret, HistoricalStock{
				Date:  q.Date,
				Price: q.Price,
			})
		}
		return ret, nil
	}
	log.Printf("We only have %d/%d stock quotes, fetching stock quotes from api...", days, *count)
	stocksP, err := api.GetHistoricalRange(ticker, start, end)
	if err != nil {
		return nil, fmt.Errorf("errored out from stock api's get historical range: %v", err)
	}
	log.Printf("Finished fetching stock quotes from api!")
	if len(*stocksP) != days {
		return nil, fmt.Errorf("api.GetHistoricalRange should've returned %d prices between %s to %s, only got %d", days, start, end, len(*stocksP))
	}
	quotes := make([]wardrobe.StockQuote, 0)
	for _, s := range *stocksP {
		quotes = append(quotes, wardrobe.StockQuote{
			Stock: ticker,
			Price: s.Price,
			Date:  s.Date,
		})
	}
	log.Printf("Bulk inserting stock quotes...")
	err = wardrobe.BatchUpsertStockQuotes(quotes)
	if err != nil {
		return nil, err
	}
	log.Printf("Done bulk inserting stock quotes!")
	return stocksP, nil
}
