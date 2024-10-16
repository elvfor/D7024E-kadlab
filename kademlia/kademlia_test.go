package kademlia

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net"
	"testing"
	"time"
)

func TestKademlia_CreatesNewInstance(t *testing.T) {
	table := &RoutingTable{}
	conn := &net.UDPConn{}
	k := NewKademlia(table, conn)

	if k.RoutingTable != table {
		t.Errorf("Expected RoutingTable to be %v, got %v", table, k.RoutingTable)
	}
	if k.Network == nil {
		t.Error("Expected Network to be initialized, got nil")
	}
	if k.Data == nil {
		t.Error("Expected Data to be initialized, got nil")
	}
	if k.ActionChannel == nil {
		t.Error("Expected ActionChannel to be initialized, got nil")
	}
}

func TestKademlia_HandlesNilRoutingTable(t *testing.T) {
	conn := &net.UDPConn{}
	k := NewKademlia(nil, conn)

	if k.RoutingTable != nil {
		t.Errorf("Expected RoutingTable to be nil, got %v", k.RoutingTable)
	}
}

func TestKademlia_HandlesNilConnection(t *testing.T) {
	table := &RoutingTable{}
	k := NewKademlia(table, nil)

	if k.Network == nil {
		t.Error("Expected Network to be initialized, got nil")
	}
}

func TestLookupContact_ReturnsClosestContacts(t *testing.T) {
	target := NewContact(NewRandomKademliaID(), "172.20.0.10:8000")
	rt := NewRoutingTable(target)
	kademlia := &Kademlia{RoutingTable: rt}
	//add contact to rooouting table
	//contact := NewContact(NewRandomKademliaID(), "172.20.0.11:8000")
	kademlia.RoutingTable.AddContact(target)
	closestContacts := kademlia.LookupContact(&target)

	if len(closestContacts) != 1 {
		t.Errorf("Expected 1 closest contact, got %d", len(closestContacts))
	}
	if !closestContacts[0].ID.Equals(target.ID) {
		t.Errorf("Expected contact ID %s, got %s", target.ID.String(), closestContacts[0].ID.String())
	}
}

func TestLookupContact_ReturnsEmptyWhenNoContacts(t *testing.T) {
	target := NewContact(NewRandomKademliaID(), "172.20.0.10:8000")
	rt := NewRoutingTable(NewContact(NewRandomKademliaID(), "172.20.0.11:8000"))
	kademlia := &Kademlia{RoutingTable: rt}

	closestContacts := kademlia.LookupContact(&target)

	if len(closestContacts) != 0 {
		t.Errorf("Expected 0 closest contacts, got %d", len(closestContacts))
	}
}
func TestLookupData_ReturnsDataWhenExists(t *testing.T) {
	kademlia := &Kademlia{Data: &map[string][]byte{"hash1": []byte("data1")}}
	data, contacts := kademlia.LookupData("hash1")

	if data == nil || string(data) != "data1" {
		t.Errorf("Expected data 'data1', got %s", string(data))
	}
	if contacts != nil {
		t.Errorf("Expected contacts to be nil, got %v", contacts)
	}
}

func TestLookupData_ReturnsClosestContactsWhenDataNotExists(t *testing.T) {
	kademlia := &Kademlia{
		Data:         &map[string][]byte{},
		RoutingTable: NewRoutingTable(NewContact(NewRandomKademliaID(), "172.20.0.1:8000")),
	}
	hasher := sha1.New()
	hasher.Write([]byte("hash1"))
	hash := hasher.Sum(nil)
	hashString := hex.EncodeToString(hash)

	contact := NewContact(NewRandomKademliaID(), "172.20.0.2:8000")
	kademlia.RoutingTable.AddContact(contact)

	data, contacts := kademlia.LookupData(hashString)

	if data != nil {
		t.Errorf("Expected data to be nil, got %s", string(data))
	}

	target := NewContact(NewRandomKademliaID(), "172.20.0.1:8000")
	kademlia.RoutingTable.AddContact(target)
	if len(contacts) != 1 || !contacts[0].ID.Equals(contact.ID) {
		t.Errorf("Expected closest contact ID %s, got %v", contact.ID.String(), contacts)
	}
}

