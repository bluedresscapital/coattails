package routes

import (
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
	"log"
	"net/http"
	"time"
)

type AddTransferRequest struct {
	Uid           string          `json:"uid"`
	PortId        int             `json:"port_id"`
	Amount        decimal.Decimal `json:"amount"`
	IsDeposit     bool            `json:"is_deposit"`
	ManuallyAdded bool            `json:"manually_added"`
	Date          time.Time       `json:"date"`
}

func registerTransferRoutes(r *mux.Router) {
	log.Printf("Registering transfer routes")
	s := r.PathPrefix("/transfer").Subrouter()
	s.HandleFunc("", authMiddleware(fetchTransfersHandler)).Methods("GET")
	s.HandleFunc("/add", authMiddleware(addTransferHandler)).Methods("POST")
}

func fetchTransfersHandler(userId *int, w http.ResponseWriter, r *http.Request) {
	transfers, err := wardrobe.FetchTransfersbyUserId(*userId)
	if err != nil {
		log.Printf("Error fetching transfers: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJsonResponse(w, transfers)
}

func addTransferHandler(userId *int, w http.ResponseWriter, r *http.Request) {
	var addTransferRequest AddTransferRequest
	err := decodeJSONBody(w, r, &addTransferRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Bad request: %v", err)
		return
	}
	// First, verify that the user does in fact even own the portfolio they're trying
	// to add to.
	port, err := wardrobe.FetchPortfolioById(addTransferRequest.PortId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if port.UserId != *userId {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("Unauthorized access of port %d by user %d", port.Id, userId)
		return
	}
	err = wardrobe.AddTransfer(
		addTransferRequest.Uid,
		addTransferRequest.PortId,
		addTransferRequest.Amount,
		addTransferRequest.IsDeposit,
		addTransferRequest.ManuallyAdded,
		addTransferRequest.Date)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Errored on insert: %v", err)
		return
	}
	t, err := wardrobe.FetchTransferByUid(addTransferRequest.Uid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJsonResponse(w, t)
}
