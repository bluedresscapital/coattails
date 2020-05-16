package wardrobe

import (
	"github.com/shopspring/decimal"
)

type Position struct {
	PortId   int             `json:"port_id"`
	Quantity decimal.Decimal `json:"quantity"`
	Value    decimal.Decimal `json:"value"`
	Stock    string          `json:"stock"`
}

func FetchPositions(userId int) ([]Position, error) {
	rows, err := db.Query(`
		SELECT p.port_id, p.quantity, p.value, s.ticker
		FROM positions p
		JOIN portfolios port ON p.port_id=port.id
		JOIN stocks s ON s.id=p.stock_id
		WHERE port.user_id=$1
	`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var positions []Position
	for rows.Next() {
		var p Position
		err = rows.Scan(&p.PortId, &p.Quantity, &p.Value, &p.Stock)
		if err != nil {
			return nil, err
		}
		positions = append(positions, p)
	}
	if positions == nil {
		return make([]Position, 0), nil
	}
	return positions, nil
}

func FetchPortfolioPositions(portId int) ([]Position, error) {
	rows, err := db.Query(`
		SELECT p.port_id, p.quantity, p.value, s.ticker
		FROM positions p
		JOIN stocks s ON s.id=p.stock_id
		WHERE p.port_id=$1`, portId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var positions []Position
	for rows.Next() {
		var p Position
		err = rows.Scan(&p.PortId, &p.Quantity, &p.Value, &p.Stock)
		if err != nil {
			return nil, err
		}
		positions = append(positions, p)
	}
	if positions == nil {
		return make([]Position, 0), nil
	}
	return positions, nil
}

func InsertPosition(p Position) error {
	err := UpsertStock(p.Stock)
	if err != nil {
		return err
	}
	_, err = db.Exec(`
		INSERT INTO positions (port_id, stock_id, quantity, value)
			SELECT $1, s.id, $3, $4
			FROM stocks s
			WHERE ticker=$2
		`, p.PortId, p.Stock, p.Quantity, p.Value)
	return err
}

func DeletePositions(portId int) error {
	_, err := db.Exec(`DELETE FROM positions WHERE port_id=$1`, portId)
	return err
}
