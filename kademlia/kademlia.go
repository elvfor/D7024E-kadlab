package kademlia

import "sort"

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

	alphaContacts := kademlia.RoutingTable.FindClosestContacts(target.ID, alpha)
	var shortList []ShortListItem
	for _, contact := range alphaContacts {
		shortList = UpdateShortList(shortList, contact, target.ID)
	}
	var probeCount int
	probeCount = 0
	// While probecount is less than 20
	for probeCount < k {
		closestNode := shortList[0].contact
		var tempProbeCount int
		//Send FIND_NODE to alpha closest contacts in shortlist that have not been probed
		shortList, tempProbeCount = kademlia.SendAlphaFindNodeMessages(shortList, target)
		if closestNode.distance.Less(shortList[0].distanceToTarget) {
			kClosestsContacts := kademlia.RoutingTable.FindClosestContacts(target.ID, k)
			for _, contact := range kClosestsContacts {
				go func() {
					contacts, err := kademlia.Network.SendFindContactMessage(&kademlia.RoutingTable.Me, &contact, target)
					if err != nil {
						for _, contact := range contacts {
							shortList = UpdateShortList(shortList, contact, target.ID)
						}
					}
				}()
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
	for i := 0; i < k; i++ {
		if shortList[i].probed == false {
			if alphaCount < alpha {
				go func() {
					contacts, err := kademlia.Network.SendFindContactMessage(&kademlia.RoutingTable.Me, &shortList[i].contact, target)
					if err != nil {
						shortList[i].probed = true
						probedCount++
						alphaCount++
						for _, contact := range contacts {
							shortList = UpdateShortList(shortList, contact, target.ID)
						}
					}
				}()
			} else {
				break
			}
		}
	}
	return shortList, probedCount
}
