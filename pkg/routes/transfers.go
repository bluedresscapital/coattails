package routes

import (
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
	"log"
	"net/http"
	"time"
)

type UpsertTransferRequest struct {
	Uid           string          `json:"uid"`
	PortId        int             `json:"port_id"`
	Amount        decimal.Decimal `json:"amount"`
	IsDeposit     bool            `json:"is_deposit"`
	ManuallyAdded bool            `json:"manually_added"`
	Date          time.Time       `json:"date"`
}

type DeleteTransferRequest struct {
	Uid string `json:"uid"`
}

func registerTransferRoutes(r *mux.Router) {
	log.Printf("Registering transfer routes")
	s := r.PathPrefix("/transfer").Subrouter()
	s.HandleFunc("", authMiddleware(fetchTransfersHandler)).Methods("GET")
	s.HandleFunc("/upsert", authMiddleware(upsertTransferHandler)).Methods("POST")
	s.HandleFunc("/delete", authMiddleware(deleteTransferHandler)).Methods("POST")
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

func upsertTransferHandler(userId *int, w http.ResponseWriter, r *http.Request) {
	var upsertTransferRequest UpsertTransferRequest
	err := decodeJSONBody(w, r, &upsertTransferRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Bad request: %v", err)
		return
	}
	// First, verify that the user does in fact even own the portfolio they're trying
	// to add to.
	port, err := wardrobe.FetchPortfolioById(upsertTransferRequest.PortId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if port.UserId != *userId {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("Unauthorized access of port %d by user %d", port.Id, userId)
		return
	}
	err = wardrobe.UpsertTransfer(
		upsertTransferRequest.Uid,
		upsertTransferRequest.PortId,
		upsertTransferRequest.Amount,
		upsertTransferRequest.IsDeposit,
		upsertTransferRequest.ManuallyAdded,
		upsertTransferRequest.Date)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Errored on insert: %v", err)
		return
	}
	ts, err := wardrobe.FetchTransfersbyUserId(*userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJsonResponse(w, ts)
}

func deleteTransferHandler(userId *int, w http.ResponseWriter, r *http.Request) {
	var deleteTransferRequest DeleteTransferRequest
	err := decodeJSONBody(w, r, &deleteTransferRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Bad request: %v", err)
		return
	}
	// Verify that the user even owns the transfer that they're trying to delete
	t, err := wardrobe.FetchTransferByUid(deleteTransferRequest.Uid)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Unable to fetch transfer by uid: %v", err)
		return
	}
	port, err := wardrobe.FetchPortfolioById(t.PortId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if port.UserId != *userId {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("Unauthorized delete of transfer by user %d", *userId)
		return
	}
	err = wardrobe.DeleteTransferByUid(deleteTransferRequest.Uid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error in deleting transfer: %v", err)
		return
	}
	ts, err := wardrobe.FetchTransfersbyUserId(*userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJsonResponse(w, ts)
}
