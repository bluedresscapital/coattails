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
type Stock struct {
	Symbol        string  `json:"symbol"`
	Name          string  `json:"companyName"`
	LatestPrice   float32 `json:"latestPrice"`
	Change        float32 `json:"change"`
	ChangePercent float32 `json:"changePercent`
}

//test1
//test2
type HistoricalStock struct {
	Date  string  `json:"date"`
	Price float32 `json:"close"`
}

type HistoricalStocks []HistoricalStock

const (
	iexCurrentPriceUrl        = "https://cloud.iexapis.com/stable/stock/%s/quote?token=%s"
	iexHistoricalDateRangeUrl = "https://cloud.iexapis.com/stable/stock/%s/chart/%s?token=%s"
	iexHistoricalDateUrl      = "https://cloud.iexapis.com/stable/stock/%s/chart/date/%s?chartByDay=true&token=%s"
	//https://cloud.iexapis.com/stable/stock/MELI/chart/date/20200102?chartByDay=true&token=pk_ec21611ca5f5492e9397b4a1879ff114
	dateLayout    = "20060102"
	iexDateLayout = "2006-01-02"
)

//example for ralles, he should refactor this to better handle error checking etc
//since this is a large struct, should we perhaps return *Stock?
func GetCurrentPrice(ticker string) (Stock, error) {
	url := fmt.Sprintf(iexCurrentPriceUrl, ticker, getKey())
	resp, err := http.Get(url)
	if err != nil {
		return Stock{}, err
	}
	if resp.StatusCode != 200 {
		return Stock{}, errors.New("http resp not 200")
	}
	var quote Stock
	err = json.NewDecoder(resp.Body).Decode(&quote)
	if err != nil {
		return Stock{}, err
	}
	return quote, nil
}

//function that returns HistoricalStock at a certain date
func GetHistoricalPrice(ticker string, date string) (HistoricalStock, error) {

	url := fmt.Sprintf(iexHistoricalDateUrl, ticker, date, getKey())

	resp, err := http.Get(url)
	if resp.StatusCode != 200 {
		return HistoricalStock{}, errors.New("http resp not 200")
	}
	var historical HistoricalStocks
	err = json.NewDecoder(resp.Body).Decode(&historical)
	if err != nil {
		return HistoricalStock{}, err
	}
	if len(historical) != 1 {
		return HistoricalStock{}, errors.New("did not return singular value after unmarshall")
	}
	return historical[0], nil
}

//function that returns a pointer to a slice of HistoricalStock's for a date range
//do we need to return *[]*HistoricalStock?
func GetHistoricalRange(ticker string, start string, end string) (*HistoricalStocks, error) {
	rangeQuery, err := getRange(start)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf(iexHistoricalDateRangeUrl, ticker, rangeQuery, getKey())

	resp, err := http.Get(url)
	if resp.StatusCode != 200 {
		return nil, errors.New("http resp not 200")
	}
	var historical HistoricalStocks
	err = json.NewDecoder(resp.Body).Decode(&historical)
	if err != nil {
		return nil, err
	}

	// startIndex := indexOfHistoricalRange(historical, start)
	// endIndex := indexOfHistoricalRange(historical, end)

	// return historical[startIndex:endIndex], nil

	//this is how this would be done in c++...modify param since it's a pointer
	err = parseHistoricalRange(&historical, start, end)
	if err != nil {
		return nil, err
	}
	return &historical, nil
}

//taken from https://github.com/addisonlynch/iexfinance/blob/master/iexfinance/stocks/historical.py
//some leap year stuff is fucked up for them so will have to rewrite this later probably
func getRange(date string) (string, error) {
	startDate, _ := time.Parse(dateLayout, date)
	endDate, _ := time.Parse(dateLayout, time.Now().Format(dateLayout))
	diff := endDate.Sub(startDate)
	days := int(diff.Hours() / 24)
	if 0 <= days && days < 6 {
		return "5d", nil
	} else if 6 <= days && days < 28 {
		return "1m", nil
	} else if 28 <= days && days < 84 {
		return "3m", nil
	} else if 84 <= days && days < 168 {
		return "6m", nil
	} else if 168 <= days && days < 365 {
		return "1y", nil
	} else if 365 <= days && days < 730 {
		return "2y", nil
	} else if 730 <= days && days < 1826 {
		return "5y", nil
	} else if 1826 <= days && days < 5478 {
		return "max", nil
	}
	return "", errors.New("invalid range")
}

// func indexOfHistoricalRange()

//modifies the parameter
func parseHistoricalRange(historicalPrices *HistoricalStocks, startDate string, endDate string) error {
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
	if startIdx > endIdx {
		return errors.New("invalid date range")
	}
	*historicalPrices = (*historicalPrices)[startIdx:endIdx]
	return nil
}

func translateIexDate(date string) string {
	startDate, _ := time.Parse(iexDateLayout, date)
	return startDate.Format(dateLayout)
}
