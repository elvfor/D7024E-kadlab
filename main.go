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
	"sync"
)

func main() {
	fmt.Println("Pretending to run the kademlia app...")
	ip := GetOutboundIP().String()
	if ip == "172.20.0.6" {
		k := JoinNetworkBootstrap(ip)
		go kademlia.Listen(k)
		go userInputHandler(k)
	} else {
		k := JoinNetwork(GetOutboundIP().String() + ":8000")
		go kademlia.Listen(k)
		go DoLookUpOnSelf(k)
		go userInputHandler(k)
	}

	// Keep the main function running to prevent container exit
	select {}
}

// Function to handle user input
func userInputHandler(k *kademlia.Kademlia) {
	consoleReader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(">")

		input, err := consoleReader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}

		input = strings.TrimSpace(input)
		parts := strings.SplitN(input, " ", 2)
		command := parts[0]
		var arg string
		if len(parts) > 1 {
			arg = parts[1]
		}

		fmt.Printf("You entered: command=%s, argument=%s\n", command, arg)

		switch strings.ToUpper(command) {
		case "PING":
			handlePing(k, arg)
		case "GET":
			handleGet(k, arg)
		case "PUT":
			handlePut(k, arg)
		case "EXIT":
			fmt.Println("Exiting program.")
			return
		case "LOOKUP":
			handleLookup(k, arg)
		case "PRINT":
			k.RoutingTable.PrintAllIP()
		default:
			fmt.Println("Error: Unknown command.")
		}
	}
}

func handlePing(k *kademlia.Kademlia, arg string) {
	if arg != "" {
		contact := kademlia.NewContact(kademlia.NewRandomKademliaID(), strings.TrimSpace(arg))
		if k.Network.SendPingMessage(&k.RoutingTable.Me, &contact) {
			k.UpdateRT(contact.ID.String(), contact.Address)
		}
	} else {
		fmt.Println("Error: No argument provided for PING.")
	}
}

func handleGet(k *kademlia.Kademlia, arg string) {
	if arg != "" {
		targetContact := kademlia.NewContact(kademlia.NewKademliaID(arg), "")
		contacts := k.NodeLookup(&targetContact)
		for _, contact := range contacts {
			go func(contact kademlia.Contact) {
				_, data, _ := k.Network.SendFindDataMessage(&k.RoutingTable.Me, &contact, arg)
				if data != nil {
					fmt.Println("Data:", string(data), "found on contact:", contact.String())
					return
				}
			}(contact)
		}
	} else {
		fmt.Println("Error: No argument provided for GET.")
	}
}

func handlePut(k *kademlia.Kademlia, arg string) {
	if arg != "" {
		data := []byte(arg)
		randomKademliaID := kademlia.NewRandomKademliaID()
		targetContact := kademlia.NewContact(randomKademliaID, "")
		contacts := k.NodeLookup(&targetContact)
		resultChan := make(chan bool, len(contacts))
		var wg sync.WaitGroup

		for _, contact := range contacts {
			wg.Add(1)
			go func(contact kademlia.Contact) {
				defer wg.Done()
				result := k.Network.SendStoreMessage(&k.RoutingTable.Me, &contact, randomKademliaID, data)
				fmt.Println("Storing data with key:", randomKademliaID.String(), "on contact:", contact.String())
				resultChan <- result
			}(contact)
		}

		go func() {
			wg.Wait()
			close(resultChan)
		}()

		successCount := 0
		for success := range resultChan {
			if success {
				successCount++
			}
		}
		if successCount > len(contacts)/2 {
			fmt.Println("Data stored successfully. Hash:" + randomKademliaID.String())
		} else {
			fmt.Println("Failed to store data.")
		}
	} else {
		fmt.Println("Error: No argument provided for PUT.")
	}
}

func handleLookup(k *kademlia.Kademlia, arg string) {
	if arg != "" {
		contact := kademlia.NewContact(kademlia.NewRandomKademliaID(), strings.TrimSpace(arg))
		bootStrapContact := kademlia.NewContact(kademlia.NewKademliaID("FFFFFFFFF0000000000000000000000000000000)"), "172.20.0.6")
		contacts, _ := k.Network.SendFindContactMessage(&contact, &bootStrapContact, &contact)
		fmt.Print(contacts)
	}
}

func JoinNetwork(ip string) *kademlia.Kademlia {
	id := kademlia.NewRandomKademliaID()
	contact := kademlia.NewContact(id, ip)
	contact.CalcDistance(id)
	routingTable := kademlia.NewRoutingTable(contact)
	bootStrapContact := kademlia.NewContact(kademlia.NewKademliaID("FFFFFFFFF0000000000000000000000000000000)"), "172.20.0.6:8000")
	routingTable.AddContact(bootStrapContact)
	network := &kademlia.Network{}
	data := make(map[string][]byte)
	return &kademlia.Kademlia{RoutingTable: routingTable, Network: network, Data: &data}
}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}

func DoLookUpOnSelf(k *kademlia.Kademlia) {
	fmt.Println("Doing lookup on self")
	bootStrapContact := kademlia.NewContact(kademlia.NewKademliaID("FFFFFFFFF0000000000000000000000000000000)"), "172.20.0.6:8000")
	kClosest, _ := k.Network.SendFindContactMessage(&k.RoutingTable.Me, &bootStrapContact, &k.RoutingTable.Me)
	fmt.Println("Length of kClosest: ", len(kClosest))
	for _, contact := range kClosest {
		k.UpdateRT(contact.ID.String(), contact.Address)
	}
}

func JoinNetworkBootstrap(ip string) *kademlia.Kademlia {
	bootStrapContact := kademlia.NewContact(kademlia.NewKademliaID("FFFFFFFFF0000000000000000000000000000000)"), "172.20.0.6:8000")
	bootStrapContact.CalcDistance(bootStrapContact.ID)
	routingTable := kademlia.NewRoutingTable(bootStrapContact)
	network := &kademlia.Network{}
	data := make(map[string][]byte)
	return &kademlia.Kademlia{RoutingTable: routingTable, Network: network, Data: &data}
}
