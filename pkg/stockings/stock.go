package stockings

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type Stock struct {
	Symbol        string  `json:"symbol"`
	Name          string  `json:"companyName"`
	LatestPrice   float32 `json:"latestPrice"`
	Change        float32 `json:"change"`
	ChangePercent float32 `json:"changePercent`
}

var iexUrl = "https://cloud.iexapis.com/stable/stock/%s/quote?token=%s"

func GetStockQuote(ticker string) Stock {
	url := fmt.Sprintf(iexUrl, ticker, os.Getenv("IEX_TOKEN"))
	resp, err := http.Get(url)
	if err != nil {
		log.Panic("oh no!")
	}
	var quote Stock
	err = json.NewDecoder(resp.Body).Decode(&quote)
	if err != nil {
		log.Fatal(err)
	}
	return quote
}
