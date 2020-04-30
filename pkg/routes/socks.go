package routes

import (
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

func testWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	c, statusCode, err := fetchCookie(r)
	if err != nil {
		w.WriteHeader(statusCode)
		return
	}
	username, err := wardrobe.FetchAuthToken(c.Value)
	if err != nil || username == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	sub := wardrobe.Sub(*username)
	client := socks.Client{
		Uid:     uuid.New().String(),
		Channel: *username,
		Conn:    conn,
		Sub:     sub,
	}
	go client.ReadFromClient()
	go client.WriteToClient()
}
