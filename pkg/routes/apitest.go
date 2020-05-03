package routes

import (
	"encoding/json"
	"fmt"
	"github.com/bluedresscapital/coattails/pkg/stockings"
	"net/http"
)

func apiCheck(w http.ResponseWriter, r *http.Request) {
	aObj, _ := stockings.GetCurrentPrice("MELI")
	a, _ := json.Marshal(aObj)
	fmt.Fprintln(w, string(a))

	bObj, err := stockings.GetHistoricalRange("MELI", "20200101", "20200105")
	b, _ := json.Marshal(bObj)
	fmt.Fprintln(w, string(b))
	fmt.Println(err)

	cObj, _ := stockings.GetHistoricalPrice("MELI", "20200102")
	c, _ := json.Marshal(cObj)
	fmt.Fprintln(w, string(c))
}
