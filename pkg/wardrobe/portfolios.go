package wardrobe

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"

	"github.com/shopspring/decimal"
)

type Portfolio struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	UserId      int    `json:"user_id"`
	TDAccountId int    `json:"tda_account_id"`
	RHAccountId int    `json:"rh_account_id"`
}

func CreatePortfolio(userId int, name string, portType string) error {
	_, err := db.Exec("INSERT INTO portfolios (user_id, name, type) VALUES ($1,$2,$3)", userId, name, portType)
	return err
}

func FetchPortfolioById(id int) (*Portfolio, error) {
	rows, err := db.Query(`
		SELECT id, name, type, user_id, tda_account_id, rh_account_id
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

type PortValue struct {
	PortId            int             `json:"port_id"`
	Date              time.Time       `json:"date"`
	DailyNetDeposited decimal.Decimal `json:"daily_net_deposited"`
	Cash              decimal.Decimal `json:"cash"`
	StockValue        decimal.Decimal `json:"stock_value"`
	NormalizedCash    decimal.Decimal `json:"normalized_cash"`
	CumChange         decimal.Decimal `json:"cum_change"`
	DailyChange       decimal.Decimal `json:"daily_change"`
}

func BulkUpsertPortfolioValuesByPortId(pvs []PortValue, portId int) error {
	if len(pvs) == 0 {
		return nil
	}
	txn, err := db.Begin()
	if err != nil {
		return err
	}
	start := pvs[0].Date
	end := pvs[len(pvs)-1].Date
	_, err = txn.Exec(`DELETE FROM portfolio_values WHERE port_id =$1 AND date >= $2 AND date <= $3`, portId, start, end)
	if err != nil {
		txn.Rollback()
		return err
	}
	stmt, _ := txn.Prepare(pq.CopyIn("portfolio_values", "port_id", "cash", "stock_value", "daily_net_deposited", "normalized_cash", "date", "cum_change", "daily_change"))
	if err != nil {
		return err
	}
	for _, pv := range pvs {
		_, err = stmt.Exec(pv.PortId, pv.Cash, pv.StockValue, pv.DailyNetDeposited, pv.NormalizedCash, pv.Date, pv.CumChange, pv.DailyChange)
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

func UpsertPortfolioValue(pv PortValue) error {
	_, err := db.Exec(`
		INSERT INTO portfolio_values (port_id, cash, stock_value, daily_net_deposited, normalized_cash, date, cum_change, daily_change)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (port_id, date) DO UPDATE
		SET cash=$2, stock_value=$3, daily_net_deposited=$4, normalized_cash=$5, cum_change=$7, daily_change=$8`,
		pv.PortId, pv.Cash, pv.StockValue, pv.DailyNetDeposited, pv.NormalizedCash, pv.Date, pv.CumChange, pv.DailyChange)
	return err
}

func FetchPortfolioValuesByPortId(portId int) ([]PortValue, error) {
	rows, err := db.Query(`
		SELECT port_id, cash, stock_value, daily_net_deposited, normalized_cash, date, cum_change, daily_change
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
		err = rows.Scan(&pv.PortId, &pv.Cash, &pv.StockValue, &pv.DailyNetDeposited, &pv.NormalizedCash, &pv.Date, &pv.CumChange, &pv.DailyChange)
		if err != nil {
			return nil, err
		}
		pvs = append(pvs, pv)
	}
	return pvs, nil
}

func FetchPortfolioValueOnDay(portId int, date time.Time) (*PortValue, error) {
	rows, err := db.Query(`
		SELECT port_id, cash, stock_value, daily_net_deposited, normalized_cash, date, cum_change, daily_change
		FROM portfolio_values
		WHERE port_id=$1 AND date=$2`, portId, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, fmt.Errorf("no portfolio value found for port %d on %s", portId, date)
	}
	var pv PortValue
	err = rows.Scan(&pv.PortId, &pv.Cash, &pv.StockValue, &pv.DailyNetDeposited, &pv.NormalizedCash, &pv.Date, &pv.CumChange, &pv.DailyChange)
	if err != nil {
		return nil, err
	}
	return &pv, nil
}

func FetchAllPortfolioIds() ([]int, error) {
	rows, err := db.Query(`SELECT id FROM portfolios`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := make([]int, 0)
	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		ids = append(ids, id)
	}
	return ids, nil
}

type DailyPortVal struct {
	PortId int             `json:"port_id"`
	Date   time.Time       `json:"date"`
	Value  decimal.Decimal `json:"value"`
}

func InsertDailyPortValue(dpv DailyPortVal) error {
	_, err := db.Exec(`INSERT INTO daily_portfolio_values (port_id, date, value) VALUES ($1, $2, $3)`,
		dpv.PortId, dpv.Date, dpv.Value)
	return err
}
