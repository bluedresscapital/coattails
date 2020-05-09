package poncho

import (
	"github.com/bluedresscapital/coattails/pkg/stockings"
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
)

type OrderAPI interface {
	GetOrders() ([]wardrobe.Order, error)
}

func ReloadOrders(order OrderAPI, stock stockings.StockAPI) error {
	orders, err := order.GetOrders()
	if err != nil {
		return err
	}
	for i, o := range orders {
		if o.Value.IsZero() {
			price, err := stockings.GetHistoricalPrice(stock, o.Stock, o.Date)
			if err != nil {
				return err
			}
			orders[i].Value = *price
		}
		err = wardrobe.UpsertOrder(o)
		if err != nil {
			return err
		}
	}
	// Just fact checking here
	// this is basically our positions code - saving this for later :)
	//port := make(map[string]decimal.Decimal)
	//for _, o := range orders {
	//	_, found := port[o.Stock]
	//	if !found {
	//		port[o.Stock] = decimal.Zero
	//	}
	//	if o.IsBuy {
	//		port[o.Stock] = port[o.Stock].Add(o.Quantity)
	//	} else {
	//		port[o.Stock] = port[o.Stock].Sub(o.Quantity)
	//	}
	//}
	//for a, b := range port {
	//	if !b.IsZero() {
	//		log.Printf("HW still owns %s shares of %s", b.String(), a)
	//	}
	//}
	return nil
}
