package main

import (
	"fmt"
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"log"
	"strings"
	"time"
)

func main() {
	wardrobe.InitCache("localhost")
	go func() {
		subRes1 := wardrobe.Sub("wei")
		for {
			select {
			case msg := <-subRes1.Channel():
				log.Printf("pubsub1 received msg: %s", msg.String())
			}
		}
	}()
	go func() {
		subRes2 := wardrobe.Sub("wei")
		for {
			select {
			case msg := <-subRes2.Channel():
				log.Printf("pubsub2 received msg: %s", msg.String())
				if strings.Contains(msg.String(), "Test 5") {
					log.Printf("pubsub2 unsubscribing")
					_ = wardrobe.Unsub(subRes2, "wei")
				}
			}
		}
	}()

	for i := 0; i < 10; i++ {
		err := wardrobe.Publish("wei", []byte(fmt.Sprintf("Test %d", i)))
		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(1 * time.Second)
	}
}
