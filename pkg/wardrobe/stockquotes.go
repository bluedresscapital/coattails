package wardrobe

import (
	"fmt"
	"time"

	"github.com/bluedresscapital/coattails/pkg/util"

	"github.com/shopspring/decimal"
)

type StockQuote struct {
	Stock       string          `json:"stock"`
	Price       decimal.Decimal `json:"price"`
	Date        time.Time       `json:"date"`
	IsValidDate bool            `json:"is_valid_date"`
}

func UpsertStockQuote(quote StockQuote) error {
	if quote.Price.IsZero() {
		return fmt.Errorf("stock quote price cannot be zero")
	}
	_, err := db.Exec(`
		INSERT INTO stock_quotes (stock_id, price, date, is_valid_date)
			SELECT stocks.id, $2, $3, $4
			FROM stocks
			WHERE ticker=$1
		ON CONFLICT(stock_id, date) DO UPDATE
		SET price=$2, is_valid_date=$4`,
		quote.Stock, quote.Price, util.GetTimelessDate(quote.Date), quote.IsValidDate)
	return err
}

func FetchStockQuote(ticker string, date time.Time) (*StockQuote, bool, error) {
	rows, err := db.Query(`
		SELECT s.ticker, q.price, q.date, q.is_valid_date 
		FROM stock_quotes q
		JOIN stocks s ON s.id=q.stock_id
		WHERE s.ticker=$1 AND q.date=$2`, ticker, date)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, false, nil
	}
	var sq StockQuote
	err = rows.Scan(&sq.Stock, &sq.Price, &sq.Date, &sq.IsValidDate)
	return &sq, true, err
}
