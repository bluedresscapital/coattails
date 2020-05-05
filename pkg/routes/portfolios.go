package routes

import (
	"fmt"
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

// All portfolio routes should be under /auth prefix
func registerPortfolioRoutes(r *mux.Router) {
	log.Printf("Registering portfolio routes")
	s := r.PathPrefix("/portfolio").Subrouter()
	s.HandleFunc("", authMiddleware(fetchPortfoliosHandler)).Methods("GET")
	s.HandleFunc("/create", authMiddleware(createPortfolioHandler)).Methods("POST")
}

type CreatePortfolioRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
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