func TestLookupData_ReturnsEmptyContactsWhenNoClosestContacts(t *testing.T) {
	kademlia := &Kademlia{
		Data:         &map[string][]byte{},
		RoutingTable: NewRoutingTable(NewContact(NewRandomKademliaID(), "172.20.0.1:8000")),
	}
	hasher := sha1.New()
	hasher.Write([]byte("hash1"))
	hash := hasher.Sum(nil)
	hashString := hex.EncodeToString(hash)

	data, contacts := kademlia.LookupData(hashString)

	if data != nil {
		t.Errorf("Expected data to be nil, got %s", string(data))
	}
	if len(contacts) != 0 {
		t.Errorf("Expected 0 closest contacts, got %d", len(contacts))
	}
}

func TestStore_SavesDataCorrectly(t *testing.T) {
	kademlia := &Kademlia{Data: &map[string][]byte{}}
	hash := "hash1"
	data := []byte("data1")

	kademlia.Store(hash, data)

	if storedData, ok := (*kademlia.Data)[hash]; !ok || string(storedData) != "data1" {
		t.Errorf("Expected data 'data1' to be stored, got %s", string(storedData))
	}
}

func TestStore_OverwritesExistingData(t *testing.T) {
	kademlia := &Kademlia{Data: &map[string][]byte{"hash1": []byte("oldData")}}
	hash := "hash1"
	data := []byte("newData")

	kademlia.Store(hash, data)

	if storedData, ok := (*kademlia.Data)[hash]; !ok || string(storedData) != "newData" {
		t.Errorf("Expected data 'newData' to be stored, got %s", string(storedData))
	}
}

func TestStore_HandlesEmptyData(t *testing.T) {
	kademlia := &Kademlia{Data: &map[string][]byte{}}
	hash := "hash1"
	data := []byte("")

	kademlia.Store(hash, data)

	if storedData, ok := (*kademlia.Data)[hash]; !ok || string(storedData) != "" {
		t.Errorf("Expected empty data to be stored, got %s", string(storedData))
	}
}

func TestStore_HandlesNilData(t *testing.T) {
	kademlia := &Kademlia{Data: &map[string][]byte{}}
	hash := "hash1"
	var data []byte = nil

	kademlia.Store(hash, data)

	if storedData, ok := (*kademlia.Data)[hash]; !ok || storedData != nil {
		t.Errorf("Expected nil data to be stored, got %v", storedData)
	}
}
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
func TestGetAlphaNodesFromKClosest_AddsNewContacts(t *testing.T) {
	rt := NewRoutingTable(NewContact(NewRandomKademliaID(), "172.20.0.1:8000"))
	kademlia := &Kademlia{RoutingTable: rt}
	shortList := []ShortListItem{}
	target := NewContact(NewRandomKademliaID(), "172.20.0.12:8000")

	alphaContacts := []Contact{
		NewContact(NewRandomKademliaID(), "172.20.0.10:8000"),
		NewContact(NewRandomKademliaID(), "172.20.0.11:8000"),
	}

	for _, contact := range alphaContacts {
		kademlia.RoutingTable.AddContact(contact)
		shortList = append(shortList, ShortListItem{Contact: contact, Probed: false})
	}

	notProbed := kademlia.GetAlphaNodesFromKClosest(shortList, &target)

	if len(notProbed) != len(alphaContacts) {
		t.Errorf("Expected %d not probed contacts, got %d", len(alphaContacts), len(notProbed))
	}
}

func TestGetAlphaNodesFromKClosest_DoesNotAddExistingContacts(t *testing.T) {
	rt := NewRoutingTable(NewContact(NewRandomKademliaID(), "172.20.0.1:8000"))
	kademlia := &Kademlia{RoutingTable: rt}
	existingContact := NewContact(NewRandomKademliaID(), "172.20.0.10:8000")
	shortList := []ShortListItem{
		{Contact: existingContact, Probed: false},
	}
	target := NewContact(NewRandomKademliaID(), "172.20.0.12:8000")

	kademlia.RoutingTable.AddContact(existingContact)

	notProbed := kademlia.GetAlphaNodesFromKClosest(shortList, &target)

	if len(notProbed) != 0 {
		t.Errorf("Expected 0 not probed contacts, got %d", len(notProbed))
	}
}

