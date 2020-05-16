package transfers

import (
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
)

type TransferAPI interface {
	GetTransfers() ([]wardrobe.Transfer, error)
}

// Reloads transfers from TransferAPI - If there are changes, it will also
// return whether it should be updated
func ReloadTransfers(transfer TransferAPI) (bool, error) {
	transfers, err := transfer.GetTransfers()
	if err != nil {
		return false, err
	}
	var portId int
	for _, t := range transfers {
		portId = t.PortId
		err = wardrobe.InsertIgnoreTransfer(t)
		if err != nil {
			return false, err
		}
	}
	port, err := wardrobe.FetchPortfolioById(portId)
	if err != nil {
		return false, err
	}
	maxTransferUpdatedAt, err := wardrobe.GetMaxTransferUpdatedAt(portId)
	if err != nil {
		return false, err
	}
	return port.TransfersUpdatedAt.Before(*maxTransferUpdatedAt), nil
}
