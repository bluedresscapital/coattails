package wardrobe

import "fmt"

func UpsertStock(ticker string) error {
	_, err := db.Exec(`INSERT INTO stocks (ticker) VALUES ($1) ON CONFLICT (ticker) DO NOTHING`, ticker)
	return err
}

func FetchStockIdFromTicker(ticker string) (*int, error) {
	rows, err := db.Query(`SELECT id from STOCKS WHERE ticker=$1`, ticker)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var id int
	if !rows.Next() {
		return nil, fmt.Errorf("no stock found with ticker %s", ticker)
	}
	err = rows.Scan(&id)
	if err != nil {
		return nil, err
	}
	return &id, nil
}
