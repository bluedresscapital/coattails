package tda

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"net/http"
	url2 "net/url"
)

//func FetchRefreshToken()

const ClientId = "GBCZDGRJAOIJHF0IETOA76NFAKZ0OGQX"

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
func FetchAccessToken(refreshToken string) (*AuthResponse, error) {
	escapedToken := url2.QueryEscape(refreshToken)
	url := "https://api.tdameritrade.com/v1/oauth2/token"
	data := fmt.Sprintf(
		"grant_type=refresh_token&refresh_token=%s&access_type=offline&code=&client_id=%s%%40AMER.OAUTHAP&redirect_uri=http%%3A%%2F%%2Flocalhost",
		escapedToken,
		ClientId)
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
	TransactionId   int               `json:"transactionId"`
	TransactionDate string            `json:"transactionDate"`
	TransactionItem TDTransactionItem `json:"transactionItem"`
	Type            string            `json:"type"`
	Description     string            `json:"description"`
	NetAmount       decimal.Decimal   `json:"netAmount"`
}

type TDTransactionItem struct {
	Amount      decimal.Decimal             `json:"amount"`
	Price       decimal.Decimal             `json:"price"`
	Instrument  TDTransactionItemInstrument `json:"instrument"`
	Instruction string                      `json:"instruction"`
}

type TDTransactionItemInstrument struct {
	Symbol string `json:"symbol"`
}

type TDTransfer struct {
	TransactionId  string          `json:"transactionId"`
	NetAmount      decimal.Decimal `json:"netAmount"`
	SettlementDate string          `json:"settlementDate"`
}

type TDTransfers []TDTransfer

func ScrapeTransactions(authTok string, accountId string) (TDTransactions, error) {
	url := fmt.Sprintf("https://api.tdameritrade.com/v1/accounts/%s/transactions", accountId)
	resp, err := makeGetRequest(authTok, url)
	if err != nil {
		return nil, err
	}
	body, _ := ioutil.ReadAll(resp.Body)
	//log.Printf("Scraped:\n%s", string(body))
	var orders TDTransactions
	err = json.Unmarshal(body, &orders)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func makeGetRequest(authTok string, url string) (*http.Response, error) {
	var bearer = fmt.Sprintf("Bearer %s", authTok)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", bearer)
	client := &http.Client{}
	return client.Do(req)
}
