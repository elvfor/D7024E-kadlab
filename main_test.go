package main

import (
	"d7024e/kademlia"
	"testing"
	"time"
)

/*
	func TestJoinNetwork_ReturnsInitializedKademliaInstance(t *testing.T) {
		k := JoinNetwork("172.20.0.1:8000")
		if k == nil {
			t.Fatal("Expected non-nil Kademlia instance")
		}
		if k.RoutingTable == nil {
			t.Fatal("Expected non-nil RoutingTable")
		}
		if k.Network == nil {
			t.Fatal("Expected non-nil Network")
		}
	}
*/

func TestDoLookUpOnSelf_UpdatesRoutingTable(t *testing.T) {
	k := &kademlia.Kademlia{
		RoutingTable:  kademlia.NewRoutingTable(kademlia.NewContact(kademlia.NewRandomKademliaID(), "172.20.0.1:8000")),
		ActionChannel: make(chan kademlia.Action, 1),
	}

	DoLookUpOnSelf(k)

	select {
	case action := <-k.ActionChannel:
		if action.Action != "UpdateRT" {
			t.Errorf("Expected action 'UpdateRT', got %s", action.Action)
		}
		if action.SenderIp != "172.20.0.2:8000" {
			t.Errorf("Expected sender IP '172.20.0.2:8000', got %s", action.SenderIp)
		}
	case <-time.After(1 * time.Second):
		t.Error("Expected action but got timeout")
	}
}

func DoLookUpOnSelf_NoContactsFound(t *testing.T) {
	k := &kademlia.Kademlia{
		RoutingTable:  kademlia.NewRoutingTable(kademlia.NewContact(kademlia.NewRandomKademliaID(), "172.20.0.1:8000")),
		ActionChannel: make(chan kademlia.Action, 1),
	}

	DoLookUpOnSelf(k)

	select {
	case <-k.ActionChannel:
		t.Error("Expected no action but got one")
	case <-time.After(1 * time.Second):
		// Expected timeout
	}
}

func DoLookUpOnSelf_MultipleContactsFound(t *testing.T) {
	k := &kademlia.Kademlia{
		RoutingTable:  kademlia.NewRoutingTable(kademlia.NewContact(kademlia.NewRandomKademliaID(), "172.20.0.1:8000")),
		ActionChannel: make(chan kademlia.Action, 2),
	}

	DoLookUpOnSelf(k)

	actions := []kademlia.Action{}
	timeout := time.After(1 * time.Second)
	for i := 0; i < 2; i++ {
		select {
		case action := <-k.ActionChannel:
			actions = append(actions, action)
		case <-timeout:
			t.Error("Expected actions but got timeout")
		}
	}

	if len(actions) != 2 {
		t.Errorf("Expected 2 actions, got %d", len(actions))
	}
}

/*func TestJoinNetwork_HandlesInvalidIP(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for invalid IP")
		}
	}()
	JoinNetwork("invalid-ip")
}

func TestJoinNetwork_AddsBootstrapContact(t *testing.T) {
	k := JoinNetwork("172.20.0.1:8000")
	if len(k.RoutingTable.GetContacts()) == 0 {
		t.Fatal("Expected at least one contact in the routing table")
	}
	bootstrapContact := k.RoutingTable.GetContacts()[0]
	if bootstrapContact.Address != "172.20.0.6:8000" {
		t.Errorf("Expected bootstrap contact address '172.20.0.6:8000', got %s", bootstrapContact.Address)
	}
}

func TestJoinNetwork_HandlesListenPacketError(t *testing.T) {
	originalListenPacket := net.ListenPacket
	defer func() { net.ListenPacket = originalListenPacket }()
	netListenPacket = func(network, address string) (net.PacketConn, error) {
		return nil, fmt.Errorf("mock error")
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic due to listen packet error")
		}
	}()
	JoinNetwork("172.20.0.1:8000")
}*/
