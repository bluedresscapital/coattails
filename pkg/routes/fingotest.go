package routes

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bluedresscapital/coattails/pkg/stockings"
)

func checkPiquette(w http.ResponseWriter, r *http.Request) {
	api := stockings.FingoPack{}
	testQuote, err := api.GetCurrentPrice("MELI")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Fprintln(w, testQuote.LatestPrice)

	// this creates a time.time format i assume
	date, _ := time.Parse(stockings.DateLayout, "20200102")
	testHistoric, err := api.GetHistoricalPrice("meli", date)

	if err != nil {
		fmt.Println(err)
	}
	fmt.Fprintln(w, testHistoric.Price)

	start, _ := time.Parse(stockings.DateLayout, "20200101")
	end, _ := time.Parse(stockings.DateLayout, "20200107")
	testHistoricRange, err := api.GetHistoricalRange("meli", start, end)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Fprintln(w, testHistoricRange)

}
