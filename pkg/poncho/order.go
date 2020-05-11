package poncho

import (
	"github.com/bluedresscapital/coattails/pkg/stockings"
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
)

type OrderAPI interface {
	GetOrders() ([]wardrobe.Order, error)
}

// Reloads orders via orderAPI, and reloads order dependents if there are changes
// in the orders
func ReloadOrders(order OrderAPI, stock stockings.StockAPI) (bool, error) {
	orders, err := order.GetOrders()
	if err != nil {
		return false, err
	}
	var portId int
	for i, o := range orders {
		portId = o.PortId
		if o.Value.IsZero() {
			price, err := stockings.GetHistoricalPrice(stock, o.Stock, o.Date)
			if err != nil {
				return false, err
			}
			orders[i].Value = *price
		}
		err = wardrobe.InsertIgnoreOrder(o)
		if err != nil {
			return false, err
		}
	}
	port, err := wardrobe.FetchPortfolioById(portId)
	if err != nil {
		return false, err
	}
	maxOrderUpdatedAt, err := wardrobe.GetMaxOrderUpdatedAt(portId)
	if err != nil {
		return false, err
	}
	return port.OrdersUpdatedAt.Before(*maxOrderUpdatedAt), nil
}
