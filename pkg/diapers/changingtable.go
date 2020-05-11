package diapers

import (
	"fmt"
	"github.com/bluedresscapital/coattails/pkg/poncho"
	"github.com/bluedresscapital/coattails/pkg/socks"
	"github.com/bluedresscapital/coattails/pkg/stockings"
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"log"
)

type Data string

const (
	Order    Data = "order"
	Transfer Data = "transfer"
	Position Data = "position"
)

var depMap map[Data][]Data

func init() {
	depMap = map[Data][]Data{
		Order: {
			Position,
		},
		Transfer: {
			Position,
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
	}
	return nil
}

func reloadPositionsAndPublish(portId int, userId int, channel string) error {
	err := poncho.ReloadPositions(portId, stockings.IexApi{})
	if err != nil {
		log.Printf("Error reloading positions: %v", err)
		return err
	}
	positions, err := wardrobe.FetchPositions(userId)
	if err != nil {
		return err
	}
	return socks.PublishFromServer(channel, "LOADED_POSITIONS", positions)
}
