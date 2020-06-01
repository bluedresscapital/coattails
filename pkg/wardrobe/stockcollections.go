package wardrobe

import (
	"github.com/bluedresscapital/coattails/pkg/util"
	"github.com/lib/pq"
)

func FetchStaleStockCollections() ([]string, error) {
	rows, err := db.Query(`SELECT ticker FROM stocks WHERE updated_collections_at IS NULL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tickers := make([]string, 0)
	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		tickers = append(tickers, id)
	}
	return tickers, nil
}

func UpsertCollections(collections []string) (map[string]int, error) {
	txn, err := db.Begin()
	if err != nil {
		return nil, err
	}
	// Upsert collections, somehow get a list of their ids??
	collectionIds := make(map[string]int)
	stmt, _ := txn.Prepare(`
		INSERT INTO collections (name) 
		VALUES ($1) 
		ON CONFLICT(name) DO UPDATE SET name=excluded.name
		RETURNING id`)
	for _, name := range collections {
		var id int
		err := stmt.QueryRow(name).Scan(&id)
		if err != nil {
			return nil, err
		}
		collectionIds[name] = id
	}
	stmt.Close()
	err = txn.Commit()
	if err != nil {
		return nil, err
	}
	return collectionIds, nil
}

func UpsertStockCollections(collectionIds []int, ticker string) error {
	if len(collectionIds) == 0 {
		return nil
	}
	id, err := FetchStockIdFromTicker(ticker)
	if err != nil {
		return err
	}
	txn, err := db.Begin()
	if err != nil {
		return err
	}
	_, err = txn.Exec(`DELETE FROM stock_collections WHERE stock_id=$1`, *id)
	if err != nil {
		txn.Rollback()
		return err
	}
	stmt, err := txn.Prepare(pq.CopyIn("stock_collections", "stock_id", "collection_id"))
	if err != nil {
		txn.Rollback()
		return err
	}
	for _, collectionId := range collectionIds {
		_, err = stmt.Exec(*id, collectionId)
		if err != nil {
			txn.Rollback()
			return err
		}
	}
	_, err = stmt.Exec()
	if err != nil {
		txn.Rollback()
		return err
	}
	err = stmt.Close()
	if err != nil {
		txn.Rollback()
		return err
	}
	updateStmt, err := txn.Prepare("UPDATE stocks SET updated_collections_at=$1")
	if err != nil {
		txn.Rollback()
		return err
	}
	now := util.GetTimelessESTOpenNow()
	_, err = updateStmt.Exec(now)
	if err != nil {
		txn.Rollback()
		return err
	}
	updateStmt.Close()
	err = txn.Commit()
	if err != nil {
		return err
	}
	return nil
}
