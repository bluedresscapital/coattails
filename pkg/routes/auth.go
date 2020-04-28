package routes

import (
	"fmt"
	"github.com/bluedresscapital/coattails/pkg/auth"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

// Registers auth routes for coattails server
func RegisterAuthRoutes(r *mux.Router) {
	log.Print("Registering auth routes")
	s := r.PathPrefix("/auth").Subrouter()
	s.HandleFunc("/login", loginHandler).Methods("POST")
	s.HandleFunc("/logout", loginHandler).Methods("POST")
	s.HandleFunc("/register", loginHandler).Methods("POST")
	s.HandleFunc("/user", loginHandler).Methods("POST")
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var l LoginRequest
	err := decodeJSONBody(w, r, &l)
	if err != nil {
		handleDecodeErr(w, err)
		return
	}
	auth.Login(l.Username, l.Password)
	_, _ = fmt.Fprintf(w, "Login: %+v", l)
}
