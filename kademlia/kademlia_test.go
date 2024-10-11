package kademlia

import (
	"fmt"
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

	kReceiver.UpdateRT(contactSender.ID, contactSender.Address)
	contacts := kReceiver.RoutingTable.FindClosestContacts(contactSender.ID, 1)
	if contacts[0].ID.Equals(contactSender.ID) == false {
		t.Error("Contact not added to routing table")
	}
	if contacts[0].Address != contactSender.Address {
		t.Error("Contact address not added to routing table")
	}
	//TODO TEST DISTANCE
}
func TestKademlia_UpdateRT_AddNewContact(t *testing.T) {
	idSender := NewRandomKademliaID()
	contactSender := NewContact(idSender, "172.20.0.10:8000")
	contactSender.CalcDistance(idSender)
	rt := NewRoutingTable(contactSender)
	networkSender := &Network{}
	_ = &Kademlia{RoutingTable: rt, Network: networkSender}

	idReceiver := NewRandomKademliaID()
	contactReceiver := NewContact(idReceiver, "172.20.0.11:8000")
	contactReceiver.CalcDistance(idReceiver)
	rtReceiver := NewRoutingTable(contactReceiver)
	networkReceiver := &Network{}
	kReceiver := &Kademlia{RoutingTable: rtReceiver, Network: networkReceiver}

	kReceiver.UpdateRT(contactSender.ID, contactSender.Address)
	contacts := kReceiver.RoutingTable.FindClosestContacts(contactSender.ID, 1)
	if contacts[0].ID.Equals(contactSender.ID) == false {
		t.Error("Contact not added to routing table")
	}
	if contacts[0].Address != contactSender.Address {
		t.Error("Contact address not added to routing table")
	}
}

type MockNetwork struct {
	Network
}

func (m *MockNetwork) SendPingMessage(sender *Contact, receiver *Contact) bool {
	// This should be the method that is used in the test
	return false
}

func TestKademlia_UpdateRT_DoNotAddSelf(t *testing.T) {
	id := NewRandomKademliaID()
	contact := NewContact(id, "172.20.0.10:8000")
	contact.CalcDistance(id)
	rt := NewRoutingTable(contact)
	network := &Network{}
	k := &Kademlia{RoutingTable: rt, Network: network}

	k.UpdateRT(id, "172.20.0.10:8000")
	contacts := k.RoutingTable.FindClosestContacts(id, 1)
	if len(contacts) != 0 {
		t.Error("Self contact should not be added to routing table")
	}
}

func TestUpdateShortList_AddsNewContact(t *testing.T) {
	targetID := NewRandomKademliaID()
	contact := NewContact(NewRandomKademliaID(), "172.20.0.10:8000")
	shortList := []ShortListItem{}

	updatedShortList := UpdateShortList(shortList, contact, targetID)

	if len(updatedShortList) != 1 {
		t.Errorf("Expected 1 contact in shortlist, got %d", len(updatedShortList))
	}
	if !updatedShortList[0].Contact.ID.Equals(contact.ID) {
		t.Errorf("Expected contact ID %s, got %s", contact.ID.String(), updatedShortList[0].Contact.ID.String())
	}
}

func TestUpdateShortList_DoesNotAddDuplicateContact(t *testing.T) {
	targetID := NewRandomKademliaID()
	contact := NewContact(NewRandomKademliaID(), "172.20.0.10:8000")
	shortList := []ShortListItem{
		{Contact: contact, DistanceToTarget: contact.ID.CalcDistance(targetID), Probed: false},
	}

	updatedShortList := UpdateShortList(shortList, contact, targetID)

	if len(updatedShortList) != 1 {
		t.Errorf("Expected 1 contact in shortlist, got %d", len(updatedShortList))
	}
}

func TestUpdateShortList_RespectsMaxK(t *testing.T) {
	targetID := NewRandomKademliaID()
	shortList := []ShortListItem{}
	for i := 0; i < k; i++ {
		contact := NewContact(NewRandomKademliaID(), fmt.Sprintf("172.20.0.%d:8000", i))
		shortList = append(shortList, ShortListItem{Contact: contact, DistanceToTarget: contact.ID.CalcDistance(targetID), Probed: false})
	}
	newContact := NewContact(NewRandomKademliaID(), "172.20.0.100:8000")

	updatedShortList := UpdateShortList(shortList, newContact, targetID)

	if len(updatedShortList) != k {
		t.Errorf("Expected %d contacts in shortlist, got %d", k, len(updatedShortList))
	}
}

