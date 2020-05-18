package robinhood

import (
	"github.com/shopspring/decimal"

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
		return nil, err
	}
	port, err := wardrobe.FetchPortfolioByRHAccountId(api.AccountId)
	if err != nil {
		return nil, err
	}

	res, err := ScrapeOrders(*bearerTok)
	if err != nil {
		return nil, err
	}
	ret := make([]wardrobe.Order, 0)
	for _, o := range res {
		if len(o.Executions) == 0 {
			continue
		}
		shares := decimal.Zero
		amount := decimal.Zero
		for _, e := range o.Executions {
			shares = shares.Add(e.Quantity)
			amount = amount.Add(e.Price.Mul(e.Quantity))
		}
		avgPrice := amount.Div(shares)
		stockP, err := FetchStockFromInstrumentId(o.Instrument)
		if err != nil {
			return nil, err
		}
		ret = append(ret, wardrobe.Order{
			Uid:           o.Id,
			PortId:        port.Id,
			Stock:         *stockP,
			Quantity:      shares,
			Value:         avgPrice,
			IsBuy:         o.Side == "buy",
			ManuallyAdded: false,
			Date:          o.LastTransactionAt,
		})
	}
	return ret, nil
}

func (api API) GetTransfers() ([]wardrobe.Transfer, error) {
	bearerTok, err := api.getAuthToken()
	if err != nil {
		return nil, err
	}
	port, err := wardrobe.FetchPortfolioByRHAccountId(api.AccountId)
	if err != nil {
		return nil, err
	}
	res, err := ScrapeTransfers(*bearerTok)
	if err != nil {
		return nil, err
	}
	ret := make([]wardrobe.Transfer, 0)
	for _, t := range res {
		ret = append(ret, wardrobe.Transfer{
			Uid:           t.Id,
			PortId:        port.Id,
			Amount:        t.Amount,
			IsDeposit:     t.IsDeposit,
			ManuallyAdded: false,
			Date:          t.Date,
		})

	}
	return ret, nil
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
