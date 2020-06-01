package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/bluedresscapital/coattails/pkg/util"

	"github.com/bluedresscapital/coattails/pkg/routes"

	"github.com/bluedresscapital/coattails/pkg/diapers"

	"github.com/bluedresscapital/coattails/pkg/stockings"

	"github.com/bluedresscapital/coattails/pkg/orders"
	"github.com/bluedresscapital/coattails/pkg/transfers"

	"github.com/bluedresscapital/coattails/pkg/robinhood"

	"github.com/bluedresscapital/coattails/pkg/tda"

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
	loadBdcKeyFromFile bool
	bdcKeyFile         string
	parallelism        int
)

func reloadPortfolios() {
	ids, err := wardrobe.FetchAllPortfolioIds()
	if err != nil {
		log.Printf("error fetching portfolio ids: %v", err)
	}
	for _, id := range ids {
		log.Printf("Reloading portfolio %d", id)
		port, err := wardrobe.FetchPortfolioById(id)
		if err != nil {
			log.Printf("error fetching portfolio by id: %v", err)
		}
		var orderAPI orders.OrderAPI
		var transferAPI transfers.TransferAPI
		var needsOrderReload bool
		var needsTransferReload bool
		if port.Type == "tda" {
			orderAPI = tda.API{AccountId: port.TDAccountId}
			transferAPI = tda.API{AccountId: port.TDAccountId}
		} else if port.Type == "rh" {
			orderAPI = robinhood.API{AccountId: port.RHAccountId}
			transferAPI = robinhood.API{AccountId: port.RHAccountId}
		} else {
			// Just check if we have uncommitted transfers or orders
			needsOrderReload, err = wardrobe.HasUncommittedOrders(port.Id)
			if err != nil {
				log.Printf("error checking for uncommitted orders: %v", err)
			}
			needsTransferReload, err = wardrobe.HasUncommittedTransfers(port.Id)
			if err != nil {
				log.Printf("error checking for uncommitted transfers: %v", err)
			}
		}
		if orderAPI != nil {
			needsOrderReload, err = orders.ReloadOrders(orderAPI, stockings.FingoPack{})
			if err != nil {
				log.Printf("error reloading orders: %v", err)
			}
		}
		if transferAPI != nil {
			needsTransferReload, err = transfers.ReloadTransfers(transferAPI)
			if err != nil {
				log.Printf("error reloading transfers: %v", err)
			}
		}
		depsChanged := make([]diapers.Data, 0)
		if needsOrderReload {
			depsChanged = append(depsChanged, diapers.Order)
		}
		if needsTransferReload {
			depsChanged = append(depsChanged, diapers.Transfer)
		}
		err = diapers.BulkReloadDepsAndPublish(depsChanged, port.Id, port.UserId, routes.GetChannelFromUserId(port.UserId))
		if err != nil {
			log.Printf("error reloading deps for %v: %v", depsChanged, err)
		}
	}
}

func reloadStockIndustryBuckets(tickers []string, doneChan chan map[string][]string) {
	res := make(map[string][]string)
	for _, t := range tickers {
		collections, err := stockings.ScrapeCollections(t)
		if err != nil {
			log.Printf("errored scraping collections for %s: %v", t, err)
		}
		log.Printf("%s: %v", t, collections)
		res[t] = collections
	}
	doneChan <- res
}

func reloadStockIndustries(parallelism int) {
	tickers, err := wardrobe.FetchStaleStockCollections()
	if err != nil {
		log.Printf("error fetching stale stock collectionMap: %v", err)
	}
	tickerBuckets := util.PartitionTickers(tickers, parallelism)
	doneChan := make(chan map[string][]string)
	for _, bucket := range tickerBuckets {
		go reloadStockIndustryBuckets(bucket, doneChan)
	}
	done := 0
	// maps a ticker to a list of collection names
	collectionMap := make(map[string][]string)
	collectionSet := make(map[string]bool)
L:
	for {
		select {
		case res := <-doneChan:
			for t, collects := range res {
				collectionMap[t] = collects
				for _, c := range collects {
					collectionSet[c] = true
				}
			}
			done++
			if done == parallelism {
				break L
			}
		}
	}
	collections := make([]string, 0)
	for c := range collectionSet {
		collections = append(collections, c)
	}
	collectionIds, err := wardrobe.UpsertCollections(collections)
	if err != nil {
		log.Printf("error upserting collections: %v", err)
	}
	for t, collects := range collectionMap {
		collectIds := make([]int, 0)
		for _, c := range collects {
			collectIds = append(collectIds, collectionIds[c])
		}
		log.Printf("Upserting stock collections for %s", t)
		err = wardrobe.UpsertStockCollections(collectIds, t)
		if err != nil {
			log.Printf("error upserting stock collection edge for %s with collections %v: %v", t, collectIds, err)
		}
	}
	log.Print("Done reloading stock industries!")
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
	flag.BoolVar(&loadBdcKeyFromFile, "load-bdc-key-from-file", false, "flag for whether or not we should get bdc key from file")
	flag.StringVar(&bdcKeyFile, "bdc-key-file", "", "file location of bdc-key. Required if load-bdc-key-from-file is set")
	flag.IntVar(&parallelism, "parallelism", 10, "parallelism")
	flag.Parse()
	// Initialize singleton instances after parsing flag
	wardrobe.InitDB(fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		pgHost, pgPort, pgUser, pgPwd, pgDb))
	wardrobe.InitCache(cacheHost)
	secrets.InitSundress(loadBdcKeyFromFile, bdcKeyFile)
	reloadPortfolios()
	reloadStockIndustries(parallelism)
}
