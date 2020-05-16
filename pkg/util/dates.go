package util

import "time"

func IsMarketOpen(date time.Time) bool {
	return int(date.Weekday()) != 0 && int(date.Weekday()) != 6
}

func GetTimelessDate(date time.Time) time.Time {
	y, m, d := date.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// Simple function to just get market dates. Ignores the fact that markets are closed on holidays for simplicity.
func GetMarketDates(start time.Time, end time.Time) []time.Time {
	start = GetTimelessDate(start)
	end = GetTimelessDate(end)
	var dates []time.Time
	for currDate := start; currDate.Before(end.AddDate(0, 0, 1)); currDate = currDate.AddDate(0, 0, 1) {
		if IsMarketOpen(currDate) {
			dates = append(dates, currDate)
		}
	}
	return dates
}
