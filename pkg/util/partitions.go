package util

func PartitionTickers(tickers []string, buckets int) [][]string {
	ret := make([][]string, buckets)
	for i := 0; i < buckets; i++ {
		ret[i] = make([]string, 0)
	}
	counter := 0
	for _, t := range tickers {
		ret[counter%buckets] = append(ret[counter%buckets], t)
		counter++
	}
	return ret
}

func PartitionPorts(ports []int, buckets int) [][]int {
	ret := make([][]int, buckets)
	for i := 0; i < buckets; i++ {
		ret[i] = make([]int, 0)
	}
	counter := 0
	for _, id := range ports {
		ret[counter%buckets] = append(ret[counter%buckets], id)
		counter++
	}
	return ret
}
