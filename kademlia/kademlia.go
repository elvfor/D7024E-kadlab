package kademlia

import "fmt"

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
func (kademlia *Kademlia) LookupContact(target *Contact) {
	// TODO
	fmt.Println("LookupContact")
}

// FIND_VALUE
func (kademlia *Kademlia) LookupData(hash string) {
	// TODO
}

// STORE
func (kademlia *Kademlia) Store(data []byte) {
	// TODO
}

func (kademlia *Kademlia) HandlePing(ip string) {
	NewContact()
	kademlia.RoutingTable.AddContact(*pingedBy)
}
