package wardrobe

import (
	"fmt"
	"github.com/shopspring/decimal"
	"time"
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
		WHERE p.user_id=$1`, userId)
	if err != nil {
		return nil, err
	}
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

func FetchOrderByUid(uid string) (*Order, error) {
	rows, err := db.Query(`
		SELECT o.uid, o.port_id, s.ticker, o.quantity, o.value, o.is_buy, o.manually_added, o.date 
		FROM orders o
		JOIN stocks s ON o.stock_id=s.id
		WHERE uid=$1`, uid)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, fmt.Errorf("unable to find order with uid %s", uid)
	}
	var o Order
	err = rows.Scan(&o.Uid, &o.PortId, &o.Stock, &o.Quantity, &o.Value, &o.IsBuy, &o.ManuallyAdded, &o.Date)
	if err != nil {
		return nil, err
	}
	if rows.Next() {
		return nil, fmt.Errorf("multiple transfers found with uid %s", uid)
	}
	return &o, nil
}

func UpsertOrder(o Order) error {
	_, err := db.Exec(`INSERT INTO stocks (ticker) VALUES ($1) ON CONFLICT (ticker) DO NOTHING`, o.Stock)
	if err != nil {
		return err
	}
	_, err = db.Exec(`
		INSERT INTO orders (uid, port_id, stock_id, quantity, value, is_buy, manually_added, date)
			SELECT $1, $2, stocks.id, $4, $5, $6, $7, $8
			FROM stocks
			WHERE ticker=$3
		ON CONFLICT(uid) DO UPDATE
		SET port_id=$2,stock_id=excluded.stock_id,quantity=$4,value=$5,is_buy=$6,manually_added=$7,date=$8`,
		o.Uid, o.PortId, o.Stock, o.Quantity, o.Value, o.IsBuy, o.ManuallyAdded, o.Date)
	return err
}

func DeleteOrder(uid string, portId int) error {
	_, err := db.Exec(`DELETE FROM orders WHERE uid=$1 AND port_id=$2`, uid, portId)
	return err
}
