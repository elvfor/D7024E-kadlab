package kademlia

import (
	"fmt"
	"net"
	"sort"
	"sync"
)

type Kademlia struct {
	RoutingTable  *RoutingTable
	Network       *Network
	Data          *map[string][]byte
	ActionChannel chan Action
}

type Action struct {
	Action   string
	Target   *Contact
	Hash     string
	Data     []byte
	SenderId *KademliaID
	SenderIp string
}

type ShortListItem struct {
	Contact          Contact
	DistanceToTarget *KademliaID
	Probed           bool
}

const alpha = 3
const k = 5

// Constructor for Kademlia
func NewKademlia(table *RoutingTable, conn net.PacketConn) *Kademlia {
	network := NewNetwork(conn)
	data := make(map[string][]byte)
	actionChannel := make(chan Action)
	return &Kademlia{table, network, &data, actionChannel}
}

// FIND_NODE
func (kademlia *Kademlia) LookupContact(target *Contact) []Contact {
	closestContacts := kademlia.RoutingTable.FindClosestContacts(target.ID, 20)
	return closestContacts
}

// FIND_VALUE
func (kademlia *Kademlia) LookupData(hash string) ([]byte, []Contact) {
	if data, ok := (*kademlia.Data)[hash]; ok {
		return data, nil
	}

	contact := NewContact(NewKademliaID(hash), "")
	closestContacts := kademlia.LookupContact(&contact)
	return nil, closestContacts
}

// STORE
func (kademlia *Kademlia) Store(hash string, data []byte) {
	(*kademlia.Data)[hash] = data
}

// NodeLookup is the main function for the NodeLookup algorithm
func (kademlia *Kademlia) NodeLookup(target *Contact, hash string) ([]Contact, Contact, []byte) {

	// Initialize the shortlist with the alpha closest contacts
	alphaContacts := kademlia.RoutingTable.FindClosestContacts(target.ID, alpha)
	var shortList []ShortListItem
	for _, contact := range alphaContacts {
		shortList = UpdateShortList(shortList, contact, target.ID)
	}

	closestNode := shortList[0]

	// Loop until k nodes are probed or no more unprobed nodes
	for {
		// Get the alpha closest unprobed contacts in the shortlist
		temp := kademlia.GetAlphaNodes(shortList)
		// If there are no unprobed contacts left (or k probed), return the shortlist
		if len(temp) == 0 {
			return GetAllContactsFromShortList(shortList), Contact{}, nil
		}
		notProbed := kademlia.GetAlphaNodes(shortList)
		var contactFoundDataOn Contact
		var foundData []byte

		// Call to send alpha FIND_NODE messages
		shortList, contactFoundDataOn, foundData = kademlia.SendAlphaFindNodeMessages(shortList, target, hash, notProbed)
		//fmt.Println("DEBUG: Shortlist after sending messages", shortList)
		// If data is found on a contact, return the contact and data
		if foundData != nil {
			fmt.Println("Done with Node lookup, found data")
			return GetAllContactsFromShortList(shortList), contactFoundDataOn, foundData
		}
		newClosestNode := shortList[0]

		// If the closest node is the same as before, check that k nodes have been probed or no more unprobed
		if closestNode.Contact.ID.Equals(newClosestNode.Contact.ID) {
			temp2 := kademlia.GetAlphaNodes(shortList)
			if CountProbedInShortList(shortList) >= k || len(temp2) == 0 {
				break
				// If there are unprobed nodes left, get alpha nodes from own routing table and send FIND_NODE messages
			} else {
				notProbedKClosest := kademlia.GetAlphaNodesFromKClosest(shortList, target)
				newShortList, _, _ := kademlia.SendAlphaFindNodeMessages(shortList, target, hash, notProbedKClosest)
				shortList = newShortList
			}
		} else {
			// Update the closest node and continue
			closestNode = newClosestNode
		}

	}
	fmt.Println("Done with Node lookup ")
	return GetAllContactsFromShortList(shortList), Contact{}, nil
}

// UpdateRT updates the routing table with a new contact
func (kademlia *Kademlia) UpdateRT(id *KademliaID, ip string) {
	NewDiscoveredContact := NewContact(id, ip)
	if !(NewDiscoveredContact.ID.Equals(kademlia.RoutingTable.Me.ID)) {
		fmt.Println("Adding contact to routing table with ID: ", NewDiscoveredContact.ID.String()+" and IP: "+NewDiscoveredContact.Address+" on"+kademlia.RoutingTable.Me.Address)
		NewDiscoveredContact.CalcDistance(kademlia.RoutingTable.Me.ID)

		// Check if bucket is full
		bucketIsFull, lastContact := kademlia.RoutingTable.AddContact(NewDiscoveredContact)
		if bucketIsFull {
			// If so, send ping to lastContact to see if it is alive
			if kademlia.Network.SendPingMessage(&kademlia.RoutingTable.Me, lastContact) {
				fmt.Println("Last contact is alive, discard new contact")
			} else {
				// If not, replace lastContact with new contact
				fmt.Println("Last contact is dead, replace with new contact")
				kademlia.RoutingTable.RemoveContact(lastContact)
				kademlia.RoutingTable.AddContact(NewDiscoveredContact)
			}
		}
	}
}

