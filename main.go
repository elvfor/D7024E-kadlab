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
	go kademlia.Listen("localhost", 8000)

	go userInputHandler()
	// Keep the main function running to prevent container exit
	select {}
}

// Function to handle user input
func userInputHandler() {
	// Create a new reader to read from standard input (os.Stdin)
	reader := bufio.NewReader(os.Stdin)

	for {
		// Print a prompt message
		fmt.Print("Enter command: ")

		// Read input from standard input until a newline is encountered
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}

		// Trim any leading or trailing whitespace from the input
		input = strings.TrimSpace(input)

		// Check if the input command is "exit"
		if input == "exit" {
			fmt.Println("Exiting...")
			os.Exit(0) // Exit the program
		}

		// Print the input command
		fmt.Println("You entered:", input)
	}
}
