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

	network := &kademlia.Network{}
	go kademlia.Listen("0.0.0.0", 8000)

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
				contact := kademlia.NewContact(kademlia.NewRandomKademliaID(), strings.TrimSpace(arg))
				// Send a ping message
				network.SendPingMessage(&contact)
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
