package poncho

import (
	"github.com/bluedresscapital/coattails/pkg/stockings"
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"log"
)

type OrderAPI interface {
	GetOrders() ([]wardrobe.Order, error)
}

func ReloadOrders(order OrderAPI, stock stockings.StockAPI) error {
	orders, err := order.GetOrders()
	if err != nil {
		return err
	}
	var portId int
	for i, o := range orders {
		portId = o.PortId
		if o.Value.IsZero() {
			price, err := stockings.GetHistoricalPrice(stock, o.Stock, o.Date)
			if err != nil {
				return err
			}
			orders[i].Value = *price
		}
		err = wardrobe.InsertIgnoreOrder(o)
		if err != nil {
			return err
		}
	}
	return UpdateOrderDependents(portId, stock)
}

// Updates all dependents of the orders table
func UpdateOrderDependents(portId int, stock stockings.StockAPI) error {
	port, err := wardrobe.FetchPortfolioById(portId)
	if err != nil {
		return err
	}
	maxOrderUpdatedAt, err := wardrobe.GetMaxOrderUpdatedAt(portId)
	if err != nil {
		return err
	}
	if port.OrdersUpdatedAt.Before(*maxOrderUpdatedAt) {
		log.Printf("Detected new changes in portfolio's orders! Updating order depentents!")
		err = ReloadPositions(portId, stock)
		if err != nil {
			return err
		}
	}
	return nil
}
