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

	"github.com/bluedresscapital/coattails/pkg/stockings"
)

func checkPiquette(w http.ResponseWriter, r *http.Request) {
	api := stockings.PiquetteAPI{}
	testQuote, err := api.GetCurrentPrice("MELI")
	fmt.Fprintln(w, testQuote.LatestPrice)
	if err != nil {
		fmt.Println(err)
	}
}
