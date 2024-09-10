// TODO: Add package documentation for `main`, like this:
// Package main something something...
package main

import (
	"bufio"
	"d7024e/kademlia"
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Println("Pretending to run the kademlia app...")
	// Using stuff from the kademlia package here. Something like...
	id := kademlia.NewRandomKademliaID()
	contact := kademlia.NewContact(id, "localhost:8000")
	fmt.Println(contact.String())
	fmt.Printf("%v\n", contact)
	//routingTable := kademlia.NewRoutingTable(contact)
	//fmt.Println(routingTable)

	//start go routine to call for kademlia.listen
	// Start the kademlia.Listen function as a goroutine
	network := &kademlia.Network{}
	go kademlia.Listen("localhost", 8000)

	userInputHandler(network)
	// Keep the main function running to prevent container exit
	select {}
}

// Function to handle user input
func userInputHandler(network *kademlia.Network) {
	// Create a new reader to read from standard input (os.Stdin)

	for {
		consoleReader := bufio.NewReader(os.Stdin)
		fmt.Print(">")

		input, _ := consoleReader.ReadString('\n')

		//print the input
		fmt.Println("you typed: " + input)
		//ip 172.10.2
		if strings.HasPrefix(input, "ip") {
			contact := kademlia.NewContact(kademlia.NewRandomKademliaID(), strings.TrimSpace(input[3:]))
			network.SendPingMessage(&contact)
		}
	}
}