func TestGetAlphaNodesFromKClosest_StopsAtAlpha(t *testing.T) {
	rt := NewRoutingTable(NewContact(NewRandomKademliaID(), "172.20.0.1:8000"))
	kademlia := &Kademlia{RoutingTable: rt}
	shortList := []ShortListItem{}
	target := NewContact(NewRandomKademliaID(), "172.20.0.12:8000")

	for i := 0; i < alpha+1; i++ {
		contact := NewContact(NewRandomKademliaID(), fmt.Sprintf("172.20.0.%d:8000", i))
		kademlia.RoutingTable.AddContact(contact)
		shortList = append(shortList, ShortListItem{Contact: contact, Probed: false})
	}

	notProbed := kademlia.GetAlphaNodesFromKClosest(shortList, &target)

	if len(notProbed) != alpha {
		t.Errorf("Expected %d not probed contacts, got %d", alpha, len(notProbed))
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
func TestListenActionChannel_UpdatesRT(t *testing.T) {
	rt := NewRoutingTable(NewContact(NewRandomKademliaID(), "172.20.0.1:8000"))
	kademlia := &Kademlia{RoutingTable: rt, ActionChannel: make(chan Action, 1)}
	action := Action{Action: "UpdateRT", SenderId: NewRandomKademliaID(), SenderIp: "172.20.0.2:8000"}
	go kademlia.ListenActionChannel()
	kademlia.ActionChannel <- action

	time.Sleep(1 * time.Second)

	contacts := kademlia.RoutingTable.FindClosestContacts(action.SenderId, 1)
	if len(contacts) == 0 || !contacts[0].ID.Equals(action.SenderId) {
		t.Error("Expected contact to be added to routing table")
	}
}

func TestListenActionChannel_StoresData(t *testing.T) {
	kademlia := &Kademlia{Data: &map[string][]byte{}, ActionChannel: make(chan Action, 1)}
	action := Action{Action: "Store", Hash: "hash1", Data: []byte("data1")}

	go kademlia.ListenActionChannel()
	kademlia.ActionChannel <- action
	time.Sleep(1 * time.Second)

	if storedData, ok := (*kademlia.Data)[action.Hash]; !ok || string(storedData) != "data1" {
		t.Errorf("Expected data 'data1' to be stored, got %s", string(storedData))
	}
}
func TestListenActionChannel_LookupContact(t *testing.T) {
	me := NewContact(NewRandomKademliaID(), "172.20.0.10:8000")
	rt := NewRoutingTable(me)
	conn := &net.UDPConn{}
	kademlia := NewKademlia(rt, conn)
	kademlia.Network = &Network{responseChan: make(chan Response, 1)}
	target := NewContact(NewRandomKademliaID(), "172.20.11:8000")
	kademlia.RoutingTable.AddContact(target)
	action := Action{Action: "LookupContact", Target: &target}
	go kademlia.ListenActionChannel()
	kademlia.ActionChannel <- action
	time.Sleep(1 * time.Second)
	response := <-kademlia.Network.responseChan
	fmt.Print(response, "response")
	if len(response.ClosestContacts) != 1 || !response.ClosestContacts[0].ID.Equals(target.ID) {
		t.Errorf("Expected contact ID %s, got %v", target.ID.String(), response.ClosestContacts)
	}
}
func TestListenActionChannel_LookupData(t *testing.T) {
	hasher := sha1.New()
	hasher.Write([]byte("hash1"))
	hash := hasher.Sum(nil)
	hashString := hex.EncodeToString(hash)
	kademlia := &Kademlia{Data: &map[string][]byte{hashString: []byte("data1")}, ActionChannel: make(chan Action, 1)}
	kademlia.Network = &Network{responseChan: make(chan Response, 1)}
	action := Action{Action: "LookupData", Hash: hashString}
	go kademlia.ListenActionChannel()
	kademlia.ActionChannel <- action
	time.Sleep(1 * time.Second)

	response := <-kademlia.Network.responseChan
	if response.Data == nil || string(response.Data) != "data1" {
		t.Errorf("Expected data 'data1', got %s", string(response.Data))
	}
	if response.ClosestContacts != nil {
		t.Errorf("Expected contacts to be nil, got %v", response.ClosestContacts)
	}
}
func TestCloseChannels_ClosesAllChannels(t *testing.T) {
	contactsChan := make(chan Contact, 1)
	dataChan := make(chan []byte, 1)
	contactChanFoundDataOn := make(chan Contact, 1)

	closeChannels(contactsChan, dataChan, contactChanFoundDataOn)

	select {
	case _, ok := <-contactsChan:
		if ok {
			t.Error("Expected contactsChan to be closed")
		}
	default:
		t.Error("Expected contactsChan to be closed")
	}

	select {
	case _, ok := <-dataChan:
		if ok {
			t.Error("Expected dataChan to be closed")
		}
	default:
		t.Error("Expected dataChan to be closed")
	}

	select {
	case _, ok := <-contactChanFoundDataOn:
		if ok {
			t.Error("Expected contactChanFoundDataOn to be closed")
		}
	default:
		t.Error("Expected contactChanFoundDataOn to be closed")
	}
}

func TestCloseChannels_HandlesNilChannels(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Expected no panic, but got %v", r)
		}
	}()

	closeChannels(nil, nil, nil)
}

