package stockings

import (
	"time"

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
	// TODO (ralles)
	// 1. First, check if that price is in our DB

	// 2. If not, return api call:
	//price, err := stock.GetHistoricalPrice(ticker, date)
	//if err != nil {
	//	return nil, err
	//}
	// 3. Now we should store it in our db so we don't have to make the same api call again
	//return &price.Price, nil
	return &decimal.Zero, nil
}

func GetCurrentPrice(api StockAPI) (*decimal.Decimal, error) {
	// TODO (ralles)
	return &decimal.Zero, nil
}

// GetHistoricalRange will return prices for *EVERY DAY* from start to end
func GetHistoricalRange(api StockAPI, ticker string, start time.Time, end time.Time) (*HistoricalStocks, error) {
	res := new(HistoricalStocks)
	missingQuote := false
	for currDate := start; currDate.Before(end.AddDate(0, 0, 1)); currDate = currDate.AddDate(0, 0, 1) {
		sq, found, err := wardrobe.FetchStockQuote(ticker, currDate)
		if err != nil || !found {
			missingQuote = true
			break
		}
		if sq.IsValidDate {
			*res = append(*res, HistoricalStock{
				Date:  sq.Date,
				Price: sq.Price,
			})
		}
	}
	if !missingQuote {
		return res, nil
	}

	stocksP, err := api.GetHistoricalRange(ticker, start, end)
	if err != nil {
		return nil, err
	}
	stocks := *stocksP
	stockMap := make(map[time.Time]decimal.Decimal)
	for _, stock := range stocks {
		stockMap[stock.Date] = stock.Price
	}

	var currPrice decimal.Decimal
	if len(stocks) == 0 || stocks[0].Date.After(start) {
		historicalPrice, err := api.GetHistoricalPrice(ticker, start)
		if err != nil {
			return nil, err
		}
		currPrice = historicalPrice.Price
	} else {
		currPrice = stocks[0].Price
	}

	// In case the underlying stock is missing, just upsert it :)
	err = wardrobe.UpsertStock(ticker)
	if err != nil {
		return nil, err
	}

	for currDate := start; currDate.Before(end.AddDate(0, 0, 1)); currDate = currDate.AddDate(0, 0, 1) {
		price, found := stockMap[currDate]
		if found {
			currPrice = price
		}
		err = wardrobe.UpsertStockQuote(wardrobe.StockQuote{
			Stock:       ticker,
			Price:       currPrice,
			Date:        currDate,
			IsValidDate: found,
		})
		if err != nil {
			return nil, err
		}
	}

	return stocksP, nil
}
