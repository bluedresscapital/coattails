package util

import (
	"time"
)

func GetESTNow() time.Time {
	est, _ := time.LoadLocation("EST")
	return time.Now().In(est)
}

// Basically returns the "effective" date - the last date in which the stock market was open?
func GetTimelessESTOpenNow() time.Time {
	estNow := GetESTNow()
	if IsMarketOpen(estNow) {
		return GetTimelessDate(estNow)
	}
	return GetTimelessDate(estNow.AddDate(0, 0, -1))
}

// Simple check to see if market is open (ignoring holidays)
func IsMarketOpen(date time.Time) bool {
	// Must be a weekday (i.e. monday - friday)
	return int(date.Weekday()) >= 1 && int(date.Weekday()) <= 5 &&
		// Hour must either be > 9 OR if its equal to 9, minute must be >= 30
		(date.Hour() > 9 || (date.Hour() == 9 && date.Minute() >= 30)) &&
		// Hour must be < 4:00pm
		date.Hour() < 16
}

func GetTimelessDate(date time.Time) time.Time {
	y, m, d := date.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// Simple function to just get market dates. Ignores the fact that markets are closed on holidays for simplicity.
func GetMarketDates(start time.Time, end time.Time) []time.Time {
	start = GetTimelessDate(start)
	end = GetTimelessDate(end)
	// Not sure what we should do if start is a weekend...
	//if int(start.Weekday()) == 0 {
	//	start = start.AddDate(0, 0, -2)
	//} else if int(start.Weekday()) == 6 {
	//	start = start.AddDate(0, 0, -1)
	//}
	var dates []time.Time
	for currDate := start; currDate.Before(end.AddDate(0, 0, 1)); currDate = currDate.AddDate(0, 0, 1) {
		dates = append(dates, currDate)
	}
	return dates
}
