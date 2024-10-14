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

// ReadUserInput reads the input from the reader, trims and parses the command and argument
func ReadUserInput(reader io.Reader, writer io.Writer) (string, string, error) {
	consoleReader := bufio.NewReader(reader)
	fmt.Fprint(writer, ">")
	input, err := consoleReader.ReadString('\n')
	if err != nil {
		return "", "", fmt.Errorf("error reading input: %w", err)
	}

	input = strings.TrimSpace(input)
	parts := strings.SplitN(input, " ", 2)
	command := parts[0]
	var arg string
	if len(parts) > 1 {
		arg = parts[1]
	}

	return strings.ToUpper(command), arg, nil
}

func UserInputHandler(k *kademlia.Kademlia, reader io.Reader, writer io.Writer) {
	for {
		command, arg, err := ReadUserInput(reader, writer)
		if err != nil {
			fmt.Fprintln(writer, err)
			continue
		}

		if handleCommand(k, command, arg, writer) {
			return
		}
	}
}

func handleCommand(k *kademlia.Kademlia, command, arg string, writer io.Writer) bool {
	fmt.Fprintf(writer, "You entered: command=%s, argument=%s\n", command, arg)

	switch command {
	case "GET":
		handleGet(k, arg)
	case "PUT":
		handlePut(k, arg)
	case "EXIT":
		fmt.Fprintln(writer, "Exiting program.")
		return true
	case "PRINT":
		k.ActionChannel <- kademlia.Action{Action: "PRINT"}
	default:
		fmt.Fprintln(writer, "Error: Unknown command.")
	}
	return false
}

func handleGet(k *kademlia.Kademlia, arg string) {
	if err := ValidateGetArg(arg); err != nil {
		fmt.Println(err)
		return
	}

	targetContact := CreateTargetContact(arg)
	foundOnContact, foundData := performNodeLookup(k, targetContact, arg)
	HandleLookupResult(foundOnContact, foundData)
}

func ValidateGetArg(arg string) error {
	if arg == "" {
		return fmt.Errorf("error: No argument provided for GET")
	}

	if len(arg) != 40 { // Kademlia ID length
		return fmt.Errorf("error: Invalid Kademlia ID length")
	}

	return nil
}

func CreateTargetContact(arg string) kademlia.Contact {
	return kademlia.NewContact(kademlia.NewKademliaID(arg), "")
}

func performNodeLookup(k *kademlia.Kademlia, targetContact kademlia.Contact, arg string) (kademlia.Contact, []byte) {
	_, foundOnContact, foundData := k.NodeLookup(&targetContact, arg)
	return foundOnContact, foundData
}

func HandleLookupResult(foundOnContact kademlia.Contact, foundData []byte) {
	if foundData != nil {
		fmt.Println("Data found on contact:", foundOnContact.String())
		fmt.Println("Data:", string(foundData))
	} else {
		fmt.Println("Data not found.")
	}
}

func handlePut(k *kademlia.Kademlia, arg string) {
	if err := ValidatePutArg(arg); err != nil {
		fmt.Println(err)
		return
	}

	data := []byte(arg)
	kadId, targetContact := CreatePutTargetContact(data)
	contacts := performPutNodeLookup(k, targetContact)
	successCount := storeDataOnContacts(k, kadId, data, contacts)
	HandleStoreResult(successCount, len(contacts), kadId.String())
}

func ValidatePutArg(arg string) error {
	if arg == "" {
		return fmt.Errorf("Error: No argument provided for PUT.")
	}
	return nil
}

func CreatePutTargetContact(data []byte) (*kademlia.KademliaID, kademlia.Contact) {
	hasher := sha1.New()
	hasher.Write([]byte("hash1"))
	hash := hasher.Sum(nil)
	hashString := hex.EncodeToString(hash)
	kadId := kademlia.NewKademliaID(hashString)
	targetContact := kademlia.NewContact(kadId, "")
	return kadId, targetContact
}

func performPutNodeLookup(k *kademlia.Kademlia, targetContact kademlia.Contact) []kademlia.Contact {
	contacts, _, _ := k.NodeLookup(&targetContact, "")
	return contacts
}

func storeDataOnContacts(k *kademlia.Kademlia, kadId *kademlia.KademliaID, data []byte, contacts []kademlia.Contact) int {
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
	return successCount
}

func HandleStoreResult(successCount, totalContacts int, data string) {
	if successCount > totalContacts/2 {
		fmt.Println("Data stored successfully. Hash: " + data)
	} else {
		fmt.Println("Failed to store data.")
	}
}
