package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"github.com/golang/gddo/httputil/header"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"strings"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Registers auth routes for coattails server
func RegisterAuthRoutes(r *mux.Router) {
	log.Print("Registering auth routes")
	s := r.PathPrefix("/auth").Subrouter()
	s.HandleFunc("/login", loginHandler).Methods("POST")
	s.HandleFunc("/logout", loginHandler).Methods("POST")
	s.HandleFunc("/register", loginHandler).Methods("POST")
	s.HandleFunc("/user", loginHandler).Methods("POST")
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var l LoginRequest
	err := decodeJSONBody(w, r, &l)
	if err != nil {
		handleDecodeErr(w, err)
		return
	}
	_, _ = fmt.Fprintf(w, "Login: %+v", l)
	_ = wardrobe.FetchUser(l.Username, l.Password)
}

type malformedRequest struct {
	status int
	msg    string
}

func (mr *malformedRequest) Error() string {
	return mr.msg
}

func handleDecodeErr(w http.ResponseWriter, err error) {
	var mr *malformedRequest
	if errors.As(err, &mr) {
		http.Error(w, mr.msg, mr.status)
	} else {
		log.Println(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// Lol stolen off of https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body
func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	if r.Header.Get("Content-Type") != "" {
		value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
		if value != "application/json" {
			msg := "Content-Type header is not application/json"
			return &malformedRequest{status: http.StatusUnsupportedMediaType, msg: msg}
		}
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(&dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := fmt.Sprintf("Request body contains badly-formed JSON")
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case err.Error() == "http: request body too large":
			msg := "Request body must not be larger than 1MB"
			return &malformedRequest{status: http.StatusRequestEntityTooLarge, msg: msg}

		default:
			return err
		}
	}

	if dec.More() {
		msg := "Request body must only contain a single JSON object"
		return &malformedRequest{status: http.StatusBadRequest, msg: msg}
	}

	return nil
}
