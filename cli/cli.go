package cli

import (
	"bufio"
	"crypto/sha1"
	"d7024e/kademlia"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"sync"
)

func UserInputHandler(k *kademlia.Kademlia, reader io.Reader, writer io.Writer) {
	consoleReader := bufio.NewReader(reader)
	for {
		fmt.Fprint(writer, ">")

		input, err := consoleReader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(writer, "Error reading input:", err)
			continue
		}

		input = strings.TrimSpace(input)
		parts := strings.SplitN(input, " ", 2)
		command := parts[0]
		var arg string
		if len(parts) > 1 {
			arg = parts[1]
		}

		fmt.Fprintf(writer, "You entered: command=%s, argument=%s\n", command, arg)

		switch strings.ToUpper(command) {
		case "PING":
			handlePing(k, arg)
		case "GET":
			handleGet(k, arg)
		case "PUT":
			handlePut(k, arg)
		case "EXIT":
			fmt.Fprintln(writer, "Exiting program.")
			return
		case "LOOKUP":
			handleLookup(k, arg)
		case "PRINT":
			k.ActionChannel <- kademlia.Action{Action: "PRINT"}
		default:
			fmt.Fprintln(writer, "Error: Unknown command.")
		}
	}
}

func handlePing(k *kademlia.Kademlia, arg string) {
	if arg != "" {
		contact := kademlia.NewContact(kademlia.NewRandomKademliaID(), strings.TrimSpace(arg))
		if k.Network.SendPingMessage(&k.RoutingTable.Me, &contact) {
			k.UpdateRT(contact.ID, contact.Address)
		}
	} else {
		fmt.Println("Error: No argument provided for PING.")
	}
}

func handleGet(k *kademlia.Kademlia, arg string) {
	if arg == "" {
		fmt.Println("Error: No argument provided for GET.")
		return
	}

	if len(arg) != 40 { // Kademlia ID length
		fmt.Println("Error: Invalid Kademlia ID length.")
		return
	}

	targetContact := kademlia.NewContact(kademlia.NewKademliaID(arg), "")
	_, foundOncontact, foundData := k.NodeLookup(&targetContact, arg)
	if foundData != nil {
		fmt.Println("Data found on contact:", foundOncontact.String())
		fmt.Println("Data:", string(foundData))
	} else {
		fmt.Println("Data not found.")
	}
}

func handlePut(k *kademlia.Kademlia, arg string) {
	if arg != "" {
		data := []byte(arg)
		hasher := sha1.New()
		hasher.Write([]byte("hash1"))
		hash := hasher.Sum(nil)
		hashString := hex.EncodeToString(hash)
		kadId := kademlia.NewKademliaID(hashString)
		targetContact := kademlia.NewContact(kadId, "")
		contacts, _, _ := k.NodeLookup(&targetContact, "")
		resultChan := make(chan bool, len(contacts))
		var wg sync.WaitGroup

		for _, contact := range contacts {
			wg.Add(1)
			go func(contact kademlia.Contact) {
				defer wg.Done()
				result := k.Network.SendStoreMessage(&k.RoutingTable.Me, &contact, kadId, data)
				fmt.Println("Storing data with key:", kadId.String(), "on contact:", contact.String())
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
			fmt.Println("Data stored successfully. Hash:" + kadId.String())
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
