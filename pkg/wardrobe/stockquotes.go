package wardrobe

import (
	"fmt"
	"time"

	"github.com/lib/pq"

	"github.com/shopspring/decimal"
)

type StockQuote struct {
	Stock string          `json:"stock"`
	Price decimal.Decimal `json:"price"`
	Date  time.Time       `json:"date"`
}

func UpsertStockQuotePrice(ticker string, date time.Time, price decimal.Decimal) error {
	id, err := FetchStockIdFromTicker(ticker)
	if err != nil {
		return err
	}
	_, err = db.Exec(`
		INSERT INTO stock_quotes (stock_id, price, date)
		VALUES ($1, $2, $3)
		ON CONFLICT (stock_id, date) DO UPDATE
		SET price=$2`, *id, price, date)
	return err
}

//func FetchStaleStockQuotes() {
//	rows, err := db.Query(`
//		SELECT s.ticker
//		FROM stock_quotes q
//		JOIN stocks s ON s.id=q.stock_id
//		WHERE
//	`)
//}

func BatchUpsertStockQuotes(quotes []StockQuote) error {
	if len(quotes) == 0 {
		return nil
	}
	id, err := FetchStockIdFromTicker(quotes[0].Stock)
	if err != nil {
		return err
	}
	txn, err := db.Begin()
	if err != nil {
		return err
	}
	start := quotes[0].Date
	end := quotes[len(quotes)-1].Date
	_, err = txn.Exec(`DELETE FROM stock_quotes WHERE stock_id=$1 AND date >= $2 AND date <= $3`, *id, start, end)
	if err != nil {
		txn.Rollback()
		return err
	}
	stmt, _ := txn.Prepare(pq.CopyIn("stock_quotes", "stock_id", "price", "date"))
	if err != nil {
		return err
	}
	for _, q := range quotes {
		_, err = stmt.Exec(*id, q.Price, q.Date)
		if err != nil {
			return err
		}
	}
	_, err = stmt.Exec()
	if err != nil {
		return err
	}
	err = stmt.Close()
	if err != nil {
		return err
	}
	err = txn.Commit()
	if err != nil {
		return err
	}
	return nil
}

func FetchStockQuoteCount(ticker string, start time.Time, end time.Time) (*int, error) {
	id, err := FetchStockIdFromTicker(ticker)
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(`SELECT COUNT(*) FROM stock_quotes WHERE stock_id=$1 AND date >= $2 AND date <= $3`, *id, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, fmt.Errorf("unable to get stock quote count: %v", err)
	}
	var count int
	err = rows.Scan(&count)
	if err != nil {
		return nil, err
	}
	return &count, nil
}

func FetchStockQuotes(ticker string, start time.Time, end time.Time) ([]StockQuote, error) {
	rows, err := db.Query(`
		SELECT s.ticker, q.price, q.date
		FROM stock_quotes q
		JOIN stocks s ON s.id=q.stock_id
		WHERE s.ticker=$1 AND q.date>=$2 AND q.date <=$3
		ORDER BY q.date`, ticker, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ret := make([]StockQuote, 0)
	for rows.Next() {
		var sq StockQuote
		err = rows.Scan(&sq.Stock, &sq.Price, &sq.Date)
		ret = append(ret, sq)
	}
	return ret, err
}
