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

const (
	CASH                = "_CASH"
	NORMALIZED_CASH     = "_NORMALIZED_CASH"
	DAILY_NET_DEPOSITED = "_DAILY_NET_DEPOSITED"
)

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
	// Computes portfolio values (cash, stock_values, daily_net_deposited) per day
	portValues := computePortValues(dates, portSnapshots, portfolio.Id)
	for _, pv := range portValues {
		err = wardrobe.UpsertPortfolioValue(pv)
		if err != nil {
			return err
		}
	}

	return nil
}

type PortPerformance struct {
	Date          time.Time       `json:"date"`
	DailyChange   decimal.Decimal `json:"daily_change"`
	CumChange     decimal.Decimal `json:"cum_change"`
	PortTotal     decimal.Decimal `json:"port_total"`
	NormPortTotal decimal.Decimal `json:"norm_port_total"`
}

func ComputePortfolioPerformance(pvs []wardrobe.PortValue) []PortPerformance {
	portPerfs := make([]PortPerformance, len(pvs))
	cumPerf := decimal.New(1, 0)
	for i, pv := range pvs {
		if i == 0 {
			portPerfs[i] = PortPerformance{
				Date:          pv.Date,
				DailyChange:   decimal.NewFromInt(1),
				CumChange:     decimal.NewFromInt(1),
				PortTotal:     pv.StockValue.Add(pv.Cash),
				NormPortTotal: pv.StockValue.Add(pv.NormalizedCash),
			}
		} else {
			prevPv := pvs[i-1]
			currVal := pv.Cash.Add(pv.StockValue).Sub(pv.DailyNetDeposited)
			prevVal := prevPv.Cash.Add(prevPv.StockValue)
			perf := currVal.Div(prevVal)
			cumPerf = cumPerf.Mul(perf)
			portPerfs[i] = PortPerformance{
				Date:          pv.Date,
				DailyChange:   perf.Truncate(4),
				CumChange:     cumPerf.Truncate(4),
				PortTotal:     pv.StockValue.Add(pv.Cash),
				NormPortTotal: pv.StockValue.Add(pv.NormalizedCash),
			}
		}
	}
	return portPerfs
}

func computePortValues(dates []time.Time, snapshots portSnapshots, portId int) map[time.Time]wardrobe.PortValue {
	portValues := make(map[time.Time]wardrobe.PortValue)
	for date := range snapshots {
		portValues[date] = wardrobe.PortValue{
			Date:              date,
			PortId:            portId,
			NormalizedCash:    snapshots[date][NORMALIZED_CASH],
			DailyNetDeposited: snapshots[date][DAILY_NET_DEPOSITED],
			Cash:              snapshots[date][CASH],
			StockValue:        decimal.Zero,
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
	port[NORMALIZED_CASH] = getTotalDeposited(transfers)
	port[CASH] = decimal.Zero
	for _, date := range dates {
		port[DAILY_NET_DEPOSITED] = decimal.Zero
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
		cash, _ := currPort[CASH]
		normCash, _ := currPort[NORMALIZED_CASH]
		quantity, found := currPort[o.Stock]
		if !found {
			quantity = decimal.Zero
		}
		if o.IsBuy {
			quantity = quantity.Add(o.Quantity)
			cash = cash.Sub(o.Quantity.Mul(o.Value))
			normCash = normCash.Sub(o.Quantity.Mul(o.Value))
		} else {
			quantity = quantity.Sub(o.Quantity)
			cash = cash.Add(o.Quantity.Mul(o.Value))
			normCash = normCash.Add(o.Quantity.Mul(o.Value))
		}
		currPort[CASH] = cash
		currPort[NORMALIZED_CASH] = normCash
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
	netDeposited := decimal.Zero
	for _, t := range dayTransfers {
		cash, _ := currPort[CASH]
		if t.IsDeposit {
			cash = cash.Add(t.Amount)
			netDeposited = netDeposited.Add(t.Amount)
		} else {
			cash = cash.Sub(t.Amount)
			netDeposited = netDeposited.Sub(t.Amount)
		}
		currPort[CASH] = cash
		currPort[DAILY_NET_DEPOSITED] = netDeposited
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
			if quantity.IsZero() || stock == CASH || stock == DAILY_NET_DEPOSITED || stock == NORMALIZED_CASH {
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
