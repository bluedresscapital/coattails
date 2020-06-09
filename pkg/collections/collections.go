package collections

import "github.com/bluedresscapital/coattails/pkg/wardrobe"

func FetchCollectionCountsFromTickers(tickers []string) (map[string]int, error) {
	counts := make(map[string]int)
	for _, t := range tickers {
		collections, err := wardrobe.FetchCollectionsFromTicker(t)
		if err != nil {
			return nil, err
		}
		for _, c := range collections {
			currCount, found := counts[c]
			if found {
				currCount += 1
			} else {
				currCount = 1
			}
			counts[c] = currCount
		}
	}
	return counts, nil
}
