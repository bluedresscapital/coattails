package robinhood

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/shopspring/decimal"

	"github.com/bluedresscapital/coattails/pkg/util"
)

const (
	ClientId               = "c82SH0WZOsabOXGP2sxqcj34FxkvfnWRZBKlBjFS"
	OrdersUrl              = "https://api.robinhood.com/orders/"
	TransfersUrl           = "https://api.robinhood.com/ach/transfers/"
	ReceivedTransfersUrl   = "https://api.robinhood.com/ach/received/transfers/"
	SettledTransactionsUrl = "https://minerva.robinhood.com/history/settled_transactions/"
)

type RHOrdersResponse struct {
	Next    string            `json:"next"`
	Results []RHOrdersResults `json:"results"`
}

type RHOrdersResults struct {
	Id                string               `json"id"`
	Instrument        string               `json:"instrument"`
	Side              string               `json:"side"`
	LastTransactionAt time.Time            `json:"last_transaction_at"`
	Executions        []RHOrdersExecutions `json:"executions"`
}

type RHOrdersExecutions struct {
	Price     decimal.Decimal `json:"price"`
	Quantity  decimal.Decimal `json:"quantity"`
	Timestamp time.Time       `json:"timestamp"`
}

type RHAuthResponse struct {
	BearerTok  string `json:"access_token"`
	RefreshTok string `json:"refresh_token"`
}

// Fetches bearer token using refresh token, HOWEVER this immediately invalidates the current
// refresh token. MUST use the new one returned by this call!
func FetchBearerToken(refreshTok string) (*RHAuthResponse, error) {
	reqBody, err := json.Marshal(map[string]string{
		"refresh_token": refreshTok,
		"client_id":     ClientId,
		"expires_in":    "86400",
		"grant_type":    "refresh_token",
		"scope":         "internal",
	})
	if err != nil {
		return nil, err
	}
	resp, err := http.Post("https://api.robinhood.com/oauth2/token/", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var res RHAuthResponse
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	if res.BearerTok == "" || res.RefreshTok == "" {
		return nil, errors.New("invalid auth grant")
	}
	return &res, nil
}

func ScrapeOrders(bearerTok string) {
	res := make([]RHOrdersResults, 0)
	url := OrdersUrl
	for {
		log.Printf("Making request to %s", url)
		resp, err := util.MakeGetRequest(bearerTok, url)
		if err != nil {
			log.Printf("Error: %v", err)
		}
		body, _ := ioutil.ReadAll(resp.Body)
		var orders RHOrdersResponse
		err = json.Unmarshal(body, &orders)
		if err != nil {
			log.Printf("Error unmarshalling: %v", err)
		}
		for _, r := range orders.Results {
			res = append(res, r)
		}
		if orders.Next == "" {
			break
		}
		url = orders.Next
	}
	for _, o := range res {
		if len(o.Executions) != 1 {
			log.Printf("This has multiple (or none?) executions: %v", o)
		}
	}
}

func ScrapeTransfers(bearerTok string) {

}
