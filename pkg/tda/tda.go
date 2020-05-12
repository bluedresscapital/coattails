package tda

import (
	"strconv"
	"time"

	"github.com/bluedresscapital/coattails/pkg/poncho"
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"github.com/shopspring/decimal"
)

type API struct {
	AccountId int
}

var _ poncho.OrderAPI = (*API)(nil)
var _ poncho.TransferAPI = (*API)(nil)

func (api API) GetOrders() ([]wardrobe.Order, error) {
	accessTok, tdAccount, err := api.getAccessToken()
	if err != nil {
		return nil, err
	}
	trans, err := ScrapeTransactions(*accessTok, tdAccount.AccountNum)
	if err != nil {
		return nil, err
	}
	port, err := wardrobe.FetchPortfolioByTDAccountId(api.AccountId)
	if err != nil {
		return nil, err
	}
	var orders []wardrobe.Order
	for _, t := range trans {
		var isBuy bool
		var price decimal.Decimal
		if t.Type == "RECEIVE_AND_DELIVER" && t.Description == "TRANSFER OF SECURITY OR OPTION IN" {
			// This is for assets that were transferred into the account, so they don't have a starting price
			// For now, we just set their cost basis as whatever the price of the stock was that day, but if
			// we wanted to be more pedantic we could allow the user to manually edit it
			isBuy = true
			price = decimal.Zero
		} else if t.Type == "TRADE" {
			isBuy = t.TransactionItem.Instruction == "BUY"
			price = t.TransactionItem.Price
		} else {
			// Don't even add the order if it isn't one of the categories listed above
			continue
		}
		date, err := time.Parse("2006-01-02T15:04:05+0000", t.TransactionDate)
		if err != nil {
			return nil, err
		}
		order := wardrobe.Order{
			Uid:           strconv.Itoa(t.TransactionId),
			PortId:        port.Id,
			Stock:         t.TransactionItem.Instrument.Symbol,
			Quantity:      t.TransactionItem.Amount,
			Value:         price,
			IsBuy:         isBuy,
			ManuallyAdded: false,
			Date:          date,
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func (api API) GetTransfers() ([]wardrobe.Transfer, error) {
	accessTok, tdAccount, err := api.getAccessToken()
	if err != nil {
		return nil, err
	}
	trans, err := ScrapeTransactions(*accessTok, tdAccount.AccountNum)
	if err != nil {
		return nil, err
	}
	port, err := wardrobe.FetchPortfolioByTDAccountId(api.AccountId)
	if err != nil {
		return nil, err
	}
	var transfers []wardrobe.Transfer
	for _, t := range trans {
		if t.Type == "ELECTRONIC_FUND" ||
			(t.Type == "JOURNAL" && t.Description == "CASH MOVEMENT OF INCOMING ACCOUNT TRANSFER") {
			date, err := time.Parse("2006-01-02T15:04:05+0000", t.TransactionDate)
			if err != nil {
				return nil, err
			}
			transfer := wardrobe.Transfer{
				Uid:           strconv.Itoa(t.TransactionId),
				PortId:        port.Id,
				Amount:        t.NetAmount,
				IsDeposit:     t.NetAmount.IsZero() || t.NetAmount.IsPositive(),
				ManuallyAdded: false,
				Date:          date,
			}
			transfers = append(transfers, transfer)
		}
	}
	return transfers, nil
}

func (api API) getAccessToken() (*string, *wardrobe.TDAccount, error) {
	tdAccount, err := wardrobe.FetchTDAccount(api.AccountId)
	if err != nil {
		return nil, nil, err
	}
	auth, err := FetchAccessToken(tdAccount.RefreshToken)
	if err != nil {
		return nil, nil, err
	}
	err = wardrobe.UpdateRefreshToken(api.AccountId, auth.RefreshToken)
	if err != nil {
		return nil, nil, err
	}
	return &auth.AccessToken, tdAccount, nil
}
