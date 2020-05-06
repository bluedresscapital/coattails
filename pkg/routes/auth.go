package routes

import (
	"encoding/json"
	"github.com/bluedresscapital/coattails/pkg/auth"
	"github.com/bluedresscapital/coattails/pkg/sundress"
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

// Registers auth routes for coattails server
func registerAuthRoutes(r *mux.Router) {
	log.Print("Registering auth routes")
	s := r.PathPrefix("/auth").Subrouter()

	s.HandleFunc("/login", loginHandler).Methods("POST")
	s.HandleFunc("/logout", logoutHandler).Methods("POST")
	s.HandleFunc("/register", registerHandler).Methods("POST")
	s.HandleFunc("/user", authMiddleware(userHandler)).Methods("POST")

	// Register all portfolio routes under /auth
	registerPortfolioRoutes(s)
	// Register all transfer routes under /auth
	registerTransferRoutes(s)
	// Register all order routes under /auth
	registerOrderRoutes(s)
	// Register all tda routes under /auth
	registerTDARoutes(s)
}

type loginRegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserResponse struct {
	Username string `json:"username"`
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	// clears auth token tied to user
	c, statusCode, err := fetchCookie(r)
	if err != nil {
		w.WriteHeader(statusCode)
		return
	}
	err = wardrobe.ClearAuthToken(c.Value)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "session_token",
		MaxAge: -1,
	})
	writeStatusResponseJson(w, "success")
}

func userHandler(userId *int, w http.ResponseWriter, r *http.Request) {
	username, err := wardrobe.FetchUserById(*userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	writeUserResponseJson(w, *username)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	loginRegisterHelper(w, r, true)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	loginRegisterHelper(w, r, false)
}

// Simple unified helper for both login and register, since both functions are very similar
func loginRegisterHelper(w http.ResponseWriter, r *http.Request, loginMode bool) {
	var l loginRegisterRequest
	err := decodeJSONBody(w, r, &l)
	if err != nil {
		handleDecodeErr(w, err)
		return
	}
	// Immediately encrypt password so we don't run it down later
	cipherPwd := sundress.Hash(l.Password)
	tok := new(string)
	if loginMode {
		tok, err = auth.Login(l.Username, cipherPwd)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
	} else {
		tok, err = auth.Register(l.Username, cipherPwd)
		if err != nil {
			log.Printf("we boomed here: %v", err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   *tok,
		Expires: time.Now().Add(wardrobe.SessionTokenTtl),
	})
	writeUserResponseJson(w, l.Username)
}

func writeUserResponseJson(w http.ResponseWriter, username string) {
	userResponse := UserResponse{Username: username}
	js, err := json.Marshal(userResponse)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(js)
}

// Fetches cookie from request header, if present
func fetchCookie(r *http.Request) (*http.Cookie, int, error) {
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			return nil, http.StatusUnauthorized, err
		}
		// For any other type of error, return a bad request status
		return nil, http.StatusBadRequest, err
	}
	return c, http.StatusOK, nil
}
