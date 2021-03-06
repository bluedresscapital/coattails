package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/bluedresscapital/coattails/pkg/secrets"

	"github.com/bluedresscapital/coattails/pkg/positions"

	"github.com/bluedresscapital/coattails/pkg/routes"
	"github.com/bluedresscapital/coattails/pkg/socks"

	"github.com/bluedresscapital/coattails/pkg/portfolios"

	"github.com/bluedresscapital/coattails/pkg/util"

	"github.com/bluedresscapital/coattails/pkg/stockings"

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

func reloadCurrentDayStockPrices(i int, tickers []string, doneChan chan bool) {
	log.Printf("worker %d reloading stock prices for %v", i, tickers)
	api := stockings.FingoPack{}
	now := util.GetTimelessDate(time.Now())
	for _, t := range tickers {
		s, err := api.GetCurrentPrice(t)
		if err != nil {
			log.Printf("errored getting stock price: %v", err)
			continue
		}
		err = wardrobe.UpsertStockQuotePrice(s.Symbol, now, s.LatestPrice)
		if err != nil {
			log.Printf("errored updating stock quote price: %v", err)
		}
	}
	doneChan <- true
}

func reloadCurrentDayPortfolioHandler(i int, now time.Time, ports []int, doneChan chan bool) {
	log.Printf("Worker %d loading ports: %v", i, ports)
	for _, portId := range ports {
		port, err := wardrobe.FetchPortfolioById(portId)
		if err != nil {
			log.Printf("error fetching portfolio %d: %v", portId, err)
			continue
		}
		pv, err := portfolios.ReloadCurrentDay(*port)
		if err != nil {
			log.Printf("error reloading current day portfolio: %v", err)
			continue
		}
		res := make(map[int]portfolios.PortValueDiff)
		res[portId] = *pv
		// Update minutely portfolio values
		if now.Minute()%5 == 0 {
			log.Printf("Worker %d saving daily portfolio value snapshots for portfolio %d", i, portId)
			err = wardrobe.InsertDailyPortValue(wardrobe.DailyPortVal{
				PortId: portId,
				Date:   now,
				Value:  (*pv).CurrVal,
			})
			if err != nil {
				log.Printf("error saving daily port value: %v", err)
				continue
			}
		}
		err = socks.PublishFromServer(routes.GetChannelFromUserId(port.UserId), "RELOAD_CURRENT_PORT_VALUES", res)
		if err != nil {
			log.Printf("error publishing current port values: %v", err)
			continue
		}
		// Reload positions data
		err = positions.Reload(portId, stockings.FingoPack{})
		if err != nil {
			log.Printf("error reloading portfolio positions: %v", err)
			continue
		}
	}
	doneChan <- true
}

func reloadCurrentDayPortfolios(parallelism int, now time.Time) {
	ports, err := wardrobe.FetchAllPortfolioIds()
	if err != nil {
		log.Printf("error fetching portfolio ids: %v", err)
	}
	portBuckets := util.PartitionPorts(ports, parallelism)
	doneChan := make(chan bool)
	for i, portBucket := range portBuckets {
		go reloadCurrentDayPortfolioHandler(i, now, portBucket, doneChan)
	}
	done := 0
L:
	for {
		select {
		case <-doneChan:
			done++
			if done == parallelism {
				break L
			}
		}
	}
}

func reloadStockPrices(parallelism int, now time.Time) {
	// Reload all current day (relevant) stock prices
	tickers, err := wardrobe.FetchNonZeroQuantityPositions()
	if err != nil {
		log.Printf("error fetching non zero ticker positions: %v", err)
	}
	tickers = removeTickerDuplicates(tickers)
	tickerPartitions := util.PartitionTickers(tickers, parallelism)
	doneChan := make(chan bool)
	for i, part := range tickerPartitions {
		go reloadCurrentDayStockPrices(i, part, doneChan)
	}
	doneCount := 0
L:
	for {
		select {
		case <-doneChan:
			doneCount++
			if doneCount == parallelism {
				break L // lol this is some go tech, breaks out of the loop we labeled as L
			}
		}
	}
	// Only check if we have a stale price after 10am EST. The reason for the 10am check is we assume
	// w/e stock api we use will have all prices up to including the previous day for any given stock after 10am.
	if now.Hour() > 10 {
		log.Print("TODO: Updating past day prices that aren't set yet")
		// 1. Fetch all stock tickers that have a stale stock quote, aka
		// 		- quotes with a date d that haven't been updated since before d+1 AND now-d > 1
		// 2. If there are multiple stock quotes for the same ticker, compute min and max range, and upsert those
		// stock prices.
		// 		- think carefully about this one, we don't want to break our invariant and constantly upsert stock prices

		// Reload positions + publish
		// Reload portfolio performances + publish
	}
	// Upsert portfolio values + update daily portfolio value if applicable, Update positions
	reloadCurrentDayPortfolios(parallelism, now)
}

func removeTickerDuplicates(tickers []string) []string {
	tickerSet := make(map[string]bool)
	for _, t := range tickers {
		if t != "_CASH" {
			tickerSet[t] = true
		}
	}
	ret := make([]string, 0)
	for t := range tickerSet {
		ret = append(ret, t)
	}
	return ret
}

func main() {
	now := util.GetESTNow()
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
	flag.IntVar(&parallelism, "parallelism", 5, "number of go routines to spin up to reload stock prices")
	flag.Parse()
	// Initialize singleton instances after parsing flag
	wardrobe.InitDB(fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		pgHost, pgPort, pgUser, pgPwd, pgDb))
	wardrobe.InitCache(cacheHost)
	secrets.InitSundress(loadBdcKeyFromFile, bdcKeyFile)
	reloadStockPrices(parallelism, now)
}
