package stockings

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"net/http"
	"time"
)

/*
   Could look to make this an interface as we move forward with different stock apis
   ie make different structs for td, iex, rh apis and implement each of these methods for the
   different apis via StockAPI interface.

*/
type IexApi struct{}

type iexStock struct {
	Symbol        string          `json:"symbol"`
	Name          string          `json:"companyName"`
	LatestPrice   decimal.Decimal `json:"latestPrice"`
	Change        float32         `json:"change"`
	ChangePercent float32         `json:"changePercent`
}

type iexHistoricalStock struct {
	Date  string          `json:"date"`
	Price decimal.Decimal `json:"close"`
}

type iexHistoricalStocks []iexHistoricalStock

const (
	iexCurrentPriceUrl        = "https://cloud.iexapis.com/stable/stock/%s/quote?token=%s"
	iexHistoricalDateRangeUrl = "https://cloud.iexapis.com/stable/stock/%s/chart/%s?token=%s"
	iexHistoricalDateUrl      = "https://cloud.iexapis.com/stable/stock/%s/chart/date/%s?chartByDay=true&token=%s"
	//https://cloud.iexapis.com/stable/stock/MELI/chart/date/20200102?chartByDay=true&token=pk_ec21611ca5f5492e9397b4a1879ff114
	DateLayout    = "20060102"
	IexDateLayout = "2006-01-02"
)

//example for ralles, he should refactor this to better handle error checking etc
//since this is a large struct, should we perhaps return *IexStock?
func (iex IexApi) GetCurrentPrice(ticker string) (*Stock, error) {
	url := fmt.Sprintf(iexCurrentPriceUrl, ticker, getKey())
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("http resp not 200")
	}
	quote := new(iexStock)
	err = json.NewDecoder(resp.Body).Decode(quote)
	if err != nil {
		return nil, err
	}
	var res = Stock{quote.Symbol, quote.Name, quote.LatestPrice, quote.Change, quote.ChangePercent}
	return &res, nil
}

//function that returns HistoricalStock at a certain date
func (iex IexApi) GetHistoricalPrice(ticker string, date time.Time) (*HistoricalStock, error) {
	parsedDate := date.Format(DateLayout)
	url := fmt.Sprintf(iexHistoricalDateUrl, ticker, parsedDate, getKey())
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("http resp not 200")
	}
	historical := new(iexHistoricalStocks)
	err = json.NewDecoder(resp.Body).Decode(historical)
	if err != nil {
		return nil, err
	}
	if len(*historical) != 1 {
		return nil, errors.New("did not return singular value after unmarshall")
	}
	formattedDate, _ := time.Parse(IexDateLayout, (*historical)[0].Date)
	return &(HistoricalStock{formattedDate, (*historical)[0].Price}), nil
}

//function that returns a pointer to a slice of IexHistoricalStock's for a date range
//do we need to return *[]*IexHistoricalStock?
func (iex IexApi) GetHistoricalRange(ticker string, start time.Time, end time.Time) (*HistoricalStocks, error) {
	parsedStart := start.Format(DateLayout)
	parsedEnd := end.Format(DateLayout)
	rangeQuery, err := getRange(parsedStart)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf(iexHistoricalDateRangeUrl, ticker, *rangeQuery, getKey())
	println(url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("http resp not 200")
	}
	historical := new(iexHistoricalStocks)
	err = json.NewDecoder(resp.Body).Decode(historical)
	if err != nil {
		return nil, err
	}

	startIndex, err := startOfHistoricalRange(historical, parsedStart)
	if err != nil {
		return nil, err
	}

	endIndex, err := endOfHistoricalRange(historical, parsedEnd)
	if err != nil {
		return nil, err
	}

	if *startIndex > *endIndex {
		return nil, errors.New("Invalid range")
	}

	*historical = (*historical)[*startIndex:*endIndex]

	return convertToHistoricalRange(historical), nil
}

//taken from https://github.com/addisonlynch/iexfinance/blob/master/iexfinance/stocks/historical.py
//some leap year stuff is fucked up for them so will have to rewrite this later probably
func getRange(date string) (*string, error) {
	startDate, _ := time.Parse(DateLayout, date)
	endDate, _ := time.Parse(DateLayout, time.Now().Format(DateLayout))
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

func startOfHistoricalRange(historicalPrices *iexHistoricalStocks, startDate string) (*int, error) {
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

func endOfHistoricalRange(historicalPrices *iexHistoricalStocks, endDate string) (*int, error) {
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
	startDate, _ := time.Parse(IexDateLayout, date)
	return startDate.Format(DateLayout)
}

func convertToHistoricalRange(stocks *iexHistoricalStocks) *HistoricalStocks {
	ret := new(HistoricalStocks)
	for i := 0; i < len(*stocks); i++ {
		formattedDate, _ := time.Parse(IexDateLayout, (*stocks)[i].Date)
		*ret = append(*ret, HistoricalStock{formattedDate, (*stocks)[i].Price})
	}
	return ret
}
