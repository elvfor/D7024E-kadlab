package kademlia

// THREE RPC functions
type Kademlia struct {
	RoutingTable *RoutingTable
	Network      *Network
}

// Constructor for Kademlia
func NewKademlia(table RoutingTable, network Network) *Kademlia {
	return &Kademlia{&table, &network}
}

// FIND_NODE
func (kademlia *Kademlia) LookupContact(target *Contact) []Contact {
	closestContacts := kademlia.RoutingTable.FindClosestContacts(target.ID, 20)
	return closestContacts
}

// FIND_VALUE
func (kademlia *Kademlia) LookupData(hash string) {
	// TODO
}

// STORE
func (kademlia *Kademlia) Store(data []byte) {
	// TODO
}

func (kademlia *Kademlia) UpdateRT(id string, ip string) {
	NewDiscoveredContact := NewContact(NewKademliaID(id), ip)
	kademlia.RoutingTable.AddContact(NewDiscoveredContact)
}
