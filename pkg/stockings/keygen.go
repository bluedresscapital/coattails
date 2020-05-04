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
	rand.Seed(time.Now().UnixNano()) //need this otherwise it only returns one value lmao
	num := rand.Intn(len(keys))
	return keys[num]
}
