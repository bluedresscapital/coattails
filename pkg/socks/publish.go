package socks

import (
	"encoding/json"
	"log"

	"github.com/bluedresscapital/coattails/pkg/wardrobe"
	"github.com/google/uuid"
)

type ReduxEvent struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

func PublishFromServer(channel string, reduxType string, payload interface{}) error {
	payloadB, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	event := ReduxEvent{
		Type:    reduxType,
		Payload: string(payloadB),
	}
	eventB, err := json.Marshal(event)
	if err != nil {
		return err
	}
	msg := Msg{
		ClientUid: uuid.New().String(),
		Payload:   string(eventB),
	}
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	err = wardrobe.Publish(channel, msgBytes)
	if err != nil {
		log.Printf("Errored in publishing: %v", err)
		return err
	}
	return nil
}
