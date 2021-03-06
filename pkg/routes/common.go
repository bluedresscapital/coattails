package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"github.com/go-redis/redis/v7"
	"github.com/golang/gddo/httputil/header"
)

type malformedRequest struct {
	status int
	msg    string
}

func (mr *malformedRequest) Error() string {
	return mr.msg
}

// Middleware wrapper function to fetch userId given cookie. If cookie is either absent or
// invalid, returns a StatusUnauthorized error
func authMiddleware(handler func(*int, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// given auth token, finds user info
		c, statusCode, err := fetchCookie(r)
		if err != nil {
			w.WriteHeader(statusCode)
			_, _ = fmt.Fprintf(w, err.Error())
			return
		}
		userId, err := wardrobe.VerifyCookie(c.Value)
		if err != nil {
			if err == redis.Nil {
				// This means there wasn't a valid user id mapped by the cookie
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = fmt.Fprintf(w, "Invalid session_token cookie: %s", c.Value)
				return
			}
			// If there is an error fetching from cache, return an internal server error status
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprintf(w, err.Error())
			return
		}
		if userId == nil {
			// If the session token is not present in cache, return an unauthorized error
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		handler(userId, w, r)
	}
}

type GenericPortIdRequest struct {
	PortId int `json:"port_id"`
}

func portAuthMiddleware(handler func(*int, *wardrobe.Portfolio, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return authMiddleware(func(userId *int, w http.ResponseWriter, r *http.Request) {
		req := new(GenericPortIdRequest)
		// NOTE(ma): this is some interesting tech - not sure if this can be improved LOL
		// basically it looks like each time we decode our request body, we can no longer read
		// from it - so we need to first save the request body, parse it, then re-insert it back
		// so our handler fn can further call it
		bodyBytes, _ := ioutil.ReadAll(r.Body)
		r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		err := json.NewDecoder(r.Body).Decode(req)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Printf("Error decoding request into generic port id request: %v", err)
			return
		}
		// We re-insert the request body here
		r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		port, err := wardrobe.FetchPortfolioById(req.PortId)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Printf("Unable to fetch portfolio with id %d", req.PortId)
			return
		}
		if port.UserId != *userId {
			w.WriteHeader(http.StatusUnauthorized)
			log.Printf("Unauthorized access of port id %d by user %d", req.PortId, *userId)
			return
		}
		handler(userId, port, w, r)
	})
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

type StatusResponse struct {
	status string `json:"status"`
}

func writeStatusResponseJson(w http.ResponseWriter, status string) {
	statusResponse := StatusResponse{status: status}
	writeJsonResponse(w, statusResponse)
}

func writeJsonResponse(w http.ResponseWriter, v interface{}) {
	js, err := json.Marshal(v)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(js)
}
