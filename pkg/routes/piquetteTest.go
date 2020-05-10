package routes

// import (
// 	//"encoding/json"
// 	"fmt"
// 	//"github.com/bluedresscapital/coattails/pkg/stockings"
// 	"net/http"
// 	//"time"
// )

// func middleOutTest(w http.ResponseWriter, r *http.Request) {
// 	fmt.Fprintf(w, "ping!")
// }

// for now I will try to use this to test my go api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bluedresscapital/coattails/pkg/stockings"
)

func checkPiquette(w http.ResponseWriter, r *http.Request) {
	api := stockings.PiquetteAPI{}
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
	end, _ := time.Parse(stockings.DateLayout, "20200105")
	testHistoricRange, err := api.GetHistoricalRange("meli", start, end)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Fprintln(w, testHistoricRange)

}
