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
	if err != nil {
		return err
	}
	return nil
}

func FetchPortfolio(userId int, name string, portType string) (*int, error) {
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
	return &id, nil
}

func FetchPortfoliosByUserId(userId int) ([]Portfolio, error) {
	rows, err := db.Query("SELECT id, name, type FROM portfolios WHERE user_id=$1", userId)
	if err != nil {
		return nil, err
	}
	var ports []Portfolio
	for rows.Next() {
		var port Portfolio
		err = rows.Scan(&port.Id, &port.Name, &port.Type)
		if err != nil {
			return nil, err
		}
		port.UserId = userId
		ports = append(ports, port)
	}
	return ports, nil
}
