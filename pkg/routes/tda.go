package routes

import (
	"github.com/bluedresscapital/coattails/pkg/tda"
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func registerTDARoutes(r *mux.Router) {
	log.Printf("Registering tda routes")
	s := r.PathPrefix("/tda").Subrouter()
	s.HandleFunc("", authMiddleware(fetchTDAccountsHandler)).Methods("GET")
	s.HandleFunc("/portfolio/create", authMiddleware(createTDPortfolioHandler)).Methods("POST")
	//s.HandleFunc("/delete", authMiddleware(deleteOrderHandler)).Methods("POST")
}

func fetchTDAccountsHandler(userId *int, w http.ResponseWriter, r *http.Request) {
	accounts, err := wardrobe.FetchTDAccountsByUserId(*userId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("error in fetching td accounts: %v", err)
		return
	}
	writeJsonResponse(w, accounts)
}

type CreateTDPortRequest struct {
	Name     string `json:"name"`
	Code     string `json:"code"`
	ClientId string `json:"client_id"`
}

func createTDPortfolioHandler(userId *int, w http.ResponseWriter, r *http.Request) {
	var req CreateTDPortRequest
	err := decodeJSONBody(w, r, &req)
	if err != nil {
		log.Printf("Error decoding create td port request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	auth, err := tda.FetchRefreshTokenUsingAuthCode(req.Code, req.ClientId)
	if err != nil {
		log.Printf("Error fetching refresh token: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = wardrobe.CreateTDPortfolio(*userId, req.Name, req.ClientId, auth.RefreshToken)
	if err != nil {
		log.Printf("Error creating td portfolio: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
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
