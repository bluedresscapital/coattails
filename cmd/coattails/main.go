package main

import (
	"fmt"
	"github.com/bluedresscapital/coattails/pkg/sundress"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func homeLink(w http.ResponseWriter, r *http.Request) {
	plaintext := "1234"
	log.Printf("plaintext: %s", plaintext)
	cipher := sundress.Encrypt(plaintext)
	log.Printf("cipher: %s", cipher)
	decryptedCipher := sundress.Decrypt(cipher)
	log.Printf("Decripted cipher: %s", decryptedCipher)
	fmt.Fprintf(w, "Welcome home!!!")
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	log.Print("Hello world!")
	router.HandleFunc("/", homeLink)
	log.Fatal(http.ListenAndServe(":8080", router))
}
