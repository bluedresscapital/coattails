package wardrobe

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// TODO: move this into stockings when rishov merges.
type Stock struct {
	Ticker string `json:"ticker"`
	Name   string `json:"name"`
}

type Order struct {
	Uid           string          `json:"uid"`
	PortId        int             `json:"port_id"`
	Stock         string          `json:"stock"`
	Quantity      decimal.Decimal `json:"quantity"`
	Value         decimal.Decimal `json:"value"`
	IsBuy         bool            `json:"is_buy"`
	ManuallyAdded bool            `json:"manually_added"`
	Date          time.Time       `json:"date"`
}

func FetchOrdersByUserId(userId int) ([]Order, error) {
	rows, err := db.Query(`
		SELECT o.uid, o.port_id, s.ticker, o.quantity, o.value, o.is_buy, o.manually_added, o.date
		FROM orders o
		JOIN portfolios p ON p.id=o.port_id
		JOIN stocks s ON s.id=o.stock_id
		WHERE p.user_id=$1
		ORDER BY o.date`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var orders []Order
	for rows.Next() {
		var o Order
		err = rows.Scan(&o.Uid, &o.PortId, &o.Stock, &o.Quantity, &o.Value, &o.IsBuy, &o.ManuallyAdded, &o.Date)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	if orders == nil {
		return make([]Order, 0), nil
	}
	return orders, nil
}

// TODO refactor this with function above, sharing a ton of similar code
func FetchOrdersByPortfolioId(portId int) ([]Order, error) {
	rows, err := db.Query(`
		SELECT o.uid, o.port_id, s.ticker, o.quantity, o.value, o.is_buy, o.manually_added, o.date
		FROM orders o
		JOIN stocks s ON s.id=o.stock_id
		WHERE o.port_id=$1
		ORDER BY o.date`, portId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var orders []Order
	for rows.Next() {
		var o Order
		err = rows.Scan(&o.Uid, &o.PortId, &o.Stock, &o.Quantity, &o.Value, &o.IsBuy, &o.ManuallyAdded, &o.Date)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	if orders == nil {
		return make([]Order, 0), nil
	}
	return orders, nil
}

// Inserts order if uid doesn't exist already
func InsertIgnoreOrder(o Order) error {
	_, err := db.Exec(`INSERT INTO stocks (ticker) VALUES ($1) ON CONFLICT (ticker) DO NOTHING`, o.Stock)
	if err != nil {
		return err
	}
	_, err = db.Exec(`
		INSERT INTO orders (uid, port_id, stock_id, quantity, value, is_buy, manually_added, date, committed)
			SELECT $1, $2, stocks.id, $4, $5, $6, $7, $8, false
			FROM stocks
			WHERE ticker=$3
		ON CONFLICT(uid) DO NOTHING`,
		o.Uid, o.PortId, o.Stock, o.Quantity, o.Value, o.IsBuy, o.ManuallyAdded, o.Date)
	return err
}

func UpsertOrder(o Order) error {
	err := UpsertStock(o.Stock)
	if err != nil {
		return err
	}
	_, err = db.Exec(`
		INSERT INTO orders (uid, port_id, stock_id, quantity, value, is_buy, manually_added, date, committed)
			SELECT $1, $2, stocks.id, $4, $5, $6, $7, $8, false
			FROM stocks
			WHERE ticker=$3
		ON CONFLICT(uid) DO UPDATE
		SET port_id=$2,stock_id=excluded.stock_id,quantity=$4,value=$5,is_buy=$6,manually_added=$7,date=$8,committed=false`,
		o.Uid, o.PortId, o.Stock, o.Quantity, o.Value, o.IsBuy, o.ManuallyAdded, o.Date)
	return err
}

func DeleteOrder(uid string, portId int) error {
	_, err := db.Exec(`DELETE FROM orders WHERE uid=$1 AND port_id=$2`, uid, portId)
	return err
}

func SetOrdersCommitted(portId int) error {
	_, err := db.Exec(`UPDATE orders SET committed=true WHERE port_id=$1`, portId)
	return err
}

func HasUncommittedOrders(portId int) (bool, error) {
	rows, err := db.Query(`SELECT COUNT(*) FROM orders WHERE committed=false AND port_id=$1`, portId)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	if !rows.Next() {
		return false, fmt.Errorf("no rows returned from count")
	}
	var count int
	err = rows.Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
