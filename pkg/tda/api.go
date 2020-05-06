package tda

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"log"
	"net/http"
	url2 "net/url"
)

//func FetchRefreshToken()

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Fetches TD Access Token.
// This is kind of annoying, but there's a pretty delicate token management that we need to be wary of.
// Seems like in order to make any TD authed request, we need an access token, which is ephemeral and lasts
// only once per request
// In order to get an access token, we need a refresh token - seems like this token at least lasts longer than once
// per request, but still only lasts a finite # of requests (like < 10). So for each request that requires an access
// token, we would need to:
// 1. Use our current refresh token to fetch a new access token
// 2. When we get the new access token, replace the old refresh token with the new one returned (lol)
// 3. Use access token.
// Note - if we ever mishandle the refresh token (i.e. delete or lose it), we would need to reauth TD
// Ideally if done correctly, the client should never realize that we're constantly swapping these refresh tokens.
func FetchAccessToken(refreshToken string, clientId string) (*AuthResponse, error) {
	escapedToken := url2.QueryEscape(refreshToken)
	url := "https://api.tdameritrade.com/v1/oauth2/token"
	data := fmt.Sprintf(
		"grant_type=refresh_token&refresh_token=%s&access_type=offline&code=&client_id=%s%%40AMER.OAUTHAP&redirect_uri=http%%3A%%2F%%2Flocalhost",
		escapedToken,
		clientId)
	resp, err := http.Post(url, "application/x-www-form-urlencoded", bytes.NewBuffer([]byte(data)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	log.Printf("resp body: %s", string(body))
	var auth AuthResponse
	err = json.Unmarshal(body, &auth)
	if err != nil {
		return nil, err
	}
	return &auth, nil
}

// Given auth code,
func FetchRefreshTokenUsingAuthCode(code string, clientId string) (*AuthResponse, error) {
	encodedCode := url2.QueryEscape(code)
	data := fmt.Sprintf(
		"grant_type=authorization_code&refresh_token=&access_type=offline&code=%s&client_id=%s%%40AMER.OAUTHAP&redirect_uri=http%%3A%%2F%%2Flocalhost",
		encodedCode,
		clientId)
	url := "https://api.tdameritrade.com/v1/oauth2/token"
	resp, err := http.Post(url, "application/x-www-form-urlencoded", bytes.NewBuffer([]byte(data)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var auth AuthResponse
	err = json.Unmarshal(body, &auth)
	if err != nil {
		return nil, err
	}
	if auth.AccessToken == "" || auth.RefreshToken == "" {
		return nil, fmt.Errorf("invalid auth response: %s", string(body))
	}
	return &auth, nil
}

type TDTransactions []TDTransactionResponse

type TDTransactionResponse struct {
	OrderDate       string            `json:"orderDate"`
	TransactionId   string            `json:"transactionId"`
	TransactionDate string            `json:"transactionDate"`
	TransactionItem TDTransactionItem `json:"transactionItem"`
	Type            string            `json:"type"`
}

type TDTransactionItem struct {
	Amount      decimal.Decimal             `json:"amount"`
	Instrument  TDTransactionItemInstrument `json:"instrument"`
	Instruction string                      `json:"instruction"`
}

type TDTransactionItemInstrument struct {
	Symbol string `json:"symbol"`
}

func ScrapeOrders(authTok string, accountId string) ([]wardrobe.Order, error) {
	var bearer = fmt.Sprintf("Bearer %s", authTok)
	url := fmt.Sprintf("https://api.tdameritrade.com/v1/accounts/%s/transactions", accountId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", bearer)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	body, _ := ioutil.ReadAll(resp.Body)
	var trans TDTransactions
	_ = json.Unmarshal(body, &trans)
	log.Printf("res: %v", trans)

	//for i, t := range trans {
	//	b
	//}

	return nil, nil
}
