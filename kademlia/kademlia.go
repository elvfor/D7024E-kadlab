package kademlia

import (
	"fmt"
	"sort"
	"sync"
)

// THREE RPC functions
type Kademlia struct {
	RoutingTable *RoutingTable
	Network      *Network
	Data         *map[string][]byte
}

type ShortListItem struct {
	Contact          Contact
	DistanceToTarget *KademliaID
	Probed           bool
}

const alpha = 3
const k = 5

// Constructor for Kademlia
func NewKademlia(table RoutingTable, network Network, data map[string][]byte) *Kademlia {
	return &Kademlia{&table, &network, &data}
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

func (kademlia *Kademlia) NodeLookup(target *Contact) []Contact {

	alphaContacts := kademlia.RoutingTable.FindClosestContacts(target.ID, alpha)
	if alphaContacts == nil {
		fmt.Println("DEBUG: alphaContacts is nil")
		return nil
	}

	var shortList []ShortListItem
	for _, contact := range alphaContacts {
		shortList = UpdateShortList(shortList, contact, target.ID)
	}

	var probeCount int
	probeCount = 0

	for probeCount < k {
		closestNode := shortList[0]
		var tempProbeCount int
		fmt.Println("DEBUG: shortList Before", shortList)
		shortList, tempProbeCount = kademlia.SendAlphaFindNodeMessages(shortList, target)
		fmt.Println("DEBUG: shortList After", shortList)
		if closestNode.DistanceToTarget.Less(shortList[0].DistanceToTarget) {
			kClosestsContacts := kademlia.RoutingTable.FindClosestContacts(target.ID, k)
			for _, contact := range kClosestsContacts {
				go func(contact Contact) {
					fmt.Println("DEBUG: Node lookup")
					contacts, err := kademlia.Network.SendFindContactMessage(&kademlia.RoutingTable.Me, &contact, target)
					if err == nil {
						for _, contact := range contacts {
							shortList = UpdateShortList(shortList, contact, target.ID)
						}
					}
				}(contact)
			}
			break
		} else {
			probeCount += tempProbeCount
		}
	}
	return GetAllContactsFromShortList(shortList)
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
		kademlia.RoutingTable.AddContact(NewDiscoveredContact)
	}
}

// UpdateShortList updates the shortlist with the new contact, list sorted by distance to target
func UpdateShortList(shortList []ShortListItem, newContact Contact, target *KademliaID) []ShortListItem {
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

func (kademlia *Kademlia) SendAlphaFindNodeMessages(shortList []ShortListItem, target *Contact) ([]ShortListItem, int) {
	var wg sync.WaitGroup

	// Channel to hold individual Contact responses
	contactsChan := make(chan Contact)

	// Get alpha (number of nodes) that haven't been probed yet
	notProbed := kademlia.GetAlphaNotProbed(shortList)

	// Start goroutines to send FindNode messages asynchronously
	for _, contact := range notProbed {
		wg.Add(1)
		go func(contact Contact) {
			defer wg.Done()
			fmt.Println("DEBUG: Sending FindNode message to", contact.String())

			// Send FindNode message
			contacts, err := kademlia.Network.SendFindContactMessage(&kademlia.RoutingTable.Me, &contact, target)
			if err == nil {
				// Send each contact from the response to the channel
				for _, foundContact := range contacts {
					contactsChan <- foundContact
				}
			}
		}(contact.Contact)
	}

	// Close the channel once all goroutines have finished sending contacts
	go func() {
		wg.Wait()
		close(contactsChan)
	}()

	// Collect all contacts from the channel and update the shortList
	for contact := range contactsChan {
		shortList = UpdateShortList(shortList, contact, target.ID)
	}

	// Return the updated shortList and the number of contacts probed
	return shortList, len(notProbed)
}

func (kademlia *Kademlia) GetAlphaNotProbed(shortList []ShortListItem) []ShortListItem {
	var notProbed []ShortListItem
	for _, item := range shortList {
		if !item.Probed {
			notProbed = append(notProbed, item)
		}
	}
	if len(notProbed) < alpha {
		return notProbed
	}
	fmt.Println("DEBUG: Not probed", notProbed)
	return notProbed[:alpha]
}
