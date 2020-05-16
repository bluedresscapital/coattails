package routes

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bluedresscapital/coattails/pkg/secrets"
	"github.com/gorilla/mux"
)

func testHandler(w http.ResponseWriter, r *http.Request) {
	plaintext := "1234"
	log.Printf("plaintext: %s", plaintext)

	cipher := secrets.Encrypt(plaintext)
	log.Printf("first cipher: %s", cipher)

	cipher2 := secrets.Encrypt(plaintext)
	log.Printf("second cipher: %s", cipher2)

	decryptedCipher := secrets.Decrypt(cipher)
	log.Printf("Decrypted cipher: %s", decryptedCipher)

	decryptedCipher2 := secrets.Decrypt(cipher2)
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
	r.HandleFunc("/apitest", apiCheck)
	//r.HandleFunc("/middleout", middleOutTest)
	r.HandleFunc("/financego", checkPiquette)
	// Register all /auth routes
	registerAuthRoutes(r)

	registerStockRoutes(r)
}
