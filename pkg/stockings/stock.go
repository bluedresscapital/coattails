package stockings

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

/*
	Could look to make this an interface as we move forward with different stock apis
	ie make different structs for td, iex, rh apis and implement each of these methods for the
	different apis via StockAPI interface.

*/
type IexStock struct {
	Symbol        string  `json:"symbol"`
	Name          string  `json:"companyName"`
	LatestPrice   float32 `json:"latestPrice"`
	Change        float32 `json:"change"`
	ChangePercent float32 `json:"changePercent`
}

type IexHistoricalStock struct {
	Date  string  `json:"date"`
	Price float32 `json:"close"`
}

type IexHistoricalStocks []IexHistoricalStock

const (
	iexCurrentPriceUrl        = "https://cloud.iexapis.com/stable/stock/%s/quote?token=%s"
	iexHistoricalDateRangeUrl = "https://cloud.iexapis.com/stable/stock/%s/chart/%s?token=%s"
	iexHistoricalDateUrl      = "https://cloud.iexapis.com/stable/stock/%s/chart/date/%s?chartByDay=true&token=%s"
	//https://cloud.iexapis.com/stable/stock/MELI/chart/date/20200102?chartByDay=true&token=pk_ec21611ca5f5492e9397b4a1879ff114
	dateLayout    = "20060102"
	iexDateLayout = "2006-01-02"
)

//example for ralles, he should refactor this to better handle error checking etc
//since this is a large struct, should we perhaps return *IexStock?
func GetCurrentPrice(ticker string) (*IexStock, error) {
	url := fmt.Sprintf(iexCurrentPriceUrl, ticker, getKey())
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("http resp not 200")
	}
	quote := new(IexStock)
	err = json.NewDecoder(resp.Body).Decode(quote)
	if err != nil {
		return nil, err
	}
	return quote, nil
}

//function that returns HistoricalStock at a certain date
func GetHistoricalPrice(ticker string, date string) (*IexHistoricalStock, error) {

	url := fmt.Sprintf(iexHistoricalDateUrl, ticker, date, getKey())

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("http resp not 200")
	}
	historical := new(IexHistoricalStocks)
	err = json.NewDecoder(resp.Body).Decode(historical)
	if err != nil {
		return nil, err
	}
	if len(*historical) != 1 {
		return nil, errors.New("did not return singular value after unmarshall")
	}

	return &((*historical)[0]), nil
}

//function that returns a pointer to a slice of IexHistoricalStock's for a date range
//do we need to return *[]*IexHistoricalStock?
func GetHistoricalRange(ticker string, start string, end string) (*IexHistoricalStocks, error) {
	rangeQuery, err := getRange(start)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf(iexHistoricalDateRangeUrl, ticker, *rangeQuery, getKey())

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("http resp not 200")
	}
	historical := new(IexHistoricalStocks)
	err = json.NewDecoder(resp.Body).Decode(historical)
	if err != nil {
		return nil, err
	}

	startIndex, err := startOfHistoricalRange(historical, start)
	if err != nil {
		return nil, err
	}

	endIndex, err := endOfHistoricalRange(historical, end)
	if err != nil {
		return nil, err
	}

	if *startIndex > *endIndex {
		return nil, errors.New("Invalid range")
	}

	*historical = (*historical)[*startIndex:*endIndex]

	return historical, nil
}

//taken from https://github.com/addisonlynch/iexfinance/blob/master/iexfinance/stocks/historical.py
//some leap year stuff is fucked up for them so will have to rewrite this later probably
func getRange(date string) (*string, error) {
	startDate, _ := time.Parse(dateLayout, date)
	endDate, _ := time.Parse(dateLayout, time.Now().Format(dateLayout))
	diff := endDate.Sub(startDate)
	days := int(diff.Hours() / 24)
	dayRange := ""

	if 0 <= days && days < 6 {
		dayRange = "5d"
	} else if 6 <= days && days < 28 {
		dayRange = "1m"
	} else if 28 <= days && days < 84 {
		dayRange = "3m"
	} else if 84 <= days && days < 168 {
		dayRange = "6m"
	} else if 168 <= days && days < 365 {
		dayRange = "1y"
	} else if 365 <= days && days < 730 {
		dayRange = "2y"
	} else if 730 <= days && days < 1826 {
		dayRange = "5y"
	} else if 1826 <= days && days < 5478 {
		dayRange = "max"
	}

	if dayRange == "" {
		return nil, errors.New("invalid range")
	}

	return &dayRange, nil

}

func startOfHistoricalRange(historicalPrices *IexHistoricalStocks, startDate string) (*int, error) {
	startIdx := -1
	it := 0
	for startIdx == -1 && it < len(*historicalPrices) {
		translated := translateIexDate((*historicalPrices)[it].Date)
		if translated >= startDate {
			return &it, nil
		}
		it++
	}
	return nil, errors.New("Couldn't find a valid start date")
}

func endOfHistoricalRange(historicalPrices *IexHistoricalStocks, endDate string) (*int, error) {
	endIdx := -1
	it := len((*historicalPrices)) - 1
	for endIdx == -1 && it >= 0 {
		translated := translateIexDate((*historicalPrices)[it].Date)
		if translated <= endDate {
			it += 1
			return &it, nil
		}
		it--
	}

	return nil, errors.New("Couldn't find a valid end date")
}

func translateIexDate(date string) string {
	startDate, _ := time.Parse(iexDateLayout, date)
	return startDate.Format(dateLayout)
}
