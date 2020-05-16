package orders

import (
	"fmt"
	"log"

	"github.com/bluedresscapital/coattails/pkg/stockings"
	"github.com/bluedresscapital/coattails/pkg/util"
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
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
	for _, o := range orders {
		portId = o.PortId
		// NOTE - assume if our order price is ZERO, that must indicate we transferred the asset
		// If this assumption ever changes, PLEASE UPDATE THIS CODE!!
		if o.Value.IsZero() {
			price, err := stockings.GetHistoricalPrice(stock, o.Stock, o.Date)
			if err != nil {
				log.Printf("Unable to get price for stock %s at date %s, setting it to -1", o.Stock, o.Date)
				o.Value = decimal.New(-1, 0)
			} else {
				o.Value = *price
				// Treat transferred assets as a simple deposit of $(price), and then buying the asset at $(price).
				t := wardrobe.Transfer{
					Uid:           fmt.Sprintf("TRANSFER_FROM_ASSETS__%s", uuid.New().String()),
					PortId:        portId,
					Amount:        o.Quantity.Mul(*price),
					IsDeposit:     true,
					ManuallyAdded: false, // Not sure if this counts as manually adding, but oh well :D
					Date:          util.GetTimelessDate(o.Date),
				}
				err = wardrobe.UpsertTransfer(t)
				if err != nil {
					log.Printf("Error upserting transfer: %v, not erroring out tho", err)
				}
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
