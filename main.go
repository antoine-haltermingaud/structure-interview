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


func GetAscendingTop10(n *PriceNode, results *[]PriceNode) {
	if n == nil || len(*results) >= 10 {
		return
	}
	GetAscendingTop10(n.Left, results)
	if len(*results) < 10 && n.Quantity > 0 {
		*results = append(*results, *n)
	}
	GetAscendingTop10(n.Right, results)
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
			for _, bid := range bids {
				priceStr := bid[0]
				quantityStr := bid[1]

				quantity, _ := strconv.ParseFloat(quantityStr, 64)
				price, _ := strconv.ParseFloat(priceStr, 64)

				bidsBst = bidsBst.Insert(quantity, price, true)
			}

			for _, ask := range asks {
				priceStr := ask[0]
				quantityStr := ask[1]

				quantity, _ := strconv.ParseFloat(quantityStr, 64)
				price, _ := strconv.ParseFloat(priceStr, 64)

				askBst = askBst.Insert(quantity, price, false)
			}
			// fmt.Printf("price: %s | qty: %s  bid ||| price: %s | qty: %s  ask\n", topBid[0], topBid[1], topAsk[0], topAsk[1])

			var bidResults []PriceNode
			var askResults []PriceNode

			GetDescendingTop10(bidsBst, &bidResults)
			GetAscendingTop10(askBst, &askResults)
			fmt.Println("\nTop bids")
			fmt.Println("----------------------")

			for i, bid := range bidResults {
				fmt.Printf("%v. Price: %v ; Qty: %v\n", i+1, bid.Price, bid.Quantity)
			}
			
			fmt.Println("\nTop asks")
			fmt.Println("----------------------")

			for i, ask := range askResults {
				fmt.Printf("%v. Price: %v ; Qty: %v\n", i+1, ask.Price, ask.Quantity)
			}

		}
	}
}
