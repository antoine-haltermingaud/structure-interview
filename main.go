package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

func main() {
	url := "wss://stream.binance.com:9443/ws/btcusdc@depth"

	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal(err)
	}

	defer c.Close()

	fmt.Println("Reading stream")

	for {
		_, message, _ := c.ReadMessage()

		var raw map[string]any
		json.Unmarshal(message, &raw)
		bids, _ := raw["b"].([]any)
		asks, _ := raw["a"].([]any)

		if len(bids) > 0 && len(asks) > 0 {
			topBid := bids[0].([]any)
			topAsk := asks[0].([]any)
			fmt.Printf("price: %s | qty: %s  bid\n", topBid[0], topBid[1])
			fmt.Printf("price: %s | qty: %s  ask\n", topAsk[0], topAsk[1])
		}
	}
}
