package main

import (
	"fmt"
	//"github.com/bluedresscapital/coattails/pkg/calc"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"github.com/bluedresscapital/coattails/pkg/sundress"
)

func homeLink(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome home!!!")
}

func main() {
	println(sundress.Decrypt(sundress.Encrypt("1234")))
	router := mux.NewRouter().StrictSlash(true)
	log.Print("Hello world!")
	fmt.Print("Serving traffic on 8080")
	router.HandleFunc("/", homeLink)
	fmt.Print("Is this still working?")
	log.Fatal(http.ListenAndServe(":8080", router))
}
