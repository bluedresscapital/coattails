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
