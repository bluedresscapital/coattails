package routes

import (
	"log"
	"net/http"
	"strconv"

	"github.com/bluedresscapital/coattails/pkg/collections"

	"github.com/bluedresscapital/coattails/pkg/wardrobe"

	"github.com/gorilla/mux"
)

func registerCollectionRoutes(r *mux.Router) {
	log.Print("Registering collection routes")
	s := r.PathPrefix("/collections").Subrouter()
	s.HandleFunc("/count", totalCollectionCountHandler).Methods("GET")
	s.HandleFunc("/count/portfolio/{port_id:[0-9]+}", portCollectionCountHandler).Methods("GET")
	s.HandleFunc("/count/ticker/{ticker}", tickerCollectionCountHandler).Methods("GET")
}

func totalCollectionCountHandler(w http.ResponseWriter, r *http.Request) {

}

func portCollectionCountHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	portId, err := strconv.Atoi(vars["port_id"])
	if err != nil {
		log.Printf("couldn't parse the port id: %s", vars["port_id"])
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	positions, err := wardrobe.FetchPortfolioPositions(portId)
	if err != nil {
		log.Printf("error fetching portfolio positions: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tickers := make([]string, 0)
	for _, p := range positions {
		tickers = append(tickers, p.Stock)
	}
	counts, err := collections.FetchCollectionCountsFromTickers(tickers)
	if err != nil {
		log.Printf("error fetching collection counts: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJsonResponse(w, counts)
}

func tickerCollectionCountHandler(w http.ResponseWriter, r *http.Request) {

}
