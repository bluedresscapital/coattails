package routes

import (
	"encoding/json"
	"fmt"
	"github.com/bluedresscapital/coattails/pkg/stockings"
	"net/http"
	"time"
)

func apiCheck(w http.ResponseWriter, r *http.Request) {
	api := stockings.IexApi{}
	aObj, err := api.GetCurrentPrice("MELI")
	a, _ := json.Marshal(aObj)
	fmt.Fprintln(w, string(a))
	if err != nil {
		fmt.Println(err)
	}
	start, _ := time.Parse(stockings.DateLayout, "20200101")
	end, _ := time.Parse(stockings.DateLayout, "20200105")
	bObj, err := api.GetHistoricalRange("MELI", start, end)
	b, _ := json.Marshal(bObj)
	fmt.Fprintln(w, string(b))
	if err != nil {
		fmt.Println(err)
	}

	date, _ := time.Parse(stockings.DateLayout, "20200102")
	cObj, err := api.GetHistoricalPrice("MELI", date)
	c, _ := json.Marshal(cObj)
	fmt.Fprintln(w, string(c))
	if err != nil {
		fmt.Println(err)
	}
}
