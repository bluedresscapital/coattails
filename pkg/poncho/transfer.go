package poncho

import (
	"github.com/bluedresscapital/coattails/pkg/stockings"
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"log"
)

type TransferAPI interface {
	GetTransfers() ([]wardrobe.Transfer, error)
}

func ReloadTransfers(transfer TransferAPI, stock stockings.StockAPI) error {
	transfers, err := transfer.GetTransfers()
	if err != nil {
		return err
	}
	var portId int
	for _, t := range transfers {
		portId = t.PortId
		err = wardrobe.InsertIgnoreTransfer(t)
		if err != nil {
			return err
		}
	}
	return UpdateTransferDependents(portId, stock)
}

// Updates all dependents of the orders table
func UpdateTransferDependents(portId int, stock stockings.StockAPI) error {
	port, err := wardrobe.FetchPortfolioById(portId)
	if err != nil {
		return err
	}
	maxTransferUpdatedAt, err := wardrobe.GetMaxTransferUpdatedAt(portId)
	if err != nil {
		return err
	}
	if port.TransfersUpdatedAt.Before(*maxTransferUpdatedAt) {
		log.Printf("Detected new changes in portfolio's transfers! Updating order depentents!")
		err = ReloadPositions(portId, stock)
		if err != nil {
			return err
		}
	}
	return nil
}
