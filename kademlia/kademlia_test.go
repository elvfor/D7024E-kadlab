package kademlia

import (
	"testing"
)

func TestKademlia_UpdateRT(t *testing.T) {
	idSender := NewRandomKademliaID()
	contactSender := NewContact(idSender, "172.20.0.10:8000")
	contactSender.CalcDistance(idSender)
	rt := NewRoutingTable(contactSender)
	networkSender := &Network{}
	kSender := &Kademlia{RoutingTable: rt, Network: networkSender}

	idReceiver := NewRandomKademliaID()
	contactReceiver := NewContact(idReceiver, "172.20.0.11:8000")
	contactReceiver.CalcDistance(idReceiver)
	rtReceiver := NewRoutingTable(contactReceiver)
	networkReceiver := &Network{}
	kReceiver := &Kademlia{RoutingTable: rtReceiver, Network: networkReceiver}

	kSender.RoutingTable.AddContact(contactReceiver)

	kReceiver.UpdateRT(contactSender.ID.String(), contactSender.Address)
	contacts := kReceiver.RoutingTable.FindClosestContacts(contactSender.ID, 1)
	if contacts[0].ID.Equals(contactSender.ID) == false {
		t.Error("Contact not added to routing table")
	}
	if contacts[0].Address != contactSender.Address {
		t.Error("Contact address not added to routing table")
	}
	//TODO TEST DISTANCE
}

func TestKademlia_LookupContact(t *testing.T) {
	//TODO
}
