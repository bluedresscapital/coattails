package diapers

import (
	"fmt"
	"log"

	"github.com/bluedresscapital/coattails/pkg/portfolios"

	"github.com/bluedresscapital/coattails/pkg/positions"
	"github.com/bluedresscapital/coattails/pkg/socks"
	"github.com/bluedresscapital/coattails/pkg/stockings"
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
)

type Data string

const (
	Order     Data = "order"
	Transfer  Data = "transfer"
	Position  Data = "position"
	Portfolio Data = "portfolio"
)

var depMap map[Data][]Data

func init() {
	depMap = map[Data][]Data{
		Order: {
			Position,
			Portfolio,
		},
		Transfer: {
			Position,
			Portfolio,
		},
	}
}

func ReloadDepsAndPublish(data Data, portId int, userId int, channel string) error {
	deps, found := depMap[data]
	if !found {
		return fmt.Errorf("no callbacks for data %v", data)
	}
	for _, dep := range deps {
		if dep == Position {
			return reloadPositionsAndPublish(portId, userId, channel)
		}
		if dep == Portfolio {
			return reloadPortfolioAndPublish(portId, userId, channel)
		}
	}
	switch data {
	case Order:
		log.Printf("Committing orders for port %d", portId)
		err := wardrobe.SetOrdersCommitted(portId)
		if err != nil {
			return err
		}
	case Transfer:
		log.Printf("Committing transfers for port %d", portId)
		err := wardrobe.SetTransfersCommitted(portId)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported dep change: %v", data)
	}
	return nil
}

// Given a list of data changes, figures out what downstream data we need to reload (just once)
func BulkReloadDepsAndPublish(data []Data, portId int, userId int, channel string) error {
	log.Printf("changing table bulk reloading %v", data)
	depSet := make(map[Data]bool)
	for _, d := range data {
		deps, found := depMap[d]
		if found {
			for _, dep := range deps {
				depSet[dep] = true
			}
		}
	}
	for d := range depSet {
		switch d {
		case Position:
			err := reloadPositionsAndPublish(portId, userId, channel)
			if err != nil {
				return err
			}
		case Portfolio:
			err := reloadPortfolioAndPublish(portId, userId, channel)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported data change: %v", d)
		}
	}
	for _, d := range data {
		switch d {
		case Order:
			log.Printf("Committing orders for port %d", portId)
			err := wardrobe.SetOrdersCommitted(portId)
			if err != nil {
				return err
			}
		case Transfer:
			log.Printf("Committing transfers for port %d", portId)
			err := wardrobe.SetTransfersCommitted(portId)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported dep change: %v", d)
		}
	}
	return nil
}

func reloadPositionsAndPublish(portId int, userId int, channel string) error {
	err := positions.Reload(portId, stockings.FingoPack{})
	if err != nil {
		log.Printf("Error reloading positions: %v", err)
		return err
	}
	p, err := wardrobe.FetchPositions(userId)
	if err != nil {
		return err
	}
	return socks.PublishFromServer(channel, "LOADED_POSITIONS", p)
}

func reloadPortfolioAndPublish(portId int, userId int, channel string) error {
	port, err := wardrobe.FetchPortfolioById(portId)
	if err != nil {
		return err
	}
	err = portfolios.ReloadHistory(*port)
	if err != nil {
		return err
	}
	// TODO add a socket here!
	return nil
}
