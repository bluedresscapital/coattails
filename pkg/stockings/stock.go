package stockings

import (
	"github.com/shopspring/decimal"
	"time"
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

func GetHistoricalPrice(stock StockAPI, ticker string, date time.Time) (*decimal.Decimal, error) {
	// TODO
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
