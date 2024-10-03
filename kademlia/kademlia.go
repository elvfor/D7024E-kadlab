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

func (kademlia *Kademlia) SendAlphaFindNodeMessages(shortList []ShortListItem, target *Contact) ([]ShortListItem, int) {
	var wg sync.WaitGroup

	// Get alpha (number of nodes) that haven't been probed yet
	notProbed := kademlia.GetAlphaNotProbed(shortList)
	// Channel to hold individual Contact responses
	contactsChan := make(chan Contact)
	// Start goroutines to send FindNode messages asynchronously
	for _, contact := range notProbed {
		wg.Add(1)
		go kademlia.SendNodeLookup(target, contact.Contact, contactsChan, &wg)
		fmt.Println("DEBUG: Sent FindNode message to", contact.Contact.String())
	}

	// Close the channel once all goroutines have completed
	go func() {
		wg.Wait()           // Wait for all goroutines to complete
		close(contactsChan) // Close the channel to stop the range loop
		fmt.Println("DEBUG: Closed contactsChan")
	}()

	// Collect all contacts from the channel and update the shortList
	for contact := range contactsChan {
		shortList = UpdateShortList(shortList, contact, target.ID)
	}

	// Return the updated shortList and the number of contacts probed
	return shortList, len(notProbed)
}

func (kademlia *Kademlia) SendNodeLookup(target *Contact, contact Contact, contactsChan chan Contact, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("DEBUG: Starting SendNodeLookup for", contact.String())

	contacts, err := kademlia.Network.SendFindContactMessage(&kademlia.RoutingTable.Me, &contact, target)
	if err != nil {
		fmt.Println("DEBUG: Error sending message to", contact.String(), err)
		return
	}

	fmt.Println("DEBUG: Received contacts from", contact.String(), contacts)
	for _, foundContact := range contacts {
		select {
		case contactsChan <- foundContact:
			fmt.Println("DEBUG: Sending contact to channel", foundContact.String())
		default:
			fmt.Println("DEBUG: Channel is full, contact not sent:", foundContact.String())
		}
	}

	fmt.Println("DEBUG: Completed SendNodeLookup for", contact.String())
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

func (kademlia *Kademlia) ListenActionChannel() {
	fmt.Println("DEBUG: Listening to action channel")
	for {
		action := <-kademlia.ActionChannel
		fmt.Println("DEBUG: Received action", action)
		switch action.Action {
		case "UpdateRT":
			kademlia.UpdateRT(action.SenderId, action.SenderIp)
		case "Store":
			kademlia.Store(action.Hash, action.Data)
			kademlia.UpdateRT(action.SenderId, action.SenderIp)
		case "LookupContact":
			kademlia.UpdateRT(action.SenderId, action.SenderIp)
			contacts := kademlia.LookupContact(action.Target)
			//send contacts back to channel
			response := Response{
				ClosestContacts: contacts,
			}
			kademlia.Network.reponseChan <- response
		case "LookupData":
			kademlia.UpdateRT(action.SenderId, action.SenderIp)
			data, contacts := kademlia.LookupData(action.Hash)
			response := Response{
				Data:            data,
				ClosestContacts: contacts,
			}
			kademlia.Network.reponseChan <- response
		}
	}
}
