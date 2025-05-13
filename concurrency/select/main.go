package main

import (
	"log"
	"time"
)

type Message struct {
	OrderID string
	Title   string
	Price   int
}

func OrderProduct(buyChan chan<- Message, orders []Message) {
	for _, order := range orders {
		log.Printf("[ORDER_TASK_PROCESSING]: %+v", order)
		time.Sleep(1 * time.Second)
		buyChan <- order
	}
	close(buyChan)
}

func CancelOrder(cancelChan chan<- string, cancelOrders []string) {
	for _, cancelOrder := range cancelOrders {
		log.Printf("[CANCEL_ORDER_TASK_PROCESSING]: %+v", cancelOrders)
		time.Sleep(5 * time.Second)
		cancelChan <- cancelOrder
	}
	close(cancelChan)
}

func HandlerOrder(buyChan <-chan Message, cancelChan <-chan string) {
	for {
		select {
		case buy, ok := <-buyChan:
			{
				if ok {
					log.Printf("[HANDLER_ORDER]: %+v", buy)
				} else {
					buyChan = nil
				}
			}
		case cancel, ok := <-cancelChan:
			{
				if ok {
					log.Printf("[HANDLER_CANCEL_ORDER]: %+v", cancel)
				} else {
					cancelChan = nil
				}
			}
		}
		if buyChan == nil && cancelChan == nil {
			break
		}
	}
}

func main() {
	buyChan := make(chan Message)
	cancelChan := make(chan string)

	// mock concurency order requests
	orders := []Message{
		{OrderID: "order-001", Title: "Book", Price: 100},
		{OrderID: "order-002", Title: "Toy", Price: 200},
		{OrderID: "order-003", Title: "Table", Price: 300},
		{OrderID: "order-004", Title: "Clothe", Price: 400},
		{OrderID: "order-005", Title: "Glasses", Price: 500},
	}

	// mock concurency cancel order requests
	cancelOrders := []string{"order-002", "order-003", "order-005"}
	go OrderProduct(buyChan, orders)
	go CancelOrder(cancelChan, cancelOrders)
	go HandlerOrder(buyChan, cancelChan)

	time.Sleep(12 * time.Second)
}
