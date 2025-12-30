package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/gorilla/websocket"
)

type JsonOutput struct {
	EventType     string     `json:"e"`
	EventTime     int64      `json:"E"`
	Symbol        string     `json:"s"`
	FirstUpdateID int64      `json:"U"`
	FinalUpdateID int64      `json:"u"`
	Bids          [][]string `json:"b"`
	Asks          [][]string `json:"a"`
}

type PriceNode struct {
	Price    float64
	Quantity float64
	Left     *PriceNode
	Right    *PriceNode
}

func (n *PriceNode) Insert(quantity, price float64) *PriceNode {
	if n == nil {
		return &PriceNode{Quantity: quantity, Price: price}
	}
	if price == n.Price {
		n.Quantity = quantity
		return n
	}

	if price > n.Price {
		n.Right = n.Right.Insert(quantity, price)
	} else {
		n.Left = n.Left.Insert(quantity, price)
	}

	return n
}

func (n *PriceNode) Delete(price float64) *PriceNode {
	if n == nil {
		return nil
	}

	if price < n.Price {
		n.Left = n.Left.Delete(price)
	} else if price > n.Price {
		n.Right = n.Right.Delete(price)
	} else {

		if n.Left == nil {
			return n.Right
		} else if n.Right == nil {
			return n.Left
		}

		temp := n.Right.FindMin()

		n.Price = temp.Price
		n.Quantity = temp.Quantity

		n.Right = n.Right.Delete(temp.Price)
	}
	return n
}

func (n *PriceNode) FindMin() *PriceNode {
	if n == nil {
		return nil
	}
	if n.Left == nil {
		return n
	} else {
		return n.Left.FindMin()
	}
}



func GetDescendingTopN(node *PriceNode, results *[]PriceNode, n int) {
	if node == nil || len(*results) >= n {
		return
	}
	GetDescendingTopN(node.Right, results, n)
	if len(*results) < n && node.Quantity > 0 {
		*results = append(*results, *node)
	}
	GetDescendingTopN(node.Left, results, n)
}

func GetAscendingTopN(node *PriceNode, results *[]PriceNode, n int) {
	if node == nil || len(*results) >= n {
		return
	}
	GetAscendingTopN(node.Left, results, n)
	if len(*results) < n && node.Quantity > 0 {
		*results = append(*results, *node)
	}
	GetAscendingTopN(node.Right, results, n)
}

func PrintTopBidsAndAsks(n int) error {
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

				quantity, err := strconv.ParseFloat(quantityStr, 64)

				if err != nil {
					return fmt.Errorf("err %v converting str to float64", err)
				}

				price, err := strconv.ParseFloat(priceStr, 64)

				if err != nil {
					return fmt.Errorf("err %v converting str to float64", err)
				}

				bidsBst = bidsBst.Insert(quantity, price)
				if len(bidsBst) > n {
					minNode := bidsBst.FindMin()
					bidsBst.Delete(minNode.Price)
				}
			}

			for _, ask := range asks {
				priceStr := ask[0]
				quantityStr := ask[1]

				quantity, err := strconv.ParseFloat(quantityStr, 64)
				if err != nil {
					return fmt.Errorf("err %v converting str to float64", err)
				}
				price, err := strconv.ParseFloat(priceStr, 64)

				if err != nil {
					return fmt.Errorf("err %v converting str to float64", err)
				}
				askBst = askBst.Insert(quantity, price)
			}
			// fmt.Printf("price: %s | qty: %s  bid ||| price: %s | qty: %s  ask\n", topBid[0], topBid[1], topAsk[0], topAsk[1])

			var bidResults []PriceNode
			var askResults []PriceNode

			GetDescendingTopN(bidsBst, &bidResults, n)
			GetAscendingTopN(askBst, &askResults, n)
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

func main() {
	PrintTopBidsAndAsks(15)
}
