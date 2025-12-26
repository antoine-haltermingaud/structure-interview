package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type PriceNode struct {
	Price 		float64
	Quantity 	float64
	Left 			*PriceNode
	Right			*PriceNode
}

func (n *PriceNode) Insert(quantity, price float64, isBid bool) *PriceNode{
	if n == nil { //pointer doesn't exist -> returns new one
		return &PriceNode{Quantity: quantity, Price: price}
	}
	if price == n.Price { 
		n.Quantity = quantity
		return n
	}
	if price > n.Price {
		n.Right = n.Right.Insert(quantity, price, isBid)
	} else {
		n.Left = n.Left.Insert(quantity, price, isBid)
	}
	return n 

}

func GetDescendingTop10(n *PriceNode, results *[]PriceNode) {
	if n == nil || len(*results) >= 10 {
		return
	}
	GetDescendingTop10(n.Right, results)
	if len(*results) < 10 && n.Quantity > 0 {
		*results = append(*results, *n)
	}
	GetDescendingTop10(n.Left, results)
}


func main() {
	url := "wss://stream.binance.com:9443/ws/btcusdc@depth"

	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal(err)
	}

	defer c.Close()

	fmt.Println("Reading stream")
	bidsBst := &PriceNode{}

	for {
		_, message, _ := c.ReadMessage()

		var raw map[string]any
		json.Unmarshal(message, &raw)
		bids, _ := raw["b"].([]any)
		asks, _ := raw["a"].([]any)

		if len(bids) > 0 && len(asks) > 0 {
			topBid := bids[0].([]float64)
			//topAsk := asks[0].([]any)
			// fmt.Printf("price: %s | qty: %s  bid ||| price: %s | qty: %s  ask\n", topBid[0], topBid[1], topAsk[0], topAsk[1])
			quantity := topBid[1]
			bidPrice := topBid[0]
			bidsBst.Insert(quantity, bidPrice, true)
		}
	}
}
