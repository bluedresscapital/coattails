package robinhood

import (
	"log"

	"github.com/bluedresscapital/coattails/pkg/orders"
	"github.com/bluedresscapital/coattails/pkg/transfers"
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
)

type API struct {
	AccountId int
}

var _ orders.OrderAPI = (*API)(nil)
var _ transfers.TransferAPI = (*API)(nil)

func (api API) GetOrders() ([]wardrobe.Order, error) {
	bearerTok, err := api.getAuthToken()
	if err != nil {
		log.Printf("Errored out: %v", err)
		return nil, err
	}
	ScrapeOrders(*bearerTok)
	return nil, nil
}

func (api API) GetTransfers() ([]wardrobe.Transfer, error) {
	return nil, nil
}

func (api API) getAuthToken() (*string, error) {
	acc, err := wardrobe.FetchRHAccount(api.AccountId)
	if err != nil {
		return nil, err
	}
	auth, err := FetchBearerToken(acc.RefreshTok)
	if err != nil {
		return nil, err
	}
	err = wardrobe.UpdateRHRefreshToken(api.AccountId, auth.RefreshTok)
	if err != nil {
		return nil, err
	}
	return &(auth.BearerTok), nil
}
