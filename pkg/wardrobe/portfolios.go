package wardrobe

import (
	"database/sql"
	"fmt"
	"time"
)

type Portfolio struct {
	Id                 int       `json:"id"`
	Name               string    `json:"name"`
	Type               string    `json:"type"`
	UserId             int       `json:"user_id"`
	TDAccountId        int       `json:"tda_account_id"`
	OrdersUpdatedAt    time.Time `json:"orders_updated_at"`
	TransfersUpdatedAt time.Time `json:"transfers_updated_at"`
}

func CreatePortfolio(userId int, name string, portType string) error {
	_, err := db.Exec("INSERT INTO portfolios (user_id, name, type) VALUES ($1,$2,$3)", userId, name, portType)
	return err
}

func FetchPortfolioById(id int) (*Portfolio, error) {
	rows, err := db.Query(`
		SELECT id, name, type, user_id, tda_account_id, orders_updated_at, transfers_updated_at 
		FROM portfolios WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, fmt.Errorf("no portfolio with id %d found", id)
	}
	var port Portfolio
	var tdAccountId sql.NullInt64
	var ordersUpdatedAt sql.NullTime
	var transfersUpdatedAt sql.NullTime
	err = rows.Scan(&port.Id, &port.Name, &port.Type, &port.UserId, &tdAccountId, &ordersUpdatedAt, &transfersUpdatedAt)
	if err != nil {
		return nil, err
	}
	if tdAccountId.Valid {
		port.TDAccountId = int(tdAccountId.Int64)
	}
	if ordersUpdatedAt.Valid {
		port.OrdersUpdatedAt = ordersUpdatedAt.Time
	}
	if transfersUpdatedAt.Valid {
		port.TransfersUpdatedAt = transfersUpdatedAt.Time
	}
	if rows.Next() {
		return nil, fmt.Errorf("multiple portfolios found with id %d", id)
	}
	return &port, nil
}

func FetchPortfolioByTDAccountId(tdAccountId int) (*Portfolio, error) {
	rows, err := db.Query("SELECT id, name, type, user_id FROM portfolios WHERE tda_account_id=$1", tdAccountId)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, fmt.Errorf("no portfolio with tda_account_id %d found", tdAccountId)
	}
	var port Portfolio
	err = rows.Scan(&port.Id, &port.Name, &port.Type, &port.UserId)
	if err != nil {
		return nil, err
	}
	port.TDAccountId = tdAccountId
	if rows.Next() {
		return nil, fmt.Errorf("multiple portfolios found with tda_account_id %d", tdAccountId)
	}
	return &port, nil
}

func FetchPortfoliosByUserId(userId int) ([]Portfolio, error) {
	rows, err := db.Query("SELECT id, name, type, user_id, tda_account_id FROM portfolios WHERE user_id=$1", userId)
	if err != nil {
		return nil, err
	}
	var ports []Portfolio
	for rows.Next() {
		var port Portfolio
		var tdAccountId sql.NullInt64
		err = rows.Scan(&port.Id, &port.Name, &port.Type, &port.UserId, &tdAccountId)
		if err != nil {
			return nil, err
		}
		if tdAccountId.Valid {
			port.TDAccountId = int(tdAccountId.Int64)
		}
		ports = append(ports, port)
	}
	if ports == nil {
		return make([]Portfolio, 0), nil
	}
	return ports, nil
}

func UpdatePortfolioOrderTransferUpdatedAt(portId int, orderUpdatedAt time.Time, transferUpdatedAt time.Time) error {
	_, err := db.Exec(`UPDATE portfolios SET orders_updated_at=$1, transfers_updated_at=$2 WHERE id=$3`,
		orderUpdatedAt, transferUpdatedAt, portId)
	return err
}
