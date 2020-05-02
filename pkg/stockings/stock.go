package stockings

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
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
	Date  string  `json:"date"`
	Price float32 `json:"close"`
}

type HistoricalStocks []HistoricalStock

const (
	iexCurrentPriceUrl   = "https://cloud.iexapis.com/stable/stock/%s/quote?token=%s"
	iexHistoricalDateUrl = "https://cloud.iexapis.com/stable/stock/%s/chart/%s?token=%s"
	//bonus points if you know what day this is :)
	dateLayout    = "20060102"
	iexDateLayout = "2006-01-02"
)

//example for ralles, he should refactor this to better handle error checking etc
//since this is a large struct, should we perhaps return *Stock?
func GetCurrentPrice(ticker string) Stock {
	url := fmt.Sprintf(iexCurrentPriceUrl, ticker, os.Getenv("IEX_TOKEN"))
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
	url := fmt.Sprintf(iexHistoricalDateUrl, ticker, "date/"+date+"?chartByDay=true", os.Getenv("IEX_TOKEN"))
	resp, err := http.Get(url)
	var historical HistoricalStocks
	err = json.NewDecoder(resp.Body).Decode(&historical)
	if err != nil {
		log.Fatal(err)
	}
	return historical[0]
}

//function that returns a pointer to a slice of HistoricalStock's for a date range
//do we need to return *[]*HistoricalStock?
func GetHistoricalRange(ticker string, start string, end string) *HistoricalStocks {
	rangeQuery := getRange(start)
	if rangeQuery == "Error" {
		log.Panic("range error")
	}
	url := fmt.Sprintf(iexHistoricalDateUrl, ticker, rangeQuery, os.Getenv("IEX_TOKEN"))
	resp, err := http.Get(url)
	var historical HistoricalStocks
	err = json.NewDecoder(resp.Body).Decode(&historical)
	if err != nil {
		log.Fatal(err)
	}
	parseHistoricalRange(&historical, start, end)
	return &historical
}

//taken from https://github.com/addisonlynch/iexfinance/blob/master/iexfinance/stocks/historical.py
//some leap year stuff is fucked up for them so will have to rewrite this later probably
func getRange(date string) string {
	startDate, _ := time.Parse(dateLayout, date)
	endDate, _ := time.Parse(dateLayout, time.Now().Format(dateLayout))
	diff := endDate.Sub(startDate)
	days := int(diff.Hours() / 24)
	if 0 <= days && days < 6 {
		return "5d"
	} else if 6 <= days && days < 28 {
		return "1m"
	} else if 28 <= days && days < 84 {
		return "3m"
	} else if 84 <= days && days < 168 {
		return "6m"
	} else if 168 <= days && days < 365 {
		return "1y"
	} else if 365 <= days && days < 730 {
		return "2y"
	} else if 730 <= days && days < 1826 {
		return "5y"
	} else if 1826 <= days && days < 5478 {
		return "max"
	}
	return "Error"
}

//modifies the parameter
func parseHistoricalRange(historicalPrices *HistoricalStocks, startDate string, endDate string) {
	startIdx := -1
	it := 0
	for startIdx == -1 {
		translated := translateIexDate((*historicalPrices)[it].Date)
		if translated >= startDate {
			startIdx = it
		}
		it++
	}
	endIdx := -1
	it = len((*historicalPrices)) - 1
	for endIdx == -1 {
		translated := translateIexDate((*historicalPrices)[it].Date)
		if translated <= endDate {
			endIdx = it + 1
		}
		it--
	}
	*historicalPrices = (*historicalPrices)[startIdx:endIdx]
}

func translateIexDate(date string) string {
	startDate, _ := time.Parse(iexDateLayout, date)
	return startDate.Format(dateLayout)
}
