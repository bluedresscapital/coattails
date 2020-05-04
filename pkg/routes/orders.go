package routes

import (
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
	"log"
	"net/http"
	"time"
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
	Uid string `json:"uid"`
}

func registerOrderRoutes(r *mux.Router) {
	log.Printf("Registering order routes")
	s := r.PathPrefix("/order").Subrouter()
	s.HandleFunc("", authMiddleware(fetchOrdersHandler)).Methods("GET")
	s.HandleFunc("/upsert", authMiddleware(upsertOrderHandler)).Methods("POST")
	s.HandleFunc("/delete", authMiddleware(deleteOrderHandler)).Methods("POST")
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

func upsertOrderHandler(userId *int, w http.ResponseWriter, r *http.Request) {
	var u UpsertOrderRequest
	err := decodeJSONBody(w, r, &u)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Bad request: %v", err)
		return
	}
	// First, verify that the user does in fact even own the portfolio they're trying
	// to add to.
	port, err := wardrobe.FetchPortfolioById(u.PortId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if port.UserId != *userId {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("Unauthorized access of port %d by user %d", port.Id, userId)
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
	orders, err := wardrobe.FetchOrdersByUserId(*userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error in fetching orders: %v", err)
		return
	}
	writeJsonResponse(w, orders)
}

func deleteOrderHandler(userId *int, w http.ResponseWriter, r *http.Request) {
	var deleteOrderRequest DeleteOrderRequest
	err := decodeJSONBody(w, r, &deleteOrderRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Bad request: %v", err)
		return
	}
	// Verify that the user even owns the transfer that they're trying to delete
	o, err := wardrobe.FetchOrderByUid(deleteOrderRequest.Uid)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Unable to fetch order by uid: %v", err)
		return
	}
	port, err := wardrobe.FetchPortfolioById(o.PortId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if port.UserId != *userId {
		w.WriteHeader(http.StatusUnauthorized)
		log.Printf("Unauthorized delete of order by user %d", *userId)
		return
	}
	err = wardrobe.DeleteOrderByUid(deleteOrderRequest.Uid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error in deleting order: %v", err)
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
