package routes

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bluedresscapital/coattails/pkg/robinhood"

	"github.com/bluedresscapital/coattails/pkg/wardrobe"

	"github.com/gorilla/mux"
)

func registerRobinhoodRoutes(r *mux.Router) {
	log.Printf("Registering rh routes")
	s := r.PathPrefix("/rh").Subrouter()
	s.HandleFunc("", authMiddleware(fetchRHAccountsHandler)).Methods("GET")
	s.HandleFunc("/portfolio/create", authMiddleware(createRHPortfolioHandler)).Methods("POST")
}

func fetchRHAccountsHandler(userId *int, w http.ResponseWriter, r *http.Request) {
	// TODO?
	return
}

type CreateRHPortRequest struct {
	Name       string `json:"name"`
	RefreshTok string `json:"refresh_token"`
}

func createRHPortfolioHandler(userId *int, w http.ResponseWriter, r *http.Request) {
	var req CreateRHPortRequest
	err := decodeJSONBody(w, r, &req)
	if err != nil {
		return
	}
	// Verify that the refresh token is valid by using it
	auth, err := robinhood.FetchBearerToken(req.RefreshTok)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// IMPORTANT: Use the NEW auth refresh token, not the request refresh token.
	// The request token is now invalid
	err = wardrobe.CreateRHPortfolio(*userId, req.Name, auth.RefreshTok)
	if err != nil {
		log.Printf("Error creating rh portfolio: %v", err)
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

// Verifies that the portfolio's rh_account is in fact owned by the user
func validateRhUsage(port wardrobe.Portfolio, userId int) error {
	acc, err := wardrobe.FetchRHAccount(port.RHAccountId)
	if err != nil {
		return fmt.Errorf("unable to fetch rh account %d", port.RHAccountId)
	}
	if acc.UserId != userId {
		return fmt.Errorf("rh account user id %d doesn't match user %d", acc.UserId, userId)
	}
	return nil
}
