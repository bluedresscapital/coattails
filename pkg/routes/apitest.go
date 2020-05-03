package routes

import (
	"encoding/json"
	"fmt"
	"github.com/bluedresscapital/coattails/pkg/stockings"
	"net/http"
)

func apiCheck(w http.ResponseWriter, r *http.Request) {
	aObj, err := stockings.GetCurrentPrice("MELI")
	a, _ := json.Marshal(aObj)
	fmt.Fprintln(w, string(a))
	if err != nil {
		fmt.Println(err)
	}

	bObj, err := stockings.GetHistoricalRange("MELI", "20200101", "20200105")
	b, _ := json.Marshal(bObj)
	fmt.Fprintln(w, string(b))
	if err != nil {
		fmt.Println(err)
	}

	cObj, err := stockings.GetHistoricalPrice("MELI", "20200102")
	c, _ := json.Marshal(cObj)
	fmt.Fprintln(w, string(c))
	if err != nil {
		fmt.Println(err)
	}
}