// UpdateShortList updates the shortlist with the new contact, list sorted by distance to target
func UpdateShortList(shortList []ShortListItem, newContact Contact, target *KademliaID) []ShortListItem {
	// If the new contact is already in the shortlist, don't add it
	for _, item := range shortList {
		if item.Contact.ID.Equals(newContact.ID) {
			return shortList
		}
	}
	// Else add the new contact to the shortlist
	newDistance := newContact.ID.CalcDistance(target)
	newItem := ShortListItem{newContact, newDistance, false}
	shortList = append(shortList, newItem)

	// Sort the shortlist by distance to target
	sort.Slice(shortList, func(i, j int) bool {
		return shortList[i].DistanceToTarget.Less(shortList[j].DistanceToTarget)
	})
	if len(shortList) < k {
		return shortList
	}
	return shortList[:k]
}

// GetAllContactsFromShortList returns all contacts from the shortlist
func GetAllContactsFromShortList(shortList []ShortListItem) []Contact {
	var contacts []Contact
	for _, item := range shortList {
		contacts = append(contacts, item.Contact)
	}
	return contacts
}

// probeContacts probes the contacts in the shortlist concurrently performing either FIND_NODE or FIND_DATA
func (kademlia *Kademlia) probeContacts(notProbed []ShortListItem, target *Contact, hash string, contactsChan chan Contact, dataChan chan []byte, contactChanFoundDataOn chan Contact) {
	var wg sync.WaitGroup
	for _, contact := range notProbed {
		wg.Add(1)
		go func(contact Contact) {
			defer wg.Done()
			// If there is no hashed value to look for, do a FIND_NODE
			if hash == "" {
				kademlia.findContact(contact, target, contactsChan, dataChan, contactChanFoundDataOn)
				// If there is a hashed value, do FIND_DATA
			} else {
				kademlia.findData(contact, hash, dataChan, contactChanFoundDataOn)
				fmt.Println("DEBUG: Done with FindData")
			}
		}(contact.Contact)
	}
	wg.Wait() // Wait for all goroutines to finish
}

// closeChannels closes the channels
func closeChannels(contactsChan chan Contact, dataChan chan []byte, contactChanFoundDataOn chan Contact) {
	if contactsChan != nil {
		close(contactsChan)
	}
	if dataChan != nil {
		close(dataChan)
	}
	if contactChanFoundDataOn != nil {
		close(contactChanFoundDataOn)
	}
}

// handleFoundData handles the found data if any and returns the data and contact
func handleFoundData(dataChan chan []byte, contactChanFoundDataOn chan Contact) (Contact, []byte) {
	select {
	case data := <-dataChan:
		//fmt.Println("DEBUG: Data received from dataChan")
		if data != nil {
			foundContact := <-contactChanFoundDataOn
			if foundContact.ID != nil {
				//fmt.Println("DEBUG: Returning data and found contact")
				return foundContact, data
			}
		}
	}
	return Contact{}, nil
}

// updateShortListWithContacts updates the shortlist with the received contacts
func (kademlia *Kademlia) updateShortListWithContacts(shortList []ShortListItem, target *Contact, contactsChan chan Contact) []ShortListItem {
	for contact := range contactsChan {
		updated := false
		for i, item := range shortList {
			if item.Contact.ID.Equals(contact.ID) {
				shortList[i].Contact = contact
				updated = true
				break
			}
		}
		if !updated {
			shortList = UpdateShortList(shortList, contact, target.ID)
		}
	}
	return shortList
}

// markProbedContacts marks the probed contacts in the shortlist
func markProbedContacts(shortList []ShortListItem, notProbed []ShortListItem) []ShortListItem {
	for i, item := range shortList {
		for _, probedContact := range notProbed {
			if item.Contact.ID.Equals(probedContact.Contact.ID) {
				shortList[i].Probed = true
			}
		}
	}
	return shortList
}

