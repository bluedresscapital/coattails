package stockings

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

/*
	Could look to make this an interface as we move forward with different stock apis
	ie make different structs for td, iex, rh apis and implement each of these methods for the
	different apis via StockAPI interface.

*/
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

const iexUrl = "https://cloud.iexapis.com/stable/stock/%s/quote?token=%s"

//example for ralles, he should refactor this to better handle error checking etc
//since this is a large struct, should we perhaps return *Stock?
func GetCurrentPrice(ticker string) Stock {
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
func GetHistoricalPrice(ticker string, date string) HistoricalStock {
	return HistoricalStock{}
}

//function that returns a pointer to a slice of HistoricalStock's for a date range
//do we need to return *[]*HistoricalStock?
func GetHistoricalRange(ticker string, start string, end string) *[]HistoricalStock {
	return &[]HistoricalStock{}
}
