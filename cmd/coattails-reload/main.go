package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/bluedresscapital/coattails/pkg/secrets"

	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"github.com/joho/godotenv"
)

var (
	pgHost             string
	pgPort             int
	pgUser             string
	pgPwd              string
	pgDb               string
	cacheHost          string
	debugNoDeps        bool
	loadBdcKeyFromFile bool
	bdcKeyFile         string
)

func reloadAllPortfolios() {

}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	flag.StringVar(&pgHost, "pg-host", "localhost", "postgresql host name")
	flag.IntVar(&pgPort, "pg-port", 5432, "postgresql port")
	flag.StringVar(&pgUser, "pg-user", "postgres", "postgresql user")
	flag.StringVar(&pgPwd, "pg-pwd", "bdc", "postgresql password")
	flag.StringVar(&pgDb, "pg-db", "wardrobe", "postgresql db")
	flag.StringVar(&cacheHost, "redis-host", "localhost", "redis host")
	flag.BoolVar(&debugNoDeps, "run-without-deps", false, "debug setting")
	flag.BoolVar(&loadBdcKeyFromFile, "load-bdc-key-from-file", false, "flag for whether or not we should get bdc key from file")
	flag.StringVar(&bdcKeyFile, "bdc-key-file", "", "file location of bdc-key. Required if load-bdc-key-from-file is set")
	flag.Parse()
	// Initialize singleton instances after parsing flag
	wardrobe.InitDB(fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		pgHost, pgPort, pgUser, pgPwd, pgDb))
	wardrobe.InitCache(cacheHost)
	secrets.InitSundress(loadBdcKeyFromFile, bdcKeyFile)

	ports, err := wardrobe.FetchPortfoliosByUserId(5)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(ports)
}
