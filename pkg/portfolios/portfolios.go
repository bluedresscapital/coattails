package portfolios

import (
	"errors"
	"log"
	"time"

	"github.com/bluedresscapital/coattails/pkg/stockings"
	"github.com/bluedresscapital/coattails/pkg/util"
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"github.com/shopspring/decimal"
)

type PortValue struct {
	Cash       decimal.Decimal
	StockValue decimal.Decimal
}

func ReloadHistory(portfolio wardrobe.Portfolio) error {
	orders, err := wardrobe.FetchOrdersByPortfolioId(portfolio.Id)
	if err != nil {
		return err
	}
	transfers, err := wardrobe.FetchTransfersByPortfolioId(portfolio.Id)
	if err != nil {
		return err
	}
	start, err := getPortfolioStartDate(transfers, orders)
	// IMPORTANT: Use this dates array as our source of truth. Ignore other potential dates that are NOT
	// in this list!!
	dates := util.GetMarketDates(start, time.Now())
	// Computes portfolio (mapping of stock to quantity) snapshot per day
	portSnapshots := getPortfolioSnapshots(orders, transfers, dates)
	// Computes portfolio values (cash and stock_values) per day
	portValues := computePortValues(dates, portSnapshots)
	for _, date := range dates {
		portValue := portValues[date]
		log.Printf("[%s] total: %s, cash: %s, stock_value: %s", date, portValue.StockValue.Add(portValue.Cash), portValue.Cash, portValue.StockValue)
	}
	return nil
}

func computePortValues(dates []time.Time, snapshots portSnapshots) map[time.Time]PortValue {
	portValues := make(map[time.Time]PortValue)
	for date := range snapshots {
		portValues[date] = PortValue{
			Cash:       snapshots[date]["_CASH"],
			StockValue: decimal.Zero,
		}
	}
	// Get the date ranges in which the user owned the stock
	sr := computeStockRanges(dates, snapshots)
	for s, v := range sr {
		log.Printf("Processing %s: %s -> %s", s, v.start, v.end)
		prices, err := stockings.GetHistoricalRange(stockings.FingoPack{}, s, v.start, v.end)
		if err != nil {
			log.Printf("Errored out fetching stock prices for %s from %s to %s: %v", s, v.start, v.end, err)
			continue
		}
		for _, price := range *prices {
			if !util.IsMarketOpen(price.Date) {
				continue
			}
			price.Date = util.GetTimelessDate(price.Date)
			portSnapshot, found := snapshots[price.Date]
			if !found {
				log.Printf("[ERROR] Found no port snapshot for date %s, continuing", price.Date)
				continue
			}
			quantity, found := portSnapshot[s]
			if !found {
				log.Printf("[ERROR] Couldn't find %s in portsnapshot %v for date %s", s, portSnapshot, price.Date)
				continue
			}
			portValue, found := portValues[price.Date]
			if !found {
				log.Printf("[ERROR] Couldn't find portValue for date %s", price.Date)
				continue
			}
			portValue.StockValue = portValue.StockValue.Add(price.Price.Mul(quantity))
			portValues[price.Date] = portValue
		}
	}
	return portValues
}

type dateRange struct {
	start time.Time
	end   time.Time
}

type portSnapshot map[string]decimal.Decimal
type portSnapshots map[time.Time]portSnapshot

func copyPort(port portSnapshot) portSnapshot {
	clonePort := make(map[string]decimal.Decimal)
	for k, v := range port {
		clonePort[k] = v
	}
	return clonePort
}

func getPortfolioSnapshots(orders []wardrobe.Order, transfers []wardrobe.Transfer, dates []time.Time) portSnapshots {
	transferBuckets := getTransferBuckets(transfers)
	orderBuckets := getOrderBuckets(orders)
	portSnapshots := make(portSnapshots)
	port := make(portSnapshot)
	//port["_CASH"] = getTotalDeposited(transfers)
	port["_CASH"] = decimal.Zero
	for _, date := range dates {
		dayOrders, found := orderBuckets[date]
		if found {
			processDayOrders(port, dayOrders)
		}
		dayTransfers, found := transferBuckets[date]
		if found {
			// DESTRUCTIVELY MODIFIES CURR PORT!!!
			processDayTransfers(port, dayTransfers)
		}
		portSnapshots[date] = copyPort(port)
	}
	return portSnapshots
}

