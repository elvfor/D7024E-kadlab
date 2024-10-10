package kademlia

import (
	"fmt"
	"sort"
	"sync"
)

// THREE RPC functions
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
func NewKademlia(table *RoutingTable) *Kademlia {
	network := NewNetwork()
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
	//TODO is this correct?
	contact := NewContact(NewKademliaID(hash), "")
	closestContacts := kademlia.LookupContact(&contact)
	return nil, closestContacts
}

func (kademlia *Kademlia) NodeLookup(target *Contact, hash string) ([]Contact, Contact, []byte) {

	alphaContacts := kademlia.RoutingTable.FindClosestContacts(target.ID, alpha)

	var shortList []ShortListItem
	for _, contact := range alphaContacts {
		shortList = UpdateShortList(shortList, contact, target.ID)
	}

	closestNode := shortList[0]

	for {
		temp := kademlia.GetAlphaNodes(shortList)
		if len(temp) == 0 {
			return GetAllContactsFromShortList(shortList), Contact{}, nil
		}
		notProbed := kademlia.GetAlphaNodes(shortList)
		var contactFoundDataOn Contact
		var foundData []byte

		shortList, contactFoundDataOn, foundData = kademlia.SendAlphaFindNodeMessages(shortList, target, hash, notProbed)
		fmt.Println("DEBUG: Shortlist after sending messages", shortList)
		if foundData != nil {
			fmt.Println("Done with Node lookup, found data")
			return GetAllContactsFromShortList(shortList), contactFoundDataOn, foundData
		}
		newClosestNode := shortList[0]
		if closestNode.Contact.ID.Equals(newClosestNode.Contact.ID) {
			temp2 := kademlia.GetAlphaNodes(shortList)
			if CountProbedInShortList(shortList) >= k || len(temp2) == 0 {
				break
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

// STORE
func (kademlia *Kademlia) Store(hash string, data []byte) {
	(*kademlia.Data)[hash] = data
}

func (kademlia *Kademlia) UpdateRT(id *KademliaID, ip string) {
	NewDiscoveredContact := NewContact(id, ip)
	if !(NewDiscoveredContact.ID.Equals(kademlia.RoutingTable.Me.ID)) {
		fmt.Println("Adding contact to routing table with ID: ", NewDiscoveredContact.ID.String()+" and IP: "+NewDiscoveredContact.Address+" on"+kademlia.RoutingTable.Me.Address)
		NewDiscoveredContact.CalcDistance(kademlia.RoutingTable.Me.ID)
		bucketIsFull, lastContact := kademlia.RoutingTable.AddContact(NewDiscoveredContact)
		if bucketIsFull {
			//send ping to lastContact to see if it is alive
			if kademlia.Network.SendPingMessage(&kademlia.RoutingTable.Me, lastContact) {
				fmt.Println("Last contact is alive, discard new contact")
			} else {
				fmt.Println("Last contact is dead, replace with new contact")
				kademlia.RoutingTable.RemoveContact(lastContact)
				kademlia.RoutingTable.AddContact(NewDiscoveredContact)
			}
		}
	}
}

// UpdateShortList updates the shortlist with the new contact, list sorted by distance to target
func UpdateShortList(shortList []ShortListItem, newContact Contact, target *KademliaID) []ShortListItem {
	//if the new contact is already in the shortlist, dont add it
	for _, item := range shortList {
		if item.Contact.ID.Equals(newContact.ID) {
			return shortList
		}
	}
	newDistance := newContact.ID.CalcDistance(target)
	newItem := ShortListItem{newContact, newDistance, false}
	shortList = append(shortList, newItem)
	sort.Slice(shortList, func(i, j int) bool {
		return shortList[i].DistanceToTarget.Less(shortList[j].DistanceToTarget)
	})
	if len(shortList) < k {
		return shortList
	}
	return shortList[:k]
}
func GetAllContactsFromShortList(shortList []ShortListItem) []Contact {
	var contacts []Contact
	for _, item := range shortList {
		contacts = append(contacts, item.Contact)
	}
	return contacts
}

func (kademlia *Kademlia) SendAlphaFindNodeMessages(shortList []ShortListItem, target *Contact, hash string, notProbed []ShortListItem) ([]ShortListItem, Contact, []byte) {
	var wg sync.WaitGroup

	contactsChan := make(chan Contact, alpha*k)
	dataChan := make(chan []byte, alpha*k)
	contactChanFoundDataOn := make(chan Contact, alpha*k)

	for _, contact := range notProbed {
		wg.Add(1)
		go func(contact Contact) {
			defer wg.Done()
			if hash == "" {
				kademlia.findContact(contact, target, contactsChan, dataChan, contactChanFoundDataOn)
			} else {
				kademlia.findData(contact, hash, dataChan, contactChanFoundDataOn)
				fmt.Println("DEBUG: Done with FindData")
			}
		}(contact.Contact)
	}

	// Wait for all goroutines to finish before proceeding
	wg.Wait()

	// Close the channels now that all goroutines are done
	close(contactsChan)
	close(dataChan)
	close(contactChanFoundDataOn)
	fmt.Println("DEBUG: Done waiting for all contacts")

	// Wait for data from dataChan or contacts from contactsChan
	var foundData []byte
	var foundContact Contact

	// Using select to retrieve found data
	select {
	case data := <-dataChan:
		fmt.Println("DEBUG: Data received from dataChan")
		if data != nil {
			foundData = data
			if contact := <-contactChanFoundDataOn; contact.ID != nil {
				foundContact = contact
				fmt.Println("DEBUG: Returning data and found contact")
				return shortList, foundContact, foundData
			}
		}
	}

	// Handle contacts from contactsChan
	for contact := range contactsChan {
		shortList = UpdateShortList(shortList, contact, target.ID)
	}

	for i, item := range shortList {
		for _, probedContact := range notProbed {
			if item.Contact.ID.Equals(probedContact.Contact.ID) {
				shortList[i].Probed = true
			}
		}
	}

	return shortList, Contact{}, nil
}

func (kademlia *Kademlia) findContact(contact Contact, target *Contact, contactsChan chan Contact, dataChan chan []byte, contactChanFoundDataOn chan Contact) {
	contacts, err := kademlia.Network.SendFindContactMessage(&kademlia.RoutingTable.Me, &contact, target)
	if err != nil {
		return
	}
	for _, foundContact := range contacts {
		select {
		case contactsChan <- foundContact:
			fmt.Println("DEBUG: ALL CONTACTS RECEIVED TO ADD TO SHORTLIST", foundContact.String())
			dataChan <- nil
			contactChanFoundDataOn <- Contact{}
		default:
			fmt.Println("DEBUG: Channel is full, contact not sent:", foundContact.String())
		}
	}
}

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
	fmt.Println("DEBUG: Not probed", notProbed)
	return notProbed[:alpha]
}

func (kademlia *Kademlia) GetAlphaNodesFromKClosest(shortList []ShortListItem, target *Contact) []ShortListItem {
	var notProbed []ShortListItem
	alphaContacts := kademlia.RoutingTable.FindClosestContacts(target.ID, k)

	for _, item := range alphaContacts {
		//if the new contact is already in the shortlist, dont add it
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
	fmt.Println("DEBUG: Not probed", notProbed)
	return notProbed[:alpha]
}

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
			contacts := kademlia.LookupContact(action.Target)
			//send contacts back to channel
			response := Response{
				ClosestContacts: contacts,
			}
			kademlia.Network.reponseChan <- response
		case "LookupData":
			data, contacts := kademlia.LookupData(action.Hash)
			response := Response{
				Data:            data,
				ClosestContacts: contacts,
			}
			kademlia.Network.reponseChan <- response
		case "PRINT":
			kademlia.RoutingTable.PrintAllIP()
			/*case "NODELOOKUP":
			contacts, target := kademlia.NodeLookup(action.Target)
			//send contacts back to channel
			response := Response{
				ClosestContacts: contacts,
				Target:          target,
			}
			for _, contact := range contacts {
				kademlia.UpdateRT(contact.ID, contact.Address)
			}

			*/
		}

	}
}

func CountProbedInShortList(shortList []ShortListItem) int {
	count := 0
	for _, item := range shortList {
		if item.Probed {
			count++
		}
	}
	return count
}
