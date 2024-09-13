// TODO: Add package documentation for `main`, like this:
// Package main something something...
package main

import (
	"bufio"
	"d7024e/kademlia"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	fmt.Println("Pretending to run the kademlia app...")
	k := JoinNetwork()

	go kademlia.Listen(k)

	userInputHandler(k)
	// Keep the main function running to prevent container exit
	select {}
}

// Function to handle user input
func userInputHandler(k *kademlia.Kademlia) {
	// Create a new reader to read from standard input (os.Stdin)

	for {
		consoleReader := bufio.NewReader(os.Stdin)
		fmt.Print(">")

		input, _ := consoleReader.ReadString('\n')

		// Trim any leading/trailing whitespace or newline characters
		input = strings.TrimSpace(input)

		// Split the input into command and argument
		parts := strings.SplitN(input, " ", 2)
		command := parts[0]
		var arg string
		if len(parts) > 1 {
			arg = parts[1]
		}

		// Print the input for confirmation
		fmt.Printf("You entered: command=%s, argument=%s\n", command, arg)

		// Switch statement to handle different commands
		// TODO : change to CLI
		switch strings.ToUpper(command) {
		case "PING":
			if arg != "" {
				// Create a new contact with a random Kademlia ID and the argument as the address
				// TODO this is not how a ping should work since a user should not ping
				contact := kademlia.NewContact(kademlia.NewRandomKademliaID(), strings.TrimSpace(arg))
				// Send a ping message
				if k.Network.SendPingMessage(&k.RoutingTable.Me, &contact) {
					k.HandlePingOrPong(contact.ID.String(), contact.Address)
				}
				k.RoutingTable.PrintRoutingTable()
			} else {
				fmt.Println("Error: No argument provided for PING.")
			}

		case "GET":
			if arg != "" {
				fmt.Printf("GET command not implemented for: %s\n", arg)
				// TODO: Implement GET command logic
			} else {
				fmt.Println("Error: No argument provided for GET.")
			}

		case "PUT":
			if arg != "" {
				fmt.Printf("PUT command not implemented for: %s\n", arg)
				// TODO: Implement PUT command logic
			} else {
				fmt.Println("Error: No argument provided for PUT.")
			}

		case "EXIT":
			fmt.Println("Exiting program.")
			return // Exit the program

		default:
			fmt.Println("Error: Unknown command.")
		}

	}
}

func JoinNetwork() *kademlia.Kademlia {
	//Preparing new contact for self with own IP
	id := kademlia.NewRandomKademliaID()
	contact := kademlia.NewContact(id, GetOutboundIP().String())
	contact.CalcDistance(id)
	fmt.Println(contact.String())
	fmt.Printf("%v\n", contact)

	//Creating new routing table with self as contact
	routingTable := kademlia.NewRoutingTable(contact)

	//Adding bootstrap contact
	bootStrapContact := kademlia.NewContact(kademlia.NewKademliaID("FFFFFFFFF0000000000000000000000000000000)"), "172.20.0.6")
	routingTable.AddContact(bootStrapContact)

	//Creating new network for self
	network := &kademlia.Network{}

	//Creating new kademlia instance with own routing table and network
	kademliaInstance := &kademlia.Kademlia{RoutingTable: routingTable, Network: network}

	//Lookup on self to update routing table
	kademliaInstance.LookupContact(&contact)

	return kademliaInstance
}

// Get preferred outbound ip of this machine
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
