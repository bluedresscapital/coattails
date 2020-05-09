package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bluedresscapital/coattails/pkg/poncho"
	"github.com/bluedresscapital/coattails/pkg/stockings"
	"github.com/bluedresscapital/coattails/pkg/tda"
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"github.com/gorilla/mux"
	"io/ioutil"
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
	s.HandleFunc("/order/reload", tdAuthMiddleware(reloadTDOrderHandler)).Methods("POST")
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

func reloadTDOrderHandler(tdAccountId *int, w http.ResponseWriter, r *http.Request) {
	order := tda.API{AccountId: *tdAccountId}
	// TODO - change this to a different API :)
	stock := stockings.IexApi{}
	err := poncho.ReloadOrders(order, stock)
	if err != nil {
		log.Printf("Damn we failed: %v", err)
	} else {
		log.Printf("Succcess!")
	}
}

func tdAuthMiddleware(handler func(*int, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return authMiddleware(func(userId *int, w http.ResponseWriter, r *http.Request) {
		req := new(TDAPIRequest)
		bodyBytes, _ := ioutil.ReadAll(r.Body)
		r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		err := json.NewDecoder(r.Body).Decode(req)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Printf("Error decoding request into generic port id request: %v", err)
			return
		}
		// We re-insert the request body here
		r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		tdAcc, err := wardrobe.FetchTDAccount(req.TDAccountID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Printf("Unable to fetch td account with id %d", req.TDAccountID)
			return
		}
		if tdAcc.UserId != *userId {
			w.WriteHeader(http.StatusUnauthorized)
			log.Printf("Unauthorized access of td account id %d by user %d", req.TDAccountID, *userId)
			return
		}
		handler(&req.TDAccountID, w, r)
	})
}

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
