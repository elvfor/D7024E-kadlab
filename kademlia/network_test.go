package kademlia

import (
	"testing"
)

// Test NewNetwork
func TestNewNetwork(t *testing.T) {
	newNetwork := NewNetwork(nil) // Assuming NewNetwork takes two arguments
	if newNetwork == nil {
		t.Error("Expected new network to be created")
	}
}
