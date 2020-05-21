package routes

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bluedresscapital/coattails/pkg/util"

	"github.com/bluedresscapital/coattails/pkg/portfolios"

	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"github.com/gorilla/mux"
)

// All portfolio routes should be under /auth prefix
func registerPortfolioRoutes(r *mux.Router) {
	log.Printf("Registering portfolio routes")
	s := r.PathPrefix("/portfolio").Subrouter()
	s.HandleFunc("", authMiddleware(fetchPortfoliosHandler)).Methods("GET")
	s.HandleFunc("/create", authMiddleware(createPortfolioHandler)).Methods("POST")
	s.HandleFunc("/history", authMiddleware(fetchPortfolioHistoryHandler)).Methods("GET")
	s.HandleFunc("/values", authMiddleware(fetchPortfolioValuesHandler)).Methods("GET")
	s.HandleFunc("/history/reload", portAuthMiddleware(reloadPortfolioHistoryHandler)).Methods("POST")
}

type CreatePortfolioRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func fetchPortfolioValuesHandler(userId *int, w http.ResponseWriter, r *http.Request) {
	ports, err := wardrobe.FetchPortfoliosByUserId(*userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	now := util.GetESTNow()
	res := make(map[int]portfolios.PortValueDiff)
	// TODO - look into fixing edge case where portfolio values len is < 2??
	for _, port := range ports {
		currPv, err := wardrobe.FetchPortfolioValueOnDay(port.Id, now)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		prevPv, err := wardrobe.FetchPortfolioValueOnDay(port.Id, now.AddDate(0, 0, -1))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		res[port.Id] = portfolios.PortValueDiff{
			CurrVal:     currPv.StockValue.Add(currPv.Cash),
			PrevVal:     prevPv.StockValue.Add(prevPv.Cash),
			DailyChange: currPv.DailyChange,
		}
	}
	writeJsonResponse(w, res)
}

func fetchPortfoliosHandler(userId *int, w http.ResponseWriter, r *http.Request) {
	ports, err := wardrobe.FetchPortfoliosByUserId(*userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Internal server error: %v", err)
		return
	}
	writeJsonResponse(w, ports)
}

func createPortfolioHandler(userId *int, w http.ResponseWriter, r *http.Request) {
	var createPortRequest CreatePortfolioRequest
	err := decodeJSONBody(w, r, &createPortRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(w, "invalid request")
		return
	}
	if createPortRequest.Type != "paper" && createPortRequest.Type != "rh" && createPortRequest.Type != "tda" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintf(w, "invalid portfolio type: %s", createPortRequest.Type)
		return
	}
	err = wardrobe.CreatePortfolio(*userId, createPortRequest.Name, createPortRequest.Type)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintf(w, "unable to create portfolio: %v", err)
		return
	}
	portfolios, err := wardrobe.FetchPortfoliosByUserId(*userId)
	if err != nil {
		log.Printf("Error fetching all portfolios by user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJsonResponse(w, portfolios)
}

func fetchPortfolioHistoryHandler(userId *int, w http.ResponseWriter, r *http.Request) {
	ps, err := wardrobe.FetchPortfoliosByUserId(*userId)
	if err != nil {
		log.Printf("Error fetching portfolios: %v", err)
	}
	perfMap := make(map[int][]wardrobe.PortValue)
	for _, portfolio := range ps {
		pvs, err := wardrobe.FetchPortfolioValuesByPortId(portfolio.Id)
		if err != nil {
			log.Printf("Error fetching portfolio values: %v", err)
		}
		perfMap[portfolio.Id] = pvs
	}
	writeJsonResponse(w, perfMap)
}

func reloadPortfolioHistoryHandler(userId *int, portfolio *wardrobe.Portfolio, w http.ResponseWriter, r *http.Request) {
	err := portfolios.ReloadHistory(*portfolio)
	if err != nil {
		log.Printf("Error reloading portfolio: %v", err)
		return
	}
	log.Printf("Succesfully reloaded portfolio history!")
	writeStatusResponseJson(w, "success")
}