func processDayOrders(currPort portSnapshot, dayOrders []wardrobe.Order) {
	for _, o := range dayOrders {
		cash, _ := currPort["_CASH"]
		quantity, found := currPort[o.Stock]
		if !found {
			quantity = decimal.Zero
		}
		if o.IsBuy {
			quantity = quantity.Add(o.Quantity)
			cash = cash.Sub(o.Quantity.Mul(o.Value))
		} else {
			quantity = quantity.Sub(o.Quantity)
			cash = cash.Add(o.Quantity.Mul(o.Value))
		}
		currPort["_CASH"] = cash
		currPort[o.Stock] = quantity
	}
}

func getTotalDeposited(transfers []wardrobe.Transfer) decimal.Decimal {
	cash := decimal.Zero
	for _, t := range transfers {
		if t.IsDeposit {
			cash = cash.Add(t.Amount)
		} else {
			cash = cash.Sub(t.Amount)
		}
	}
	return cash
}

func processDayTransfers(currPort portSnapshot, dayTransfers []wardrobe.Transfer) {
	for _, t := range dayTransfers {
		cash, _ := currPort["_CASH"]
		if t.IsDeposit {
			cash = cash.Add(t.Amount)
		} else {
			cash = cash.Add(t.Amount)
		}
		currPort["_CASH"] = cash
	}
}

func computeStockRanges(dates []time.Time, snapshots portSnapshots) map[string]dateRange {
	stockRange := make(map[string]dateRange)
	for _, date := range dates {
		snapshot, found := snapshots[date]
		if !found {
			log.Panicf("ERROR: Unable to find snapshot for (valid) date: %s. Snapshots: %v", date, snapshots)
		}
		for stock, quantity := range snapshot {
			if quantity.IsZero() || stock == "_CASH" {
				continue
			}
			dr, found := stockRange[stock]
			if !found {
				dr = dateRange{
					start: date,
					end:   date.AddDate(0, 0, 1),
				}
			} else {
				dr.end = date
			}
			stockRange[stock] = dr
		}
	}
	return stockRange
}

func getOrderBuckets(orders []wardrobe.Order) map[time.Time][]wardrobe.Order {
	buckets := make(map[time.Time][]wardrobe.Order)
	for _, o := range orders {
		bucket, found := buckets[util.GetTimelessDate(o.Date)]
		if !found {
			bucket = make([]wardrobe.Order, 0)
		}
		bucket = append(bucket, o)
		buckets[util.GetTimelessDate(o.Date)] = bucket
	}
	return buckets
}

func getTransferBuckets(transfers []wardrobe.Transfer) map[time.Time][]wardrobe.Transfer {
	buckets := make(map[time.Time][]wardrobe.Transfer)
	for _, t := range transfers {
		bucket, found := buckets[util.GetTimelessDate(t.Date)]
		if !found {
			bucket = make([]wardrobe.Transfer, 0)
		}
		bucket = append(bucket, t)
		buckets[util.GetTimelessDate(t.Date)] = bucket
	}
	return buckets
}

func reloadStockRangePrices(sr map[string]dateRange) error {
	stockApi := stockings.FingoPack{}
	for s, dates := range sr {
		_, err := stockings.GetHistoricalRange(stockApi, s, dates.start, dates.end)
		if err != nil {
			log.Printf("Error in getting historical range: %v", err)
			return err
		}
	}
	return nil
}

func getPortfolioStartDate(transfers []wardrobe.Transfer, orders []wardrobe.Order) (time.Time, error) {
	if len(orders) == 0 {
		if len(transfers) == 0 {
			return time.Now(), errors.New("orders and transfers are both empty for portfolio")
		}
		return transfers[0].Date, nil
	} else if len(transfers) == 0 {
		return orders[0].Date, nil
	} else if transfers[0].Date.Before(orders[0].Date) {
		return transfers[0].Date, nil
	}
	return orders[0].Date, nil
}
