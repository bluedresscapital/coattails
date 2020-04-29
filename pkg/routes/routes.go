package routes

import (
	"fmt"
	"github.com/bluedresscapital/coattails/pkg/sundress"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func testHandler(w http.ResponseWriter, r *http.Request) {
	plaintext := "1234"
	log.Printf("plaintext: %s", plaintext)
	cipher := sundress.Encrypt(plaintext)
	log.Printf("cipher: %s", cipher)
	decryptedCipher := sundress.Decrypt(cipher)
	log.Printf("Decripted cipher: %s", decryptedCipher)
	_, _ = fmt.Fprintf(w, "Welcome home!!!")
}

func RegisterAllRoutes(r *mux.Router) {
	r.HandleFunc("/test", testHandler)
	// Register all /auth routes
	registerAuthRoutes(r)
}
