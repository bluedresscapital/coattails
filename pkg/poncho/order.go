package poncho

import (
	"log"

	"github.com/shopspring/decimal"

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
		log.Printf("Parsing order %s %s", o.Stock, o.Date)
		portId = o.PortId
		if o.Value.IsZero() {
			log.Printf("order's %s value at %s is zero, so checking stockings price...", o.Stock, o.Date)
			price, err := stockings.GetHistoricalPrice(stock, o.Stock, o.Date)
			log.Printf("???")
			if err != nil {
				log.Printf("Unable to get price for stock %s at date %s", o.Stock, o.Date)
				orders[i].Value = decimal.Zero
			} else {
				orders[i].Value = *price
			}
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