func TestHandleFoundData_ReturnsContactAndData(t *testing.T) {
	dataChan := make(chan []byte, 1)
	contactChan := make(chan Contact, 1)
	expectedData := []byte("data")
	expectedContact := Contact{ID: NewRandomKademliaID(), Address: "172.20.0.10:8000"}

	dataChan <- expectedData
	contactChan <- expectedContact

	contact, data := handleFoundData(dataChan, contactChan)

	if !contact.ID.Equals(expectedContact.ID) {
		t.Errorf("Expected contact ID %s, got %s", expectedContact.ID.String(), contact.ID.String())
	}
	if string(data) != string(expectedData) {
		t.Errorf("Expected data %s, got %s", string(expectedData), string(data))
	}
}

func TestHandleFoundData_ReturnsEmptyContactAndNilDataWhenContactIDIsNil(t *testing.T) {
	dataChan := make(chan []byte, 1)
	contactChan := make(chan Contact, 1)
	expectedData := []byte("data")
	expectedContact := Contact{ID: nil, Address: "172.20.0.10:8000"}

	dataChan <- expectedData
	contactChan <- expectedContact

	contact, data := handleFoundData(dataChan, contactChan)

	if contact.ID != nil {
		t.Errorf("Expected empty contact, got %s", contact.ID.String())
	}
	if data != nil {
		t.Errorf("Expected nil data, got %s", string(data))
	}
}
func TestUpdateShortListWithContacts_AddsNewContacts(t *testing.T) {
	shortList := []ShortListItem{}
	target := NewContact(NewRandomKademliaID(), "172.20.0.1:8000")
	contactsChan := make(chan Contact, 1)
	newContact := NewContact(NewRandomKademliaID(), "172.20.0.2:8000")

	go func() {
		contactsChan <- newContact
		close(contactsChan)
	}()
	kademlia := &Kademlia{ActionChannel: make(chan Action, 1)}
	kademlia.Network = &Network{responseChan: make(chan Response, 1)}
	updatedShortList := kademlia.updateShortListWithContacts(shortList, &target, contactsChan)

	if len(updatedShortList) != 1 {
		t.Errorf("Expected 1 contact in shortlist, got %d", len(updatedShortList))
	}
	if !updatedShortList[0].Contact.ID.Equals(newContact.ID) {
		t.Errorf("Expected contact ID %s, got %s", newContact.ID.String(), updatedShortList[0].Contact.ID.String())
	}
}

func TestUpdateShortListWithContacts_HandlesEmptyChannel(t *testing.T) {
	shortList := []ShortListItem{}
	target := NewContact(NewRandomKademliaID(), "172.20.0.1:8000")
	contactsChan := make(chan Contact)

	close(contactsChan)
	kademlia := &Kademlia{ActionChannel: make(chan Action, 1)}
	kademlia.Network = &Network{responseChan: make(chan Response, 1)}
	updatedShortList := kademlia.updateShortListWithContacts(shortList, &target, contactsChan)

	if len(updatedShortList) != 0 {
		t.Errorf("Expected 0 contacts in shortlist, got %d", len(updatedShortList))
	}
}