func TestUpdateShortList_SortsByDistance(t *testing.T) {
	targetID := NewRandomKademliaID()
	contact1 := NewContact(NewRandomKademliaID(), "172.20.0.10:8000")
	contact2 := NewContact(NewRandomKademliaID(), "172.20.0.11:8000")
	shortList := []ShortListItem{
		{Contact: contact1, DistanceToTarget: contact1.ID.CalcDistance(targetID), Probed: false},
	}

	updatedShortList := UpdateShortList(shortList, contact2, targetID)

	if !updatedShortList[0].DistanceToTarget.Less(updatedShortList[1].DistanceToTarget) {
		t.Error("Expected contacts to be sorted by distance to target")
	}
}
func TestGetAllContactsFromShortList_ReturnsAllContacts(t *testing.T) {
	shortList := []ShortListItem{
		{Contact: NewContact(NewRandomKademliaID(), "172.20.0.10:8000")},
		{Contact: NewContact(NewRandomKademliaID(), "172.20.0.11:8000")},
	}

	contacts := GetAllContactsFromShortList(shortList)

	if len(contacts) != 2 {
		t.Errorf("Expected 2 contacts, got %d", len(contacts))
	}
}

func TestGetAllContactsFromShortList_EmptyShortList(t *testing.T) {
	shortList := []ShortListItem{}

	contacts := GetAllContactsFromShortList(shortList)

	if len(contacts) != 0 {
		t.Errorf("Expected 0 contacts, got %d", len(contacts))
	}
}

func TestGetAllContactsFromShortList_HandlesNilShortList(t *testing.T) {
	var shortList []ShortListItem

	contacts := GetAllContactsFromShortList(shortList)

	if len(contacts) != 0 {
		t.Errorf("Expected 0 contacts, got %d", len(contacts))
	}
}

func TestGetAllContactsFromShortList_HandlesSingleContact(t *testing.T) {
	shortList := []ShortListItem{
		{Contact: NewContact(NewRandomKademliaID(), "172.20.0.10:8000")},
	}

	contacts := GetAllContactsFromShortList(shortList)

	if len(contacts) != 1 {
		t.Errorf("Expected 1 contact, got %d", len(contacts))
	}
	if contacts[0].Address != "172.20.0.10:8000" {
		t.Errorf("Expected contact address 172.20.0.10:8000, got %s", contacts[0].Address)
	}
}
func TestGetAlphaNodes_ReturnsNotProbedContacts(t *testing.T) {
	rt := NewRoutingTable(NewContact(NewRandomKademliaID(), "172.20.0.1:8000"))
	kademlia := &Kademlia{RoutingTable: rt}
	shortList := []ShortListItem{
		{Contact: NewContact(NewRandomKademliaID(), "172.20.0.10:8000"), Probed: false},
		{Contact: NewContact(NewRandomKademliaID(), "172.20.0.11:8000"), Probed: true},
	}

	notProbed := kademlia.GetAlphaNodes(shortList)

	if len(notProbed) != 1 {
		t.Errorf("Expected 1 not probed contact, got %d", len(notProbed))
	}
}

func TestGetAlphaNodes_ExcludesSelfContact(t *testing.T) {
	me := NewContact(NewRandomKademliaID(), "172.20.0.1:8000")
	rt := NewRoutingTable(me)
	kademlia := &Kademlia{RoutingTable: rt}
	shortList := []ShortListItem{
		{Contact: me, Probed: false},
		{Contact: NewContact(NewRandomKademliaID(), "172.20.0.10:8000"), Probed: false},
	}

	notProbed := kademlia.GetAlphaNodes(shortList)

	if len(notProbed) != 1 {
		t.Errorf("Expected 1 not probed contact excluding self, got %d", len(notProbed))
	}
}

func TestGetAlphaNodes_ReturnsUpToAlphaContacts(t *testing.T) {
	rt := NewRoutingTable(NewContact(NewRandomKademliaID(), "172.20.0.1:8000"))
	kademlia := &Kademlia{RoutingTable: rt}
	shortList := []ShortListItem{
		{Contact: NewContact(NewRandomKademliaID(), "172.20.0.10:8000"), Probed: false},
		{Contact: NewContact(NewRandomKademliaID(), "172.20.0.11:8000"), Probed: false},
		{Contact: NewContact(NewRandomKademliaID(), "172.20.0.12:8000"), Probed: false},
		{Contact: NewContact(NewRandomKademliaID(), "172.20.0.13:8000"), Probed: false},
	}

	notProbed := kademlia.GetAlphaNodes(shortList)

	if len(notProbed) != alpha {
		t.Errorf("Expected %d not probed contacts, got %d", alpha, len(notProbed))
	}
}

