package tda

import (
	"fmt"
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"log"
)

// Initializes TD Account, which entails:
// - fetching refresh token for td account
// - storing encrypted client_id and refresh_token into db, linked to user_id
func InitTDAccount(userId int, portId int, code string, clientId string) error {
	auth, err := FetchRefreshTokenUsingAuthCode(code, clientId)
	if err != nil {
		return err
	}
	if auth.RefreshToken == "" || auth.AccessToken == "" {
		return fmt.Errorf("invalid td auth credentials: %s, %s", code, clientId)
	}
	return wardrobe.UpsertTDAccount(userId, clientId, auth.RefreshToken)
}

func FakeRequest(userId int, tdAccountId int) error {
	tdAccount, err := wardrobe.FetchTDAccount(tdAccountId, userId)
	if err != nil {
		return err
	}
	auth, err := FetchAccessToken(tdAccount.RefreshToken, tdAccount.ClientId)
	if err != nil {
		return err
	}

	err = wardrobe.UpsertTDAccount(userId, tdAccount.ClientId, auth.RefreshToken)
	if err != nil {
		return err
	}

	log.Printf("We're using this access token: %s", auth.AccessToken)
	return nil
}
