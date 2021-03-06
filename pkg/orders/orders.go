package orders

import (
	"fmt"
	"log"

	"github.com/bluedresscapital/coattails/pkg/stockings"
	"github.com/bluedresscapital/coattails/pkg/util"
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
	if orders == nil || len(orders) == 0 {
		return false, nil
	}
	var portId int
	for _, o := range orders {
		portId = o.PortId
		err = wardrobe.InsertIgnoreOrder(o)
		if err != nil {
			return false, err
		}
	}
	zeroValOrders, err := wardrobe.FetchZeroPriceOrdersByPortfolioId(portId)
	if len(zeroValOrders) > 0 {
		log.Print("Detected NEW orders with zero values, fetching stock prices for them...")
		for _, o := range zeroValOrders {
			// NOTE - assume if our order price is ZERO, that must indicate we transferred the asset
			// If this assumption ever changes, PLEASE UPDATE THIS CODE!!
			if o.Value.IsZero() {
				price, err := stockings.GetHistoricalPrice(stock, o.Stock, o.Date)
				if err != nil {
					return false, fmt.Errorf("unable to get a price for stock %s at date %s: %v", o.Stock, o.Date, err)
				}
				if (*price).IsZero() {
					return false, fmt.Errorf("price for %s on %s is still 0, erroring out", o.Stock, o.Date)
				}
				o.Value = *price
				// Treat transferred assets as a simple deposit of $(price), and then buying the asset at $(price).
				t := wardrobe.Transfer{
					Uid:           fmt.Sprintf("TRANSFER_FROM_ASSETS__%s", o.Uid),
					PortId:        portId,
					Amount:        o.Quantity.Mul(*price),
					IsDeposit:     true,
					ManuallyAdded: false, // Not sure if this counts as manually adding, but oh well :D
					Date:          util.GetTimelessDate(o.Date),
				}
				err = wardrobe.InsertIgnoreTransfer(t)
				if err != nil {
					log.Printf("Error upserting transfer: %v, not erroring out tho", err)
				}
				err = wardrobe.UpsertOrder(o)
				if err != nil {
					log.Printf("Error upserting order: %v", err)
				}
			}
		}
	}
	return wardrobe.HasUncommittedOrders(portId)
}
