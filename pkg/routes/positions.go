package routes

import (
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func registerPositionRoutes(r *mux.Router) {
	log.Printf("Registering tda routes")
	s := r.PathPrefix("/positions").Subrouter()
	s.HandleFunc("", authMiddleware(fetchPositionsHandler)).Methods("GET")
	s.HandleFunc("/portfolio", portAuthMiddleware(fetchPortfolioPositionsHandler)).Methods("GET")
}

func fetchPositionsHandler(userId *int, w http.ResponseWriter, r *http.Request) {
	positions, err := wardrobe.FetchPositions(*userId)
	if err != nil {
		log.Printf("Error fetching positions: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJsonResponse(w, positions)
	return
}

func fetchPortfolioPositionsHandler(userId *int, port *wardrobe.Portfolio, w http.ResponseWriter, r *http.Request) {
	// TODO fetch all positions tied to this portfolio
	positions, err := wardrobe.FetchPortfolioPositions(port.Id)
	if err != nil {
		return
	}
	writeJsonResponse(w, positions)
	return
}
