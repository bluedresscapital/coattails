package routes

import (
	"log"
	"net/http"
	"time"

	"github.com/bluedresscapital/coattails/pkg/stockings"

	"github.com/gorilla/mux"
)

func registerStockRoutes(r *mux.Router) {
	log.Print("Registering stock routes")
	s := r.PathPrefix("/stock").Subrouter()
	s.HandleFunc("/quote/{ticker}", stockQuoteHandler).Methods("GET")
}

func stockQuoteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		log.Printf("Invalid start string: %s", startStr)
		return
	}
	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		log.Printf("Invalid end string: %s", endStr)
		return
	}
	prices, err := stockings.GetHistoricalRange(stockings.FingoPack{}, vars["ticker"], start, end)
	if err != nil {
		log.Printf("Error in getting historical range: %v", err)
		return
	}
	writeJsonResponse(w, *prices)
}
