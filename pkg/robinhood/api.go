package robinhood

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/bluedresscapital/coattails/pkg/wardrobe"

	"github.com/shopspring/decimal"

	"github.com/bluedresscapital/coattails/pkg/util"
)

const (
	ClientId               = "c82SH0WZOsabOXGP2sxqcj34FxkvfnWRZBKlBjFS"
	AuthUrl                = "https://api.robinhood.com/oauth2/token/"
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

func Login(username string, password string, deviceTok string) (*RHAuthResponse, error) {
	reqBody, err := json.Marshal(map[string]string{
		"username":     username,
		"password":     password,
		"device_token": deviceTok,
		"client_id":    ClientId,
		"expires_in":   "86400",
		"grant_type":   "password",
		"scope":        "internal",
	})
	if err != nil {
		return nil, err
	}
	return fetchRHAuthResponse(reqBody)
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
	return fetchRHAuthResponse(reqBody)
}

func fetchRHAuthResponse(reqBody []byte) (*RHAuthResponse, error) {
	resp, err := http.Post(AuthUrl, "application/json", bytes.NewBuffer(reqBody))
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

func ScrapeOrders(bearerTok string) ([]RHOrdersResults, error) {
	res := make([]RHOrdersResults, 0)
	url := OrdersUrl
	for {
		resp, err := util.MakeGetRequest(bearerTok, url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		var orders RHOrdersResponse
		err = json.Unmarshal(body, &orders)
		if err != nil {
			return nil, err
		}
		for _, r := range orders.Results {
			res = append(res, r)
		}
		if orders.Next == "" {
			break
		}
		url = orders.Next
	}
	return res, nil
}

type InstrumentResponse struct {
	Symbol string `json:"symbol"`
}

func FetchStockFromInstrumentId(instrument string) (*string, error) {
	stock, err := wardrobe.GetStockFromInstrumentId(instrument)
	if err == nil {
		return stock, nil
	}
	resp, err := http.Get(instrument)
	if err != nil {
		return nil, err
	}
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	var res InstrumentResponse
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}
	wardrobe.SetStockFromInstrument(instrument, res.Symbol)
	return &res.Symbol, nil
}

type RHTransfersResults struct {
	Id        string          `json:"id"`
	IsDeposit bool            `json:"is_deposit"`
	Amount    decimal.Decimal `json:"amount"`
	Date      time.Time       `json:"date"`
}

func ScrapeTransfers(bearerTok string) ([]RHTransfersResults, error) {
	res := make([]RHTransfersResults, 0)
	// Normal bank transfers, where you directly withdraw from
	bankTransfers, err := scrapeBankTransfers(bearerTok)
	if err != nil {
		return nil, err
	}
	for _, t := range bankTransfers {
		if t.State == "completed" {
			res = append(res, RHTransfersResults{
				Id:        t.Id,
				IsDeposit: t.Direction == "deposit",
				Amount:    t.Amount,
				Date:      t.CreatedAt,
			})
		}
	}
	// Transfers related to RH Checking Account (i.e. Direct deposit, Venmo)
	receivedTransfers, err := scrapeReceviedTransfers(bearerTok)
	if err != nil {
		return nil, err
	}
	for _, t := range receivedTransfers {
		if t.State == "completed" {
			res = append(res, RHTransfersResults{
				Id:        t.Id,
				IsDeposit: t.Direction == "credit",
				Amount:    t.Amount.Amount,
				Date:      t.CreatedAt,
			})
		}
	}
	// Transfers related to RH debit card usage (Spending, ATM withdrawals)
	settledTransactions, err := scrapeSettledTransactions(bearerTok)
	if err != nil {
		return nil, err
	}
	for _, t := range settledTransactions {
		res = append(res, RHTransfersResults{
			Id:        t.Id,
			IsDeposit: t.Direction == "credit",
			Amount:    t.Amount.Amount,
			Date:      t.InitiatedAt,
		})
	}
	return res, nil
}

type RHBankTransfersResponse struct {
	Next    string                   `json:"next"`
	Results []RHBankTransfersResults `json:"results"`
}

type RHBankTransfersResults struct {
	Id        string          `json:"id"`
	Direction string          `json:"direction"`
	Amount    decimal.Decimal `json:"amount"`
	State     string          `json:"state"`
	CreatedAt time.Time       `json:"created_at"`
}

func scrapeBankTransfers(bearerTok string) ([]RHBankTransfersResults, error) {
	log.Print("bank transfers")
	res := make([]RHBankTransfersResults, 0)
	url := TransfersUrl
	for {
		resp, err := util.MakeGetRequest(bearerTok, url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		log.Print(string(body))
		var transfers RHBankTransfersResponse
		err = json.Unmarshal(body, &transfers)
		if err != nil {
			return nil, err
		}
		for _, r := range transfers.Results {
			res = append(res, r)
		}
		if transfers.Next == "" {
			break
		}
		url = transfers.Next
	}
	return res, nil
}

type RHReceivedTransfersResponse struct {
	Next    string                       `json:"next"`
	Results []RHReceivedTransfersResults `json:"results"`
}

type RHReceivedTransfersResults struct {
	Id        string                    `json:"id"`
	Amount    RHReceivedTransfersAmount `json:"amount"`
	Direction string                    `json:"direction"`
	State     string                    `json:"state"`
	CreatedAt time.Time                 `json:"created_at"`
}

type RHReceivedTransfersAmount struct {
	Amount decimal.Decimal `json:"amount"`
}

func scrapeReceviedTransfers(bearerTok string) ([]RHReceivedTransfersResults, error) {
	log.Print("received transfers")
	res := make([]RHReceivedTransfersResults, 0)
	url := ReceivedTransfersUrl
	for {
		resp, err := util.MakeGetRequest(bearerTok, url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		log.Print(string(body))
		var transfers RHReceivedTransfersResponse
		err = json.Unmarshal(body, &transfers)
		if err != nil {
			return nil, err
		}
		for _, r := range transfers.Results {
			res = append(res, r)
		}
		if transfers.Next == "" {
			break
		}
		url = transfers.Next
	}
	return res, nil
}

type RHSettledTransactionsResponse struct {
	Next    string                         `json:"next"`
	Results []RHSettledTransactionsResults `json:"results"`
}

type RHSettledTransactionsResults struct {
	Id          string                      `json:"id"`
	Amount      RHSettledTransactionsAmount `json:"amount"`
	Direction   string                      `json:"direction"`
	InitiatedAt time.Time                   `json:"initiated_at"`
}

type RHSettledTransactionsAmount struct {
	Amount decimal.Decimal `json:"amount"`
}

func scrapeSettledTransactions(bearerTok string) ([]RHSettledTransactionsResults, error) {
	res := make([]RHSettledTransactionsResults, 0)
	url := SettledTransactionsUrl
	for {
		resp, err := util.MakeGetRequest(bearerTok, url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		//log.Print(string(body))
		var transfers RHSettledTransactionsResponse
		err = json.Unmarshal(body, &transfers)
		if err != nil {
			return nil, err
		}
		for _, r := range transfers.Results {
			res = append(res, r)
		}
		if transfers.Next == "" {
			break
		}
		url = transfers.Next
	}
	return res, nil
}
