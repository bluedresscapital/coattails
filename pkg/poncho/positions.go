package poncho

import (
	"log"

	"github.com/bluedresscapital/coattails/pkg/stockings"
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"github.com/shopspring/decimal"
)

func ReloadPositions(portId int, stockAPI stockings.StockAPI) error {
	log.Printf("Reloading positions for port %d", portId)
	orders, err := wardrobe.FetchOrdersByPortfolioId(portId)
	if err != nil {
		return err
	}
	transfers, err := wardrobe.FetchTransfersByPortfolioId(portId)
	if err != nil {
		return err
	}
	// Compute total cash deposited
	cash := decimal.Zero
	for _, t := range transfers {
		if t.IsDeposit {
			cash = cash.Add(t.Amount)
		} else {
			cash = cash.Sub(t.Amount)
		}
	}
	// Compute stock positions
	port := make(map[string]decimal.Decimal)
	for _, o := range orders {
		_, found := port[o.Stock]
		if !found {
			port[o.Stock] = decimal.Zero
		}
		if o.IsBuy {
			port[o.Stock] = port[o.Stock].Add(o.Quantity)
			cash = cash.Sub(o.Quantity.Mul(o.Value))
		} else {
			port[o.Stock] = port[o.Stock].Sub(o.Quantity)
			cash = cash.Add(o.Quantity.Mul(o.Value))
		}
	}
	// Delete ALL positions for this portfolio, and re-insert
	// We need to do this in case we delete an order, and we don't track that anymore in the previous portfolio's
	// positions.
	err = wardrobe.DeletePositions(portId)
	if err != nil {
		return err
	}

	// Upsert cash position
	err = wardrobe.InsertPosition(wardrobe.Position{
		PortId:   portId,
		Quantity: decimal.NewFromInt(1),
		Value:    cash,
		Stock:    "_CASH",
	})
	if err != nil {
		return err
	}
	// Upsert stock positions
	for stock, quantity := range port {
		var price decimal.Decimal
		priceP, err := stockings.GetCurrentPrice(stockAPI, stock)
		if err != nil {
			log.Printf("Errored in finding current price for %s: %v", stock, err)
			price = decimal.Zero
		} else {
			price = *priceP
		}
		p := wardrobe.Position{
			PortId:   portId,
			Quantity: quantity,
			Value:    quantity.Mul(price),
			Stock:    stock,
		}
		err = wardrobe.InsertPosition(p)
		if err != nil {
			return err
		}
	}
	// Update portfolio's order and transfer updated at
	maxOrderUpdatedAt, err := wardrobe.GetMaxOrderUpdatedAt(portId)
	if err != nil {
		return err
	}
	maxTransferUpdatedAt, err := wardrobe.GetMaxTransferUpdatedAt(portId)
	if err != nil {
		return err
	}
	return wardrobe.UpdatePortfolioOrderTransferUpdatedAt(portId, *maxOrderUpdatedAt, *maxTransferUpdatedAt)
}
