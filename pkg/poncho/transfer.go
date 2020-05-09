package poncho

import "github.com/bluedresscapital/coattails/pkg/wardrobe"

type TransferAPI interface {
	GetTransfers() ([]wardrobe.Transfer, error)
}

func ReloadTransfers(transfer TransferAPI) error {
	transfers, err := transfer.GetTransfers()
	if err != nil {
		return err
	}
	for _, t := range transfers {
		err = wardrobe.UpsertTransfer(t)
		if err != nil {
			return err
		}
	}
	return nil
}
