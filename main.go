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
	var bidsBst *PriceNode
	var askBst *PriceNode

	for {
		_, message, _ := c.ReadMessage()

		var output JsonOutput
		json.Unmarshal(message, &output)
		bids := output.Bids
		asks := output.Asks

		if len(bids) > 0 && len(asks) > 0 {
			topBid := bids[0]
			topAsk := asks[0]
			// fmt.Printf("price: %s | qty: %s  bid ||| price: %s | qty: %s  ask\n", topBid[0], topBid[1], topAsk[0], topAsk[1])
			bidPriceStr := topBid[0]
			bidQuantityStr := topBid[1]

			askPriceStr := topAsk[0]
			askQuantityStr := topAsk[1]

			bidQuantity, _ := strconv.ParseFloat(bidQuantityStr, 64)
			bidPrice, _ := strconv.ParseFloat(bidPriceStr, 64)

			askQuantity, _ := strconv.ParseFloat(askQuantityStr, 64)
			askPrice, _ := strconv.ParseFloat(askPriceStr, 64)

			bidsBst = bidsBst.Insert(bidQuantity, bidPrice, true)
			askBst = bidsBst.Insert(askQuantity, askPrice, true)

			var bidResults []PriceNode
			var askResults []PriceNode

			GetDescendingTop10(bidsBst, &bidResults)
			GetDescendingTop10(askBst, &askResults)
			fmt.Printf("Top 10 Bids: %+v\n", bidResults)
			fmt.Printf("Top 10 Asks: %+v\n", askResults)

		}
	}


}
