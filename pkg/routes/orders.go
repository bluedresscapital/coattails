package routes

import (
	"log"
	"net/http"
	"time"

	"github.com/bluedresscapital/coattails/pkg/robinhood"

	"github.com/bluedresscapital/coattails/pkg/diapers"
	"github.com/bluedresscapital/coattails/pkg/orders"
	"github.com/bluedresscapital/coattails/pkg/socks"
	"github.com/bluedresscapital/coattails/pkg/stockings"
	"github.com/bluedresscapital/coattails/pkg/tda"
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
)

type UpsertOrderRequest struct {
	Uid           string          `json:"uid"`
	PortId        int             `json:"port_id"`
	Stock         string          `json:"stock"`
	Quantity      decimal.Decimal `json:"quantity"`
	Value         decimal.Decimal `json:"value"`
	IsBuy         bool            `json:"is_buy"`
	ManuallyAdded bool            `json:"manually_added"`
	Date          time.Time       `json:"date"`
}

type DeleteOrderRequest struct {
	PortId int    `json:"port_id"`
	Uid    string `json:"uid"`
}

func registerOrderRoutes(r *mux.Router) {
	log.Printf("Registering order routes")
	s := r.PathPrefix("/order").Subrouter()
	s.HandleFunc("", authMiddleware(fetchOrdersHandler)).Methods("GET")
	s.HandleFunc("/upsert", portAuthMiddleware(upsertOrderHandler)).Methods("POST")
	s.HandleFunc("/delete", portAuthMiddleware(deleteOrderHandler)).Methods("POST")
	s.HandleFunc("/reload", portAuthMiddleware(reloadOrderHandler)).Methods("POST")
}

func fetchOrdersHandler(userId *int, w http.ResponseWriter, r *http.Request) {
	orders, err := wardrobe.FetchOrdersByUserId(*userId)
	if err != nil {
		log.Printf("Error fetching orders: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJsonResponse(w, orders)
}

func upsertOrderHandler(userId *int, port *wardrobe.Portfolio, w http.ResponseWriter, r *http.Request) {
	var u UpsertOrderRequest
	err := decodeJSONBody(w, r, &u)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Bad request: %v", err)
		return
	}
	err = wardrobe.UpsertOrder(wardrobe.Order{
		Uid:           u.Uid,
		PortId:        u.PortId,
		Stock:         u.Stock,
		Quantity:      u.Quantity,
		Value:         u.Value,
		IsBuy:         u.IsBuy,
		ManuallyAdded: u.ManuallyAdded,
		Date:          u.Date,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error in upserting order: %v", err)
		return
	}
	err = diapers.ReloadDepsAndPublish(diapers.Order, port.Id, *userId, GetChannelFromUserId(*userId))
	if err != nil {
		return
	}
	orders, err := wardrobe.FetchOrdersByUserId(*userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error in fetching orders: %v", err)
		return
	}
	writeJsonResponse(w, orders)
}

func deleteOrderHandler(userId *int, port *wardrobe.Portfolio, w http.ResponseWriter, r *http.Request) {
	var deleteOrderRequest DeleteOrderRequest
	err := decodeJSONBody(w, r, &deleteOrderRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Bad request: %v", err)
		return
	}
	err = wardrobe.DeleteOrder(deleteOrderRequest.Uid, deleteOrderRequest.PortId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error in deleting order: %v", err)
		return
	}
	err = diapers.ReloadDepsAndPublish(diapers.Order, port.Id, *userId, GetChannelFromUserId(*userId))
	if err != nil {
		return
	}
	orders, err := wardrobe.FetchOrdersByUserId(*userId)
	if err != nil {
		log.Printf("Error fetching orders: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeJsonResponse(w, orders)
}

func reloadOrderHandler(userId *int, port *wardrobe.Portfolio, w http.ResponseWriter, r *http.Request) {
	var order orders.OrderAPI
	if port.Type == "tda" {
		log.Print("Reloading tda orders...")
		err := validateTdaUsage(*port, *userId)
		if err != nil {
			log.Printf("Unable to validate td account usage: %v", err)
			return
		}
		order = tda.API{AccountId: port.TDAccountId}
	} else if port.Type == "rh" {
		log.Print("Reloading rh orders...")
		err := validateRhUsage(*port, *userId)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		order = robinhood.API{AccountId: port.RHAccountId}
	}
	if order != nil {
		needsUpdate, err := orders.ReloadOrders(order, stockings.FingoPack{})
		if err != nil {
			log.Printf("Encountered error while reloading orders: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if needsUpdate {
			err = diapers.ReloadDepsAndPublish(diapers.Order, port.Id, *userId, GetChannelFromUserId(*userId))
			if err != nil {
				return
			}
		}
	}
	orders, err := wardrobe.FetchOrdersByUserId(*userId)
	err = socks.PublishFromServer(GetChannelFromUserId(*userId), "RELOADED_ORDERS", orders)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
