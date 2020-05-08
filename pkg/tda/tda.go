package tda

import (
	"github.com/bluedresscapital/coattails/pkg/poncho"
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"log"
	"strconv"
	"time"
)

type API struct {
	AccountId int
}

var _ poncho.OrderAPI = (*API)(nil)

func (api API) GetOrders() ([]wardrobe.Order, error) {
	accessTok, tdAccount, err := api.getAccessToken()
	if err != nil {
		return nil, err
	}
	log.Printf("Using this accessToken: %s", *accessTok)
	trans, err := ScrapeOrders(*accessTok, tdAccount.AccountNum)
	if err != nil {
		return nil, err
	}
	port, err := wardrobe.FetchPortfolioByTDAccountId(api.AccountId)
	if err != nil {
		return nil, err
	}
	var orders []wardrobe.Order
	for _, t := range trans {
		if t.Type != "TRADE" && t.Type != "RECEIVE_AND_DELIVER" && t.Type != "TRANSFER OF SECURITY OR OPTION IN" {
			continue
		}
		date, err := time.Parse("2006-01-02T15:04:05+0000", t.TransactionDate)
		if err != nil {
			return nil, err
		}
		log.Printf("date: %s, instruction: %s", t.TransactionDate, t.TransactionItem.Instruction)
		order := wardrobe.Order{
			Uid:           strconv.Itoa(t.TransactionId),
			PortId:        port.Id,
			Stock:         t.TransactionItem.Instrument.Symbol,
			Quantity:      t.TransactionItem.Amount,
			Value:         t.TransactionItem.Price,
			IsBuy:         t.TransactionItem.Instruction == "BUY",
			ManuallyAdded: false,
			Date:          date,
		}
		orders = append(orders, order)
	}
	return orders, nil
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
