package main

import (
	"testing"
)

func TestJoinNetwork_ReturnsNonNilKademliaInstance(t *testing.T) {
	k, _ := JoinNetwork("172.20.0.1", "8000")
	if k == nil {
		t.Fatal("Expected non-nil Kademlia instance")
	}
}

func TestJoinNetwork_InitializesRoutingTable(t *testing.T) {
	k, _ := JoinNetwork("172.20.0.1", "8001")
	if k.RoutingTable == nil {
		t.Fatal("Expected non-nil RoutingTable")
	}
}

func TestJoinNetwork_InitializesNetwork(t *testing.T) {
	k, _ := JoinNetwork("172.20.0.1", "8002")
	if k.Network == nil {
		t.Fatal("Expected non-nil Network")
	}
}

func TestJoinNetwork_HandlesPortInUse(t *testing.T) {
	_, err := JoinNetwork("172.20.0.1", "8003")
	if err != nil {
		t.Fatal("Expected no error on first call to JoinNetwork")
	}

	_, err = JoinNetwork("172.20.0.1", "8003")
	if err == nil {
		t.Fatal("Expected error due to port in use")
	}
}

func TestJoinNetworkBootstrap_ReturnsNonNilKademliaInstance(t *testing.T) {
	k, err := JoinNetworkBootstrap("172.20.0.1", "8004")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if k == nil {
		t.Fatal("Expected non-nil Kademlia instance")
	}
}

func TestJoinNetworkBootstrap_InitializesRoutingTable(t *testing.T) {
	k, err := JoinNetworkBootstrap("172.20.0.1", "8005")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if k.RoutingTable == nil {
		t.Fatal("Expected non-nil RoutingTable")
	}
}

func TestJoinNetworkBootstrap_InitializesNetwork(t *testing.T) {
	k, err := JoinNetworkBootstrap("172.20.0.1", "8006")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if k.Network == nil {
		t.Fatal("Expected non-nil Network")
	}
}

func TestJoinNetworkBootstrap_HandlesPortInUse(t *testing.T) {
	_, err := JoinNetworkBootstrap("172.20.0.1", "8007")
	if err != nil {
		t.Fatal("Expected no error on first call to JoinNetworkBootstrap")
	}

	_, err = JoinNetworkBootstrap("172.20.0.1", "8007")
	if err == nil {
		t.Fatal("Expected error due to port in use")
	}
}
func TestReturnsLocalIPAddress(t *testing.T) {
	ip, _ := GetOutboundIP()
	if ip == nil {
		t.Fatal("Expected non-nil IP address")
	}
}

func TestReturnsValidIPAddress(t *testing.T) {
	ip, _ := GetOutboundIP()
	if ip.To4() == nil {
		t.Fatal("Expected a valid IPv4 address")
	}
}
