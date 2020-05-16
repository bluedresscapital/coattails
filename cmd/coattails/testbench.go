package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/bluedresscapital/coattails/pkg/secrets"
	"github.com/bluedresscapital/coattails/pkg/tda"
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"github.com/joho/godotenv"
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
	secrets.InitSundress()
	wardrobe.InitDB(fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		pgHost, pgPort, pgUser, pgPwd, pgDb))

	for i := 0; i < 10; i++ {
		tdaAPI := tda.API{AccountId: 2}
		_, err = tdaAPI.GetOrders()
		if err != nil {
			log.Fatalf("Error making fake request: %v", err)
		}
		tdaAPI = tda.API{AccountId: 1}
		_, err = tdaAPI.GetOrders()
		if err != nil {
			log.Fatalf("Error making fake request: %v", err)
		}
	}
}
