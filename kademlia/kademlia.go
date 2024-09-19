package kademlia

// THREE RPC functions
type Kademlia struct {
	RoutingTable *RoutingTable
	Network      *Network
	Data         *map[string][]byte
}

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
	//TODO
	return nil
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
