package main

import "fmt"

// Client is the client that uses the adapter
type Client struct{}

// InsertLightningConnectorIntoComputer inserts a Lightning connector into a computer
func (c *Client) InsertLightningConnectorIntoComputer(com Computer) {
	fmt.Println("Client inserts Lightning connector into computer.")
	com.InsertIntoLightningPort()
}
