package wardrobe

import (
	"fmt"
)

type Portfolio struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	UserId int    `json:"user_id"`
}

func CreatePortfolio(userId int, name string, portType string) error {
	_, err := db.Exec("INSERT INTO portfolios (user_id, name, type) VALUES ($1,$2,$3)", userId, name, portType)
	return err
}

func FetchPortfolio(userId int, name string, portType string) (*Portfolio, error) {
	rows, err := db.Query("SELECT id FROM portfolios WHERE user_id=$1 and name=$2 and type=$3", userId, name, portType)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, fmt.Errorf("no portfolio with name %s and type %s found for user %d", name, portType, userId)
	}
	var id int
	err = rows.Scan(&id)
	if err != nil {
		return nil, err
	}
	if rows.Next() {
		return nil, fmt.Errorf("multiple portfolios found with name %s and type %s for user %d", name, portType, userId)
	}
	return &Portfolio{
		Id:     id,
		Name:   name,
		Type:   portType,
		UserId: userId,
	}, nil
}

func FetchPortfolioById(id int) (*Portfolio, error) {
	rows, err := db.Query("SELECT id, name, type, user_id FROM portfolios WHERE id=$1", id)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, fmt.Errorf("no portfolio with id %d found", id)
	}
	var port Portfolio
	err = rows.Scan(&port.Id, &port.Name, &port.Type, &port.UserId)
	if err != nil {
		return nil, err
	}
	if rows.Next() {
		return nil, fmt.Errorf("multiple portfolios found with id %d", id)
	}
	return &port, nil
}

func FetchPortfoliosByUserId(userId int) ([]Portfolio, error) {
	rows, err := db.Query("SELECT id, name, type, user_id FROM portfolios WHERE user_id=$1", userId)
	if err != nil {
		return nil, err
	}
	var ports []Portfolio
	for rows.Next() {
		var port Portfolio
		err = rows.Scan(&port.Id, &port.Name, &port.Type, &port.UserId)
		if err != nil {
			return nil, err
		}
		ports = append(ports, port)
	}
	if ports == nil {
		return make([]Portfolio, 0), nil
	}
	return ports, nil
}
