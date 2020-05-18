package routes

import (
	"log"
	"net/http"
	"time"

	"github.com/bluedresscapital/coattails/pkg/robinhood"

	"github.com/bluedresscapital/coattails/pkg/transfers"

	"github.com/bluedresscapital/coattails/pkg/diapers"
	"github.com/bluedresscapital/coattails/pkg/socks"
	"github.com/bluedresscapital/coattails/pkg/tda"
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
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
	s.HandleFunc("/reload", portAuthMiddleware(reloadTransferHandler)).Methods("POST")
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

func upsertTransferHandler(userId *int, port *wardrobe.Portfolio, w http.ResponseWriter, r *http.Request) {
	var req UpsertTransferRequest
	err := decodeJSONBody(w, r, &req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Bad request: %v", err)
		return
	}
	err = wardrobe.UpsertTransfer(wardrobe.Transfer{
		Uid:           req.Uid,
		PortId:        req.PortId,
		Amount:        req.Amount,
		IsDeposit:     req.IsDeposit,
		ManuallyAdded: true,
		Date:          req.Date,
	})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Errored on insert: %v", err)
		return
	}
	err = diapers.ReloadDepsAndPublish(diapers.Transfer, port.Id, *userId, GetChannelFromUserId(*userId))
	if err != nil {
		return
	}
	ts, err := wardrobe.FetchTransfersbyUserId(*userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJsonResponse(w, ts)
}

func deleteTransferHandler(userId *int, port *wardrobe.Portfolio, w http.ResponseWriter, r *http.Request) {
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
	err = diapers.ReloadDepsAndPublish(diapers.Transfer, port.Id, *userId, GetChannelFromUserId(*userId))
	if err != nil {
		return
	}
	ts, err := wardrobe.FetchTransfersbyUserId(*userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJsonResponse(w, ts)
}

func reloadTransferHandler(userId *int, port *wardrobe.Portfolio, w http.ResponseWriter, r *http.Request) {
	var transfer transfers.TransferAPI
	if port.Type == "tda" {
		log.Println("Reloading tda transfers")
		err := validateTdaUsage(*port, *userId)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		transfer = tda.API{AccountId: port.TDAccountId}
	} else if port.Type == "rh" {
		log.Println("Reloading rh transfers")
		err := validateRhUsage(*port, *userId)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		transfer = robinhood.API{AccountId: port.RHAccountId}
	}
	if transfer != nil {
		needsUpdate, err := transfers.ReloadTransfers(transfer)
		if needsUpdate {
			err = diapers.ReloadDepsAndPublish(diapers.Transfer, port.Id, *userId, GetChannelFromUserId(*userId))
			if err != nil {
				return
			}
		}
	}
	ts, err := wardrobe.FetchTransfersbyUserId(*userId)
	err = socks.PublishFromServer(GetChannelFromUserId(*userId), "RELOADED_TRANSFERS", ts)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
