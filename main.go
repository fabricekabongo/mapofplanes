package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello, World!")
	client := NewADSBClient("adsb.h.fabricekabongo.com", "30003")
	//client := TCPReader.NewADSBClient("data.adsbhub.org", "5002")
	err := client.Connect()
	if err != nil {
		panic(err)
	}

	client.StartListening(10)
}
