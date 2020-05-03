package stockings

import (
	"math/rand"
	"os"
	"strings"
	"time"
)

var keys []string

func InitKeygen() {
	keys = strings.Split(os.Getenv("IEX_TOKEN"), ",")
}

func getKey() string {
	//need this otherwise it only returns one value lmao
	rand.Seed(time.Now().UnixNano())
	num := rand.Intn(len(keys))
	return keys[num]
}
