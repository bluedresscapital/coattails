package routes

import (
	"fmt"
	"github.com/bluedresscapital/coattails/pkg/socks"
	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func getChannelFromUserId(userId int) string {
	return fmt.Sprintf("chanel_user_id_%d", userId)
}

func testWebSocket(userId *int, w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("err upgrading: %v", err)
		return
	}
	channel := getChannelFromUserId(*userId)
	log.Printf("Creating web socket connection for channel %s!", channel)
	sub := wardrobe.Sub(channel)
	client := socks.Client{
		Uid:     uuid.New().String(),
		Channel: channel,
		Conn:    conn,
		Sub:     sub,
	}
	go client.ReadFromClient()
	go client.WriteToClient()
}
