package routes

import (
	"fmt"
	"github.com/bluedresscapital/coattails/pkg/tda"
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type TDAPIRequest struct {
	TDAccountID int `json:"account_id"`
}

func registerTDARoutes(r *mux.Router) {
	log.Printf("Registering tda routes")
	s := r.PathPrefix("/tda").Subrouter()
	s.HandleFunc("", authMiddleware(fetchTDAccountsHandler)).Methods("GET")
	s.HandleFunc("/portfolio/create", authMiddleware(createTDPortfolioHandler)).Methods("POST")
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
	Name       string `json:"name"`
	AccountNum string `json:"account_num"`
	Code       string `json:"code"`
}

func createTDPortfolioHandler(userId *int, w http.ResponseWriter, r *http.Request) {
	var req CreateTDPortRequest
	err := decodeJSONBody(w, r, &req)
	if err != nil {
		log.Printf("Error decoding create td port request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	auth, err := tda.FetchRefreshTokenUsingAuthCode(req.Code, tda.ClientId)
	if err != nil {
		log.Printf("Error fetching refresh token: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = wardrobe.CreateTDPortfolio(*userId, req.Name, req.AccountNum, auth.RefreshToken)
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

// Verifies that the portfolio's tda_account is in fact owned by the user id
func validateTdaUsage(port wardrobe.Portfolio, userId int) error {
	auth, err := wardrobe.FetchTDAccount(port.TDAccountId)
	if err != nil {
		return fmt.Errorf("unable to fetch td account %d", port.TDAccountId)
	}
	if auth.UserId != userId {
		return fmt.Errorf("unauthorized access of td account %d by user %d", auth.Id, userId)
	}
	return nil
}