func TestGetAlphaNodes_ReturnsAllIfLessThanAlpha(t *testing.T) {
	rt := NewRoutingTable(NewContact(NewRandomKademliaID(), "172.20.0.1:8000"))
	kademlia := &Kademlia{RoutingTable: rt}
	shortList := []ShortListItem{
		{Contact: NewContact(NewRandomKademliaID(), "172.20.0.10:8000"), Probed: false},
	}

	notProbed := kademlia.GetAlphaNodes(shortList)

	if len(notProbed) != 1 {
		t.Errorf("Expected 1 not probed contact, got %d", len(notProbed))
	}
}
func TestGetAlphaNodesFromKClosest_ReturnsNotProbedContacts(t *testing.T) {
	rt := NewRoutingTable(NewContact(NewRandomKademliaID(), "172.20.0.1:8000"))
	kademlia := &Kademlia{RoutingTable: rt}
	shortList := []ShortListItem{
		{Contact: NewContact(NewRandomKademliaID(), "172.20.0.10:8000"), Probed: true},
		{Contact: NewContact(NewRandomKademliaID(), "172.20.0.11:8000"), Probed: true},
	}
	target := NewContact(NewRandomKademliaID(), "172.20.0.12:8000")

	notProbed := kademlia.GetAlphaNodesFromKClosest(shortList, &target)

	if len(notProbed) != 0 {
		t.Errorf("Expected 0 not probed contacts, got %d", len(notProbed))
	}
}

func TestGetAlphaNodesFromKClosest_ExcludesShortListContacts(t *testing.T) {
	rt := NewRoutingTable(NewContact(NewRandomKademliaID(), "172.20.0.1:8000"))
	kademlia := &Kademlia{RoutingTable: rt}
	shortList := []ShortListItem{
		{Contact: NewContact(NewRandomKademliaID(), "172.20.0.10:8000"), Probed: false},
	}
	target := NewContact(NewRandomKademliaID(), "172.20.0.12:8000")

	notProbed := kademlia.GetAlphaNodesFromKClosest(shortList, &target)

	if len(notProbed) != 0 {
		t.Errorf("Expected 0 not probed contacts, got %d", len(notProbed))
	}
}
func TestCountProbedInShortList_ReturnsCorrectCount(t *testing.T) {
	shortList := []ShortListItem{
		{Probed: true},
		{Probed: false},
		{Probed: true},
	}

	count := CountProbedInShortList(shortList)

	if count != 2 {
		t.Errorf("Expected 2 probed contacts, got %d", count)
	}
}

func TestCountProbedInShortList_ReturnsZeroForEmptyList(t *testing.T) {
	shortList := []ShortListItem{}

	count := CountProbedInShortList(shortList)

	if count != 0 {
		t.Errorf("Expected 0 probed contacts, got %d", count)
	}
}

func TestCountProbedInShortList_ReturnsZeroWhenNoProbedContacts(t *testing.T) {
	shortList := []ShortListItem{
		{Probed: false},
		{Probed: false},
	}

	count := CountProbedInShortList(shortList)

	if count != 0 {
		t.Errorf("Expected 0 probed contacts, got %d", count)
	}
}

func TestCountProbedInShortList_ReturnsCountWhenAllContactsProbed(t *testing.T) {
	shortList := []ShortListItem{
		{Probed: true},
		{Probed: true},
		{Probed: true},
	}

	count := CountProbedInShortList(shortList)

	if count != 3 {
		t.Errorf("Expected 3 probed contacts, got %d", count)
	}
}

func TestListenActionChannel_PrintsAllIP(t *testing.T) {
	rt := NewRoutingTable(NewContact(NewRandomKademliaID(), "172.20.0.1:8000"))
	kademlia := &Kademlia{RoutingTable: rt, ActionChannel: make(chan Action, 1)}
	action := Action{Action: "PRINT"}
	kademlia.ActionChannel <- action

	go kademlia.ListenActionChannel()

}
