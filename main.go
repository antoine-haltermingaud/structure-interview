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
	Size     int
	Height   int
}

func (n *PriceNode) GetSize() int {
	if n == nil {
		return 0
	}
	return n.Size
}

func (n *PriceNode) getHeight() int {
	if n == nil {
		return 0
	}
	return n.Height
}

func (n *PriceNode) getBalance() int {
	if n == nil {
		return 0
	}
	return n.Left.getHeight() - n.Right.getHeight()
}

func (n *PriceNode) update() {
	n.Size = 1 + n.Left.GetSize() + n.Right.GetSize()
	lh := n.Left.getHeight()
	rh := n.Right.getHeight()
	if lh > rh {
		n.Height = lh + 1
	} else {
		n.Height = rh + 1
	}
}

func (y *PriceNode) rightRotate() *PriceNode {
	x := y.Left
	T2 := x.Right

	x.Right = y
	y.Left = T2
	y.update()
	x.update()
	return x
}

func (x *PriceNode) leftRotate() *PriceNode {
	y := x.Right
	T2 := y.Left

	y.Left = x
	x.Right = T2
	x.update()
	y.update()

	return y
}

func (n *PriceNode) Insert(quantity, price float64) *PriceNode {
	if n == nil {
		return &PriceNode{Quantity: quantity, Price: price, Size: 1, Height: 1}
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

	n.update()

	balance := n.getBalance()

	if balance > 1 && price < n.Left.Price {
		return n.rightRotate()
	}

	if balance < -1 && price > n.Right.Price {
		return n.leftRotate()
	}

	if balance > 1 && price > n.Left.Price {
		n.Left = n.Left.leftRotate()
		return n.rightRotate()
	}

	if balance < -1 && price < n.Right.Price {
		n.Right = n.Right.rightRotate()
		return n.leftRotate()
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

	n.update()

	balance := n.getBalance()

	if balance > 1 && n.Left.getBalance() >= 0 {
		return n.rightRotate()
	}
	if balance > 1 && n.Left.getBalance() < 0 {
		n.Left = n.Left.leftRotate()
		return n.rightRotate()
	}

	if balance < -1 && n.Right.getBalance() <= 0 {
		return n.leftRotate()
	}
	if balance < -1 && n.Right.getBalance() > 0 {
		n.Right = n.Right.rightRotate()
		return n.leftRotate()
	}

	return n
}

func (n *PriceNode) FindMin() *PriceNode {
	if n == nil {
		return nil
	}
	if n.Left == nil {
		return n
	}
	return n.Left.FindMin()
}

func (n *PriceNode) FindMax() *PriceNode {
	if n == nil {
		return nil
	}
	if n.Right == nil {
		return n
	}
	return n.Right.FindMax()
}

func (n *PriceNode) DeleteMin() *PriceNode {
	if n == nil {
		return nil
	}
	minNode := n.FindMin()
	return n.Delete(minNode.Price)
}

func (n *PriceNode) DeleteMax() *PriceNode {
	if n == nil {
		return nil
	}
	maxNode := n.FindMax()
	return n.Delete(maxNode.Price)
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
	var asksBst *PriceNode

	for {
		_, message, _ := c.ReadMessage()

		var output JsonOutput
		if err := json.Unmarshal(message, &output); err != nil {
			continue
		}
		bids := output.Bids
		asks := output.Asks

		if len(bids) > 0 {
			for _, bid := range bids {
				priceStr := bid[0]
				quantityStr := bid[1]

				quantity, _ := strconv.ParseFloat(quantityStr, 64)
				price, _ := strconv.ParseFloat(priceStr, 64)

				if quantity == 0 {
					bidsBst = bidsBst.Delete(price)
				} else {
					bidsBst = bidsBst.Insert(quantity, price)
				}
			}
			for bidsBst.GetSize() > 10*n {
				bidsBst = bidsBst.DeleteMin()
			}
		}

		if len(asks) > 0 {
			for _, ask := range asks {
				priceStr := ask[0]
				quantityStr := ask[1]

				quantity, _ := strconv.ParseFloat(quantityStr, 64)
				price, _ := strconv.ParseFloat(priceStr, 64)

				if quantity == 0 {
					asksBst = asksBst.Delete(price)
				} else {
					asksBst = asksBst.Insert(quantity, price)
				}
			}
			for asksBst.GetSize() > 10*n {
				asksBst = asksBst.DeleteMax()
			}
		}
		// fmt.Printf("price: %s | qty: %s  bid ||| price: %s | qty: %s  ask\n", topBid[0], topBid[1], topAsk[0], topAsk[1])

		var bidResults []PriceNode
		var askResults []PriceNode

		GetDescendingTopN(bidsBst, &bidResults, n)
		GetAscendingTopN(asksBst, &askResults, n)
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

func main() {
	PrintTopBidsAndAsks(15)
}
