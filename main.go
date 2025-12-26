package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/gorilla/websocket"
)



type JsonOutput struct {
	e string 	`json:"e"`
	E int64 	`json:"E"`
	s string 	`json:"s"`
	U int64 	`json:"U"`
	u int64 	`json:"u"`
	Bids [][]string		`json:"b"`
	Asks [][]string		`json:"a"`
}

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

		
		var output JsonOutput
		json.Unmarshal(message, &output)
		bids := output.Bids
		asks := output.Asks


		if len(bids) > 0 && len(asks) > 0 {
			topBid := bids[0]
			//topAsk := asks[0].([]any)
			// fmt.Printf("price: %s | qty: %s  bid ||| price: %s | qty: %s  ask\n", topBid[0], topBid[1], topAsk[0], topAsk[1])
			quantityStr := topBid[0]
			bidPriceStr := topBid[0]

			quantity, _ := strconv.ParseFloat(quantityStr, 64)
			bidPrice, _ := strconv.ParseFloat(bidPriceStr, 64)

			bidsBst.Insert(quantity, bidPrice, true)

			fmt.Println(GetDescendingTop10(PriceNode))
		}
	}


}