func TestUpdateShortListWithContacts_UpdatesExistingContacts(t *testing.T) {
	target := NewContact(NewRandomKademliaID(), "172.20.0.1:8000")
	existingContact := NewContact(NewRandomKademliaID(), "172.20.0.2:8000")
	shortList := []ShortListItem{
		{Contact: existingContact, DistanceToTarget: existingContact.ID.CalcDistance(target.ID), Probed: false},
	}
	contactsChan := make(chan Contact, 1)
	updatedContact := NewContact(existingContact.ID, "172.20.0.3:8000")

	go func() {
		contactsChan <- updatedContact
		close(contactsChan)
	}()
	kademlia := &Kademlia{ActionChannel: make(chan Action, 1)}
	kademlia.Network = &Network{responseChan: make(chan Response, 1)}
	updatedShortList := kademlia.updateShortListWithContacts(shortList, &target, contactsChan)

	if len(updatedShortList) != 1 {
		t.Errorf("Expected 1 contact in shortlist, got %d", len(updatedShortList))
	}
	if updatedShortList[0].Contact.Address != updatedContact.Address {
		t.Errorf("Expected contact address %s, got %s", updatedContact.Address, updatedShortList[0].Contact.Address)
	}
}
func TestMarkProbedContacts_UpdatesProbedStatus(t *testing.T) {
	shortList := []ShortListItem{
		{Contact: NewContact(NewRandomKademliaID(), "172.20.0.10:8000"), Probed: false},
		{Contact: NewContact(NewRandomKademliaID(), "172.20.0.11:8000"), Probed: false},
	}
	notProbed := []ShortListItem{
		{Contact: shortList[0].Contact},
	}

	updatedShortList := markProbedContacts(shortList, notProbed)

	if !updatedShortList[0].Probed {
		t.Errorf("Expected contact %s to be probed", updatedShortList[0].Contact.ID.String())
	}
	if updatedShortList[1].Probed {
		t.Errorf("Expected contact %s to not be probed", updatedShortList[1].Contact.ID.String())
	}
}

func TestMarkProbedContacts_HandlesEmptyShortList(t *testing.T) {
	shortList := []ShortListItem{}
	notProbed := []ShortListItem{
		{Contact: NewContact(NewRandomKademliaID(), "172.20.0.10:8000")},
	}

	updatedShortList := markProbedContacts(shortList, notProbed)

	if len(updatedShortList) != 0 {
		t.Errorf("Expected empty shortlist, got %d", len(updatedShortList))
	}
}

func TestMarkProbedContacts_HandlesEmptyNotProbedList(t *testing.T) {
	shortList := []ShortListItem{
		{Contact: NewContact(NewRandomKademliaID(), "172.20.0.10:8000"), Probed: false},
	}
	notProbed := []ShortListItem{}

	updatedShortList := markProbedContacts(shortList, notProbed)

	if updatedShortList[0].Probed {
		t.Errorf("Expected contact %s to not be probed", updatedShortList[0].Contact.ID.String())
	}
}

func TestMarkProbedContacts_HandlesNilShortList(t *testing.T) {
	var shortList []ShortListItem
	notProbed := []ShortListItem{
		{Contact: NewContact(NewRandomKademliaID(), "172.20.0.10:8000")},
	}

	updatedShortList := markProbedContacts(shortList, notProbed)

	if len(updatedShortList) != 0 {
		t.Errorf("Expected empty shortlist, got %d", len(updatedShortList))
	}
}

func TestMarkProbedContacts_HandlesNilNotProbedList(t *testing.T) {
	shortList := []ShortListItem{
		{Contact: NewContact(NewRandomKademliaID(), "172.20.0.10:8000"), Probed: false},
	}
	var notProbed []ShortListItem

	updatedShortList := markProbedContacts(shortList, notProbed)

	if updatedShortList[0].Probed {
		t.Errorf("Expected contact %s to not be probed", updatedShortList[0].Contact.ID.String())
	}
}
