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
	if transfers == nil || len(transfers) == 0 {
		return false, nil
	}
	var portId int
	for _, t := range transfers {
		portId = t.PortId
		err = wardrobe.InsertIgnoreTransfer(t)
		if err != nil {
			return false, err
		}
	}
	return wardrobe.HasUncommittedTransfers(portId)
}
