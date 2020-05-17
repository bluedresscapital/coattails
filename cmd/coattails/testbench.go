package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/bluedresscapital/coattails/pkg/robinhood"
	"github.com/bluedresscapital/coattails/pkg/secrets"
	"github.com/bluedresscapital/coattails/pkg/stockings"
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	var (
		wait        time.Duration
		pgHost      string
		pgPort      int
		pgUser      string
		pgPwd       string
		pgDb        string
		cacheHost   string
		debugNoDeps bool
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
	flag.BoolVar(&debugNoDeps, "run-without-deps", false, "debug setting")
	//first arg is a pointer, second arg is the value we are checking for, third value is what we set if we don't see the flag, fourth is description
	flag.Parse()
	// Initialize singleton instances after parsing flag
	stockings.InitKeygen()
	if debugNoDeps {
		log.Println("Warning: You are starting a server without a Database and Cache")
		log.Println("Calls to functions that use a Database or Cache will segfault")
	} else {
		wardrobe.InitDB(fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			pgHost, pgPort, pgUser, pgPwd, pgDb))
		wardrobe.InitCache(cacheHost)
	}
	secrets.InitSundress()

	rhApi := robinhood.API{AccountId: 2}
	rhApi.GetOrders()
}
