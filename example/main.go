package main

import (
	"log"

	"github.com/fleegrid/tun"
)

func main() {
	ifce, err := tun.NewDevice()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Interface Name: [%s]\n", ifce.Name())

	packet := make([]byte, 2000)
	for {
		n, err := ifce.Read(packet)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Packet Received: % x\n", packet[:n])
	}
}
