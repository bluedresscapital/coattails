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

type HistoricalStock struct {
	Date  string
	Price float32
}

var iexUrl = "https://cloud.iexapis.com/stable/stock/%s/quote?token=%s"

//example for ralles, he should refactor this to better handle error checking etc
func getCurrentPrice(ticker string) Stock {
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

//function that returns HistoricalStock at a certain date
func getHistoricalPrice(ticker string, date string) HistoricalStock {
	return nil
}

//function that returns a pointer to a slice of HistoricalStock's for a date range
func getHistoricalRange(ticker string, start string, end string) *[]HistoricalStock {
	return nil
}
