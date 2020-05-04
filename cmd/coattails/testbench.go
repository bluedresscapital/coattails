package main

import (
	"flag"
	"fmt"
	"github.com/bluedresscapital/coattails/pkg/sundress"
	"github.com/bluedresscapital/coattails/pkg/tda"
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"github.com/joho/godotenv"
	"log"
	"time"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	var (
		wait      time.Duration
		pgHost    string
		pgPort    int
		pgUser    string
		pgPwd     string
		pgDb      string
		cacheHost string
	)

	flag.DurationVar(&wait,
		"graceful-timeout",
		time.Second*15,
		"the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.StringVar(&pgHost, "pg-host", "localhost", "postgresql host name")
	flag.IntVar(&pgPort, "pg-port", 5432, "postgresql port")
	flag.StringVar(&pgUser, "pg-user", "postgres", "postgresql user")
	flag.StringVar(&pgPwd, "pg-pwd", "bdc", "postgresql password")
	flag.StringVar(&pgDb, "pg-db", "wardrobe", "postgresql db")
	flag.StringVar(&cacheHost, "redis-host", "localhost", "redis host")
	flag.Parse()
	// Initialize singleton instances after parsing flag
	sundress.InitSundress()
	wardrobe.InitDB(fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		pgHost, pgPort, pgUser, pgPwd, pgDb))

	// This code should be called ONCE by an api endpoint
	//code := "BxVlR5Y6eBat1yzDPUUUfxtbkvbY5YSwadki5/GzCPuhnJiZg5ggy1YUC7TPqMF26I3zT7Km8YkA/ZZ+4KTtkKtXAv0pFErjLuUnZMm8Jg3Y5nsGyFZf3E8nRsma5HZWYoRCe5qyF9DqtG71dF2E+fyO3z19wG9IGqbs/bMV56pG/hDV+AMby4pwkuuG5KaIpAdaV0JntayYTqYBi9WFRo5AUBbfGOJsgpHzRajZlKbsgC1R8RA5YIOd3ly3bZzykkYCT5mKsH6nxARU/SCBprq6J+fzOrYfUPfKACm8RRr4T6/0ce2jRgxANVATb3/ZoDjB52laWaeuACqbAdrYWNJsfCyhuU20O5TpfOkcbB3XetXPtGNRqs9AwhdVXrhGHCz3MXGmE3/aGr0Cad2OuXYjglK6It6PpvoQ3h7llqJG5j6/J16ntVL/LJA100MQuG4LYrgoVi/JHHvly38ki0Z/ZA5WbektBLvvCRQGzNpP0uYQF5+vkEFBdOJSs9HSTCLdO+ZcAcolW3qZVOVVpSZV3cpCdMxP9lEjFEFCCLDv2wFbkiI0Loe3VvbzgvheT8a4WptkgrroCu8H90hxTl/6CuJASahs9uwHHCzTJ6bT6vxkVHgOF6eNhNMWbpabM2T0y3ZdaJ4TjKk8v5PAxnXR6/eFo7T3YMLHVrsx5Vd6/fwghC0/qqcRAn5fODD5IE1gwLdlC/c7tqe5U4GxJlv5+bn+Q991dRchbneOe2v5kghkZOggrEEoH8V0yjT258cZWgtRHjeMvEZDNKDmFXc7DejrEgnb4doPnkcLBDjpzD4r/IQP/F+AmU/ucadL6KuYxNrK5KijXt+EeHW/PQsNxrPTBmXT4NFWNW3kFZRRUmqF/jLqTj8F6e4sINXelGqvRNXKGtY=212FD3x19z9sWBHDJACbC00B75E"
	//clientId := "GBCZDGRJAOIJHF0IETOA76NFAKZ0OGQX"
	//err = tda.InitTDAccount(5, code, clientId)
	//if err != nil {
	//	log.Fatalf("Error initializing td account: %v", err)
	//}

	err = tda.FakeRequest(5, 1)
	if err != nil {
		log.Fatalf("Error making fake request: %v", err)
	}

}
