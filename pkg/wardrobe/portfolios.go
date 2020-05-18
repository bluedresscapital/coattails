package wardrobe

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

type Portfolio struct {
	Id                 int       `json:"id"`
	Name               string    `json:"name"`
	Type               string    `json:"type"`
	UserId             int       `json:"user_id"`
	TDAccountId        int       `json:"tda_account_id"`
	RHAccountId        int       `json:"rh_account_id"`
	OrdersUpdatedAt    time.Time `json:"orders_updated_at"`
	TransfersUpdatedAt time.Time `json:"transfers_updated_at"`
}

func CreatePortfolio(userId int, name string, portType string) error {
	_, err := db.Exec("INSERT INTO portfolios (user_id, name, type) VALUES ($1,$2,$3)", userId, name, portType)
	return err
}

func FetchPortfolioById(id int) (*Portfolio, error) {
	rows, err := db.Query(`
		SELECT id, name, type, user_id, tda_account_id, rh_account_id, orders_updated_at, transfers_updated_at 
		FROM portfolios WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, fmt.Errorf("no portfolio with id %d found", id)
	}
	var port Portfolio
	var tdAccountId sql.NullInt64
	var rhAccountId sql.NullInt64
	var ordersUpdatedAt sql.NullTime
	var transfersUpdatedAt sql.NullTime
	err = rows.Scan(&port.Id, &port.Name, &port.Type, &port.UserId, &tdAccountId, &rhAccountId, &ordersUpdatedAt, &transfersUpdatedAt)
	if err != nil {
		return nil, err
	}
	if tdAccountId.Valid {
		port.TDAccountId = int(tdAccountId.Int64)
	}
	if rhAccountId.Valid {
		port.RHAccountId = int(rhAccountId.Int64)
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
	defer rows.Close()
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

func FetchPortfolioByRHAccountId(rhAccountId int) (*Portfolio, error) {
	rows, err := db.Query("SELECT id, name, type, user_id FROM portfolios WHERE rh_account_id=$1", rhAccountId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, fmt.Errorf("no portfolio with rh_account_id %d found", rhAccountId)
	}
	var port Portfolio
	err = rows.Scan(&port.Id, &port.Name, &port.Type, &port.UserId)
	if err != nil {
		return nil, err
	}
	port.RHAccountId = rhAccountId
	if rows.Next() {
		return nil, fmt.Errorf("multiple portfolios found with rh_account_id %d", rhAccountId)
	}
	return &port, nil
}

func FetchPortfoliosByUserId(userId int) ([]Portfolio, error) {
	rows, err := db.Query("SELECT id, name, type, user_id, tda_account_id, rh_account_id FROM portfolios WHERE user_id=$1", userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ports []Portfolio
	for rows.Next() {
		var port Portfolio
		var tdAccountId sql.NullInt64
		var rhAccountId sql.NullInt64
		err = rows.Scan(&port.Id, &port.Name, &port.Type, &port.UserId, &tdAccountId, &rhAccountId)
		if err != nil {
			return nil, err
		}
		if tdAccountId.Valid {
			port.TDAccountId = int(tdAccountId.Int64)
		}
		if rhAccountId.Valid {
			port.RHAccountId = int(rhAccountId.Int64)
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

type PortValue struct {
	PortId            int
	Date              time.Time
	DailyNetDeposited decimal.Decimal
	Cash              decimal.Decimal
	StockValue        decimal.Decimal
	NormalizedCash    decimal.Decimal
}

func UpsertPortfolioValue(pv PortValue) error {
	_, err := db.Exec(`
		INSERT INTO portfolio_values (port_id, cash, stock_value, daily_net_deposited, normalized_cash, date)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (port_id, date) DO UPDATE
		SET cash=$2, stock_value=$3, daily_net_deposited=$4, normalized_cash=$5`,
		pv.PortId, pv.Cash, pv.StockValue, pv.DailyNetDeposited, pv.NormalizedCash, pv.Date)
	return err
}

func FetchPortfolioValuesByPortId(portId int) ([]PortValue, error) {
	rows, err := db.Query(`
		SELECT port_id, cash, stock_value, daily_net_deposited, normalized_cash, date
		FROM portfolio_values
		WHERE port_id=$1
		ORDER BY date`, portId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	pvs := make([]PortValue, 0)
	for rows.Next() {
		var pv PortValue
		err = rows.Scan(&pv.PortId, &pv.Cash, &pv.StockValue, &pv.DailyNetDeposited, &pv.NormalizedCash, &pv.Date)
		if err != nil {
			return nil, err
		}
		pvs = append(pvs, pv)
	}
	return pvs, nil
}
