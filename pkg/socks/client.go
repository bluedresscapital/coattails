package socks

import (
	"encoding/json"
	"log"
	"time"

	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"github.com/go-redis/redis/v7"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type Client struct {
	Uid     string
	Channel string
	Conn    *websocket.Conn
	Sub     *redis.PubSub
}

type Msg struct {
	ClientUid string
	Payload   string
}

// read reads messages from client and writes to redis
// channel
func (c *Client) ReadFromClient() {
	defer c.shutdown()
	c.Conn.SetReadLimit(maxMessageSize)
	_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { _ = c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			return
		}
		redisMsg := Msg{
			ClientUid: c.Uid,
			Payload:   string(msg),
		}
		b, err := json.Marshal(redisMsg)
		if err != nil {
			return
		}
		_ = wardrobe.Publish(c.Channel, b)
	}
}

// write reads from redis channel and writes to client
func (c *Client) WriteToClient() {
	ticker := time.NewTicker(pingPeriod)
	defer c.shutdown()
	for {
		select {
		case msg := <-c.Sub.Channel():
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			var redisMsg Msg
			err := json.Unmarshal([]byte(msg.Payload), &redisMsg)
			if err != nil {
				log.Printf("Error in unmarshalling redisMsg: %v", err)
				return
			}
			if redisMsg.ClientUid == c.Uid {
				continue
			}
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Printf("Error in getting next writer: %v", err)
				return
			}
			// Write to client, ignore all errors :D
			_, _ = w.Write([]byte(redisMsg.Payload))
			if err := w.Close(); err != nil {
				log.Printf("Error in writing to client: %v", err)
				return
			}
		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Error in writing ping message: %v", err)
				return
			}
		}
	}
}

func (c *Client) shutdown() {
	// Unsubscribe from redis channel
	_ = wardrobe.Unsub(c.Sub, c.Channel)
	// Close client websocket connection
	_ = c.Conn.Close()
	log.Printf("Shutting down connection for channel %s", c.Channel)
}
