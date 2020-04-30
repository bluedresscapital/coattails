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
	log.Printf("first cipher: %s", cipher)

	cipher2 := sundress.Encrypt(plaintext)
	log.Printf("second cipher: %s", cipher2)

	decryptedCipher := sundress.Decrypt(cipher)
	log.Printf("Decrypted cipher: %s", decryptedCipher)

	decryptedCipher2 := sundress.Decrypt(cipher2)
	log.Printf("Decrypted second cipher: %s", decryptedCipher2)

	_, _ = fmt.Fprintf(w, "Welcome home!!!")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "ping!")
}

func RegisterAllRoutes(r *mux.Router) {
	r.HandleFunc("/health", healthHandler)
	r.HandleFunc("/test", testHandler)
	// Register all /auth routes
	registerAuthRoutes(r)
}