// SendAlphaFindNodeMessages sends alpha FIND_NODE messages to the contacts in the shortlist
func (kademlia *Kademlia) SendAlphaFindNodeMessages(shortList []ShortListItem, target *Contact, hash string, notProbed []ShortListItem) ([]ShortListItem, Contact, []byte) {
	contactsChan := make(chan Contact, alpha*k)
	dataChan := make(chan []byte, alpha*k)
	contactChanFoundDataOn := make(chan Contact, alpha*k)

	// Probe the contacts concurrently
	kademlia.probeContacts(notProbed, target, hash, contactsChan, dataChan, contactChanFoundDataOn)

	// Close channels after probing
	closeChannels(contactsChan, dataChan, contactChanFoundDataOn)
	fmt.Println("DEBUG: Done waiting for all contacts")

	// Handle found data if any
	foundContact, foundData := handleFoundData(dataChan, contactChanFoundDataOn)
	if foundData != nil {
		return shortList, foundContact, foundData
	}

	// Update shortlist with received contacts
	shortList = kademlia.updateShortListWithContacts(shortList, target, contactsChan)

	// Mark probed contacts in the shortlist
	shortList = markProbedContacts(shortList, notProbed)

	return shortList, Contact{}, nil
}

// findContact sends a FIND_NODE message to a contact and returns the contacts found
func (kademlia *Kademlia) findContact(contact Contact, target *Contact, contactsChan chan Contact, dataChan chan []byte, contactChanFoundDataOn chan Contact) {
	contacts, err := kademlia.Network.SendFindContactMessage(&kademlia.RoutingTable.Me, &contact, target)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, foundContact := range contacts {
		select {
		case contactsChan <- foundContact:
			dataChan <- nil
			contactChanFoundDataOn <- Contact{}
		default:
			fmt.Println("DEBUG: Channel is full, contact not sent:", foundContact.String())
		}
	}
}

// findData sends a FIND_DATA message to a contact and returns the data if found
func (kademlia *Kademlia) findData(contact Contact, hash string, dataChan chan []byte, contactChanFoundDataOn chan Contact) {
	_, data, err := kademlia.Network.SendFindDataMessage(&kademlia.RoutingTable.Me, &contact, hash)
	if err != nil {
		return
	}
	if data != nil {
		fmt.Println("DEBUG: Found data INSIDE NODE LOOKUP on contact", contact.String())
		dataChan <- data
		contactChanFoundDataOn <- contact
	}
}

// GetAlphaNodes returns the alpha closest unprobed contacts in the shortlist
func (kademlia *Kademlia) GetAlphaNodes(shortList []ShortListItem) []ShortListItem {
	var notProbed []ShortListItem
	for _, item := range shortList {
		if !item.Probed && !item.Contact.ID.Equals(kademlia.RoutingTable.Me.ID) {
			notProbed = append(notProbed, item)
		}
	}
	if len(notProbed) < alpha {
		return notProbed
	}
	return notProbed[:alpha]
}

// GetAlphaNodesFromKClosest returns the alpha closest unprobed contacts from the k closest contacts of the target in the routing table
func (kademlia *Kademlia) GetAlphaNodesFromKClosest(shortList []ShortListItem, target *Contact) []ShortListItem {
	var notProbed []ShortListItem
	alphaContacts := kademlia.RoutingTable.FindClosestContacts(target.ID, k)

	for _, item := range alphaContacts {
		//if the new contact is already in the shortlist, don't add it
		for _, shortItem := range shortList {
			if item.ID.Equals(shortItem.Contact.ID) || len(notProbed) >= alpha {
				continue
			} else {
				notProbed = append(notProbed, ShortListItem{item, item.ID.CalcDistance(target.ID), false})
			}
		}
	}
	if len(notProbed) < alpha {
		return notProbed
	}
	return notProbed[:alpha]
}

// ListenActionChannel listens to the action channel and performs the action received
func (kademlia *Kademlia) ListenActionChannel() {
	fmt.Println("DEBUG: Listening to action channel")
	for {
		action := <-kademlia.ActionChannel
		fmt.Println("DEBUG: Received action", action)
		switch action.Action {
		case "UpdateRT":
			fmt.Println("DEBUG: Updating RT")
			kademlia.UpdateRT(action.SenderId, action.SenderIp)
		case "Store":
			kademlia.Store(action.Hash, action.Data)
		case "LookupContact":
			fmt.Println("DEBUG: Looking up contact")
			contacts := kademlia.LookupContact(action.Target)
			//send contacts back to channel
			response := Response{
				ClosestContacts: contacts,
			}
			fmt.Println("DEBUG: Sending response", response)
			kademlia.Network.responseChan <- response
		case "LookupData":
			data, contacts := kademlia.LookupData(action.Hash)
			response := Response{
				Data:            data,
				ClosestContacts: contacts,
			}
			kademlia.Network.responseChan <- response
		case "PRINT":
			kademlia.RoutingTable.PrintAllIP()
		}
	}
}

// CountProbedInShortList counts the number of probed contacts in the shortlist
func CountProbedInShortList(shortList []ShortListItem) int {
	count := 0
	for _, item := range shortList {
		if item.Probed {
			count++
		}
	}
	return count
}
