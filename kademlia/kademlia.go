package kademlia

import (
	"fmt"
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

type ShortList struct {
	ls  []ShortListItem
	mux sync.Mutex
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
	var shortList ShortList
	//c := make(chan int, alpha)
	if len(alphaContacts) == 0 {
		fmt.Println("DEBUG: alphaContacts is empty")
		return []Contact{}
	}
	if alphaContacts == nil {
		fmt.Println("DEBUG: alphaContacts is nil")
		return nil
	}

	for _, contact := range alphaContacts {
		shortList.UpdateShortList(contact, target.ID)
	}

	probeCount := 0

	for probeCount < k {
		closestNode := shortList.ls[0]
		var tempProbeCount int
		fmt.Println("DEBUG: shortList Before", shortList.ls)
		shortList.ls, tempProbeCount = kademlia.SendAlphaFindNodeMessages(shortList.ls, target)
		fmt.Println("DEBUG: shortList After", shortList.ls)

		if closestNode.DistanceToTarget.Less(shortList.ls[0].DistanceToTarget) {
			break
		} else {
			probeCount += tempProbeCount
		}
	}

	return shortList.GetAllContacts()
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

// Append method for ShortList
func (shortList *ShortList) Append(item ShortListItem) {
	shortList.mux.Lock()
	defer shortList.mux.Unlock()
	shortList.ls = append(shortList.ls, item)
}

// UpdateShortList updates the shortlist with the new contact, sorted by distance to target
func (shortList *ShortList) UpdateShortList(newContact Contact, target *KademliaID) {
	shortList.mux.Lock()
	defer shortList.mux.Unlock()

	// Calculate distance of new contact to the target
	newContactDist := newContact.ID.CalcDistance(target)

	// Check if the contact is already in the shortlist
	for _, item := range shortList.ls {
		if item.Contact.ID.Equals(newContact.ID) {
			return
		}
	}

	// Find the correct position to insert the new contact based on distance
	inserted := false
	for i, item := range shortList.ls {
		// Compare the distances
		if newContactDist.Less(item.DistanceToTarget) {
			// Insert the new contact at the correct position
			shortList.ls = append(shortList.ls[:i], append([]ShortListItem{
				{Contact: newContact, DistanceToTarget: newContactDist, Probed: false}},
				shortList.ls[i:]...)...)
			inserted = true
			break
		}
	}

	// If new contact is the farthest, append it to the end
	if !inserted {
		shortList.ls = append(shortList.ls, ShortListItem{
			Contact: newContact, DistanceToTarget: newContactDist, Probed: false,
		})
	}

	// Ensure the shortlist doesn't exceed `k` size, trimming if necessary
	if len(shortList.ls) > k {
		shortList.ls = shortList.ls[:k]
	}
}

func (shortList *ShortList) GetAllContacts() []Contact {
	var contacts []Contact
	for _, item := range shortList.ls {
		contacts = append(contacts, item.Contact)
	}
	return contacts
}

func (kademlia *Kademlia) SendAlphaFindNodeMessages(shortList []ShortListItem, target *Contact) ([]ShortListItem, int) {
	var wg sync.WaitGroup

	// Get alpha (number of nodes) that haven't been probed yet
	notProbed := kademlia.GetAlphaNotProbed(shortList)
	// Channel to hold individual Contact responses
	contactsChan := make(chan Contact, alpha*k)
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
		shortList = append(shortList, ShortListItem{
			Contact: contact, DistanceToTarget: contact.ID.CalcDistance(target.ID), Probed: false,
		})
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
	//check if Me is in the shortlist
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
		case "Print":
			fmt.Println("DEBUG: Printing routing table...")
			kademlia.RoutingTable.PrintAllIP() // Add this line to print
		}
	}
}
