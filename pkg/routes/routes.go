package routes

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func testHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("Sleeping 5s..")
	time.Sleep(time.Duration(5) * time.Second)
	log.Print("done!")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "ping!")
}

func RegisterAllRoutes(r *mux.Router) {
	r.HandleFunc("/health", healthHandler)
	r.HandleFunc("/test", testHandler)
	r.HandleFunc("/apitest", apiCheck)
	//r.HandleFunc("/middleout", middleOutTest)
	r.HandleFunc("/financego", checkPiquette)
	// Register all /auth routes
	registerAuthRoutes(r)

	registerStockRoutes(r)
	registerCollectionRoutes(r)
}
