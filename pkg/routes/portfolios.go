package routes

import (
	"encoding/json"
	"fmt"
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

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
		return
	}
	writePortfoliosResponse(w, ports)
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
	id, err := wardrobe.FetchPortfolio(*userId, createPortRequest.Name, createPortRequest.Type)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "unable to create portfolio: %v", err)
		return
	}
	writePortfolioResponse(w, wardrobe.Portfolio{
		Id:     *id,
		Name:   createPortRequest.Name,
		Type:   createPortRequest.Type,
		UserId: *userId,
	})
}

func writePortfolioResponse(w http.ResponseWriter, portfolio wardrobe.Portfolio) {
	js, err := json.Marshal(portfolio)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(js)
}

func writePortfoliosResponse(w http.ResponseWriter, portfolios []wardrobe.Portfolio) {
	js, err := json.Marshal(portfolios)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(js)
}
