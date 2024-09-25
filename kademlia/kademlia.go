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
	contact          Contact
	distanceToTarget *KademliaID
	probed           bool
}

const alpha = 3
const k = 20

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

// NODE LOOKUP
func (kademlia *Kademlia) NodeLookup(target *Contact) []Contact {
	if target == nil {
		fmt.Println("DEBUG: Target contact is nil")
		return nil
	}
	alphaContacts := kademlia.RoutingTable.FindClosestContacts(target.ID, alpha)

	var shortList []ShortListItem
	for _, contact := range alphaContacts {
		shortList = UpdateShortList(shortList, contact, target.ID)
	}

	var probeCount int
	probeCount = 0

	for probeCount < k {
		if len(shortList) == 0 {
			break
		}
		closestNode := shortList[0].contact
		var tempProbeCount int
		shortList, tempProbeCount = kademlia.SendAlphaFindNodeMessages(shortList, target)
		if closestNode.distance.Less(shortList[0].distanceToTarget) {
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

func (kademlia *Kademlia) UpdateRT(id string, ip string) {
	NewDiscoveredContact := NewContact(NewKademliaID(id), ip)
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
		return shortList[i].distanceToTarget.Less(shortList[j].distanceToTarget)
	})
	if len(shortList) < k {
		return shortList
	}
	return shortList[:k]
}
func GetAllContactsFromShortList(shortList []ShortListItem) []Contact {
	var contacts []Contact
	for _, item := range shortList {
		contacts = append(contacts, item.contact)
	}
	return contacts
}

func (kademlia *Kademlia) SendAlphaFindNodeMessages(shortList []ShortListItem, target *Contact) ([]ShortListItem, int) {
	var alphaCount int
	alphaCount = 0
	var probedCount int
	probedCount = 0
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Clone the shortlist to ensure consistent state during iteration
	shortListCopy := make([]ShortListItem, len(shortList))
	copy(shortListCopy, shortList)

	for i := 0; i < len(shortListCopy) && i < k; i++ {
		// Protect access to shortList[i].probed with mutex
		mu.Lock()
		if shortListCopy[i].probed == false && alphaCount < alpha {
			shortListCopy[i].probed = true // Mark as probed to avoid repeated checks
			alphaCount++
			wg.Add(1)
			mu.Unlock() // Release the lock before entering the goroutine
			go func(i int) {
				defer wg.Done()
				contacts, err := kademlia.Network.SendFindContactMessage(&kademlia.RoutingTable.Me, &shortList[i].contact, target)
				mu.Lock()
				defer mu.Unlock()
				if err == nil {
					probedCount++
					// Add new contacts to the shortlist
					for _, contact := range contacts {
						shortList = UpdateShortList(shortList, contact, target.ID)
					}
				} else {
					fmt.Printf("error sending FIND_NODE message: %v\n", err)
				}
			}(i)
		} else {
			mu.Unlock() // Always release the lock
		}
	}
	wg.Wait()
	return shortList, probedCount
}
