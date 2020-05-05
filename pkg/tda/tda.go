package tda

import (
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"log"
)

func GetOrders(tdAccountId int) ([]wardrobe.Order, error) {
	accessTok, tdAccount, err := getAccessToken(tdAccountId)
	if err != nil {
		return err
	}
	log.Printf("Using this accessToken: %s", *accessTok)
	return ScrapeOrders(*accessTok, tdAccount.AccountNum)
}

func getAccessToken(tdAccountId int) (*string, *wardrobe.TDAccount, error) {
	tdAccount, err := wardrobe.FetchTDAccount(tdAccountId)
	if err != nil {
		return nil, nil, err
	}
	auth, err := FetchAccessToken(tdAccount.RefreshToken, tdAccount.ClientId)
	if err != nil {
		return nil, nil, err
	}
	err = wardrobe.UpdateRefreshToken(tdAccountId, auth.RefreshToken)
	if err != nil {
		return nil, nil, err
	}
	return &auth.AccessToken, tdAccount, nil
}
