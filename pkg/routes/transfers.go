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
	PortId int    `json:"port_id"`
	Uid    string `json:"uid"`
}

func registerTransferRoutes(r *mux.Router) {
	log.Printf("Registering transfer routes")
	s := r.PathPrefix("/transfer").Subrouter()
	s.HandleFunc("", authMiddleware(fetchTransfersHandler)).Methods("GET")
	s.HandleFunc("/upsert", portAuthMiddleware(upsertTransferHandler)).Methods("POST")
	s.HandleFunc("/delete", portAuthMiddleware(deleteTransferHandler)).Methods("POST")
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
	err = wardrobe.DeleteTransfer(deleteTransferRequest.Uid, deleteTransferRequest.PortId)
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
