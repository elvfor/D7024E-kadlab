package cli

import (
	"bufio"
	"crypto/sha1"
	"d7024e/kademlia"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

type CLI struct {
	kademlia *kademlia.Kademlia
	reader   io.Reader
	writer   io.Writer
}

// NewCLI creates a new CLI instance with a Kademlia instance, reader, and writer
func NewCLI(k *kademlia.Kademlia) *CLI {
	return &CLI{
		kademlia: k,
		reader:   os.Stdin,
		writer:   os.Stdout,
	}
}

// ReadUserInput reads the input from the reader, trims and parses the command and argument
func (cli *CLI) ReadUserInput() (string, string, error) {
	consoleReader := bufio.NewReader(cli.reader)
	fmt.Fprint(cli.writer, ">")
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

// UserInputHandler continuously handles user input until the "EXIT" command is received
func (cli *CLI) UserInputHandler() bool {
	for {
		command, arg, err := cli.ReadUserInput()
		if err != nil {
			fmt.Fprintln(cli.writer, err)
			continue
		}
		if cli.handleCommand(command, arg) {
			return true
		}
	}
}

// handleCommand processes individual commands entered by the user
func (cli *CLI) handleCommand(command, arg string) bool {
	fmt.Fprintf(cli.writer, "You entered: command=%s, argument=%s\n", command, arg)

	switch command {
	case "GET":
		cli.handleGet(arg)
	case "PUT":
		cli.handlePut(arg)
	case "EXIT":
		fmt.Fprintln(cli.writer, "Exiting program.")
		return true
	case "PRINT":
		cli.kademlia.ActionChannel <- kademlia.Action{Action: "PRINT"}
	default:
		fmt.Fprintln(cli.writer, "Error: Unknown command.")
	}
	return false
}

// handleGet handles the "GET" command by performing a node lookup
func (cli *CLI) handleGet(arg string) {
	if err := cli.ValidateGetArg(arg); err != nil {
		fmt.Fprintln(cli.writer, err)
		return
	}

	targetContact := cli.CreateTargetContact(arg)
	foundOnContact, foundData := cli.performNodeLookup(targetContact, arg)
	cli.HandleLookupResult(foundOnContact, foundData)
}

// ValidateGetArg ensures the argument for GET is valid
func (cli *CLI) ValidateGetArg(arg string) error {
	if arg == "" {
		return fmt.Errorf("error: No argument provided for GET")
	}

	if len(arg) != 40 { // Kademlia ID length
		return fmt.Errorf("error: Invalid Kademlia ID length")
	}

	return nil
}

// CreateTargetContact creates a target Kademlia contact for lookup
func (cli *CLI) CreateTargetContact(arg string) kademlia.Contact {
	return kademlia.NewContact(kademlia.NewKademliaID(arg), "")
}

// performNodeLookup performs a node lookup using the Kademlia instance
func (cli *CLI) performNodeLookup(targetContact kademlia.Contact, arg string) (kademlia.Contact, []byte) {
	_, foundOnContact, foundData := cli.kademlia.NodeLookup(&targetContact, arg)
	return foundOnContact, foundData
}

// HandleLookupResult prints the result of the lookup
func (cli *CLI) HandleLookupResult(foundOnContact kademlia.Contact, foundData []byte) {
	if foundData != nil {
		fmt.Fprintln(cli.writer, "Data found on contact:", foundOnContact.String())
		fmt.Fprintln(cli.writer, "Data:", string(foundData))
	} else {
		fmt.Fprintln(cli.writer, "Data not found.")
	}
}

// handlePut handles the "PUT" command by storing data on contacts
func (cli *CLI) handlePut(arg string) {
	if err := cli.ValidatePutArg(arg); err != nil {
		fmt.Fprintln(cli.writer, err)
		return
	}

	data := []byte(arg)
	kadId, targetContact := cli.CreatePutTargetContact(data)
	contacts := cli.performPutNodeLookup(targetContact)
	successCount := cli.storeDataOnContacts(kadId, data, contacts)
	cli.HandleStoreResult(successCount, len(contacts), kadId.String())
}

// ValidatePutArg ensures the argument for PUT is valid
func (cli *CLI) ValidatePutArg(arg string) error {
	if arg == "" {
		return fmt.Errorf("error: No argument provided for PUT")
	}
	return nil
}

// CreatePutTargetContact creates a target Kademlia contact for storing data
func (cli *CLI) CreatePutTargetContact(data []byte) (*kademlia.KademliaID, kademlia.Contact) {
	hasher := sha1.New()
	hasher.Write(data)
	hash := hasher.Sum(nil)
	hashString := hex.EncodeToString(hash)
	kadId := kademlia.NewKademliaID(hashString)
	targetContact := kademlia.NewContact(kadId, "")
	return kadId, targetContact
}

// performPutNodeLookup performs a node lookup for storing data
func (cli *CLI) performPutNodeLookup(targetContact kademlia.Contact) []kademlia.Contact {
	contacts, _, _ := cli.kademlia.NodeLookup(&targetContact, "")
	return contacts
}

// storeDataOnContacts stores data on the provided contacts
func (cli *CLI) storeDataOnContacts(kadId *kademlia.KademliaID, data []byte, contacts []kademlia.Contact) int {
	resultChan := make(chan bool, len(contacts))
	var wg sync.WaitGroup

	for _, contact := range contacts {
		wg.Add(1)
		go func(contact kademlia.Contact) {
			defer wg.Done()
			result := cli.kademlia.Network.SendStoreMessage(&cli.kademlia.RoutingTable.Me, &contact, kadId, data)
			fmt.Fprintln(cli.writer, "Storing data with key:", kadId.String(), "on contact:", contact.String())
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

// HandleStoreResult prints the result of storing data
func (cli *CLI) HandleStoreResult(successCount, totalContacts int, data string) {
	if successCount > totalContacts/2 {
		fmt.Fprintln(cli.writer, "Data stored successfully. Hash: "+data)
	} else {
		fmt.Fprintln(cli.writer, "Failed to store data.")
	}
}
