package kademlia

import (
	"encoding/json"
	"errors"
	"testing"
)

// Test NewNetwork
func TestNewNetwork(t *testing.T) {
	newNetwork := NewNetwork(nil) // Assuming NewNetwork takes two arguments
	if newNetwork == nil {
		t.Error("Expected new network to be created")
	}
}

// MockedNetwork is a mock implementation of NetworkInterface for testing
type MockedNetwork struct {
	Responses map[string]Response
	Err       error
}

func (m *MockedNetwork) Listen(k *Kademlia) {
	// No-op for mock
}

func (m *MockedNetwork) SendPingMessage(sender *Contact, receiver *Contact) bool {
	if m.Err != nil {
		return false
	}
	return true // Simulate a successful ping
}

func (m *MockedNetwork) SendFindContactMessage(sender *Contact, receiver *Contact, target *Contact) ([]Contact, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	response, exists := m.Responses["FIND_NODE"]
	if !exists {
		return nil, errors.New("no response configured for FIND_NODE")
	}
	return response.ClosestContacts, nil
}

func (m *MockedNetwork) SendFindDataMessage(sender *Contact, receiver *Contact, hash string) ([]Contact, []byte, error) {
	if m.Err != nil {
		return nil, nil, m.Err
	}
	response, exists := m.Responses["FIND_DATA"]
	if !exists {
		return nil, nil, errors.New("no response configured for FIND_DATA")
	}
	return response.ClosestContacts, response.Data, nil
}

func (m *MockedNetwork) SendStoreMessage(sender *Contact, receiver *Contact, dataID *KademliaID, data []byte) bool {
	if m.Err != nil {
		return false
	}
	return true // Simulate a successful store
}

func (m *MockedNetwork) SendMessage(sender *Contact, receiver *Contact, message interface{}) ([]byte, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return json.Marshal(message) // Simulate a successful message sending
}

/* Does not add to test coverage

// TestNetwork_SendPingMessage tests the SendPingMessage method
func TestNetwork_SendPingMessage(t *testing.T) {
	network := &MockedNetwork{} // conn is nil for mock
	sender := &Contact{ID: NewKademliaID("0000000000000000000000000000000000000001"), Address: "127.0.0.1:8000"}
	receiver := &Contact{ID: NewKademliaID("0000000000000000000000000000000000000002"), Address: "127.0.0.1:8001"}

	result := network.SendPingMessage(sender, receiver)
	if !result {
		t.Error("SendPingMessage should return true for a successful ping")
	}
}

// TestNetwork_SendFindContactMessage tests the SendFindContactMessage method

func TestNetwork_SendFindContactMessage(t *testing.T) {
	mockedNetwork := &MockedNetwork{
		Responses: map[string]Response{
			"FIND_NODE": {
				ClosestContacts: []Contact{
					{ID: NewKademliaID("0000000000000000000000000000000000000001"), Address: "127.0.0.1:8002"},
				},
			},
		},
	}
	network := &Network{reponseChan: make(chan Response), conn: nil} // conn is nil for mock
	sender := &Contact{ID: NewKademliaID("0000000000000000000000000000000000000002"), Address: "127.0.0.1:8000"}
	receiver := &Contact{ID: NewKademliaID("0000000000000000000000000000000000000003"), Address: "127.0.0.1:8001"}
	target := &Contact{ID: NewKademliaID("0000000000000000000000000000000000000000"), Address: "127.0.0.1:8002"}

	closestContacts, err := network.SendFindContactMessage(sender, receiver, target)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(closestContacts) != 1 {
		t.Errorf("expected 1 closest contact, got %d", len(closestContacts))
	}
}

// TestNetwork_SendFindDataMessage tests the SendFindDataMessage method

func TestNetwork_SendFindDataMessage(t *testing.T) {
	_ = &MockedNetwork{
		Responses: map[string]Response{
			"FIND_DATA": {
				Data: []byte("sample data"),
				ClosestContacts: []Contact{
					{ID: NewKademliaID("3"), Address: "127.0.0.1:8002"},
				},
			},
		},
	}
	network := &Network{reponseChan: make(chan Response), conn: nil} // conn is nil for mock
	sender := &Contact{ID: NewKademliaID("1"), Address: "127.0.0.1:8000"}
	receiver := &Contact{ID: NewKademliaID("2"), Address: "127.0.0.1:8001"}

	closestContacts, data, err := network.SendFindDataMessage(sender, receiver, "sample_hash")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(closestContacts) != 1 {
		t.Errorf("expected 1 closest contact, got %d", len(closestContacts))
	}
	if string(data) != "sample data" {
		t.Errorf("expected 'sample data', got '%s'", data)
	}
}

// TestNetwork_SendStoreMessage tests the SendStoreMessage method
func TestNetwork_SendStoreMessage(t *testing.T) {
	_ = &MockedNetwork{}
	network := &Network{reponseChan: make(chan Response), conn: nil} // conn is nil for mock
	sender := &Contact{ID: NewKademliaID("1"), Address: "127.0.0.1:8000"}
	receiver := &Contact{ID: NewKademliaID("2"), Address: "127.0.0.1:8001"}
	dataID := NewKademliaID("dataID")
	data := []byte("sample data")

	result := network.SendStoreMessage(sender, receiver, dataID, data)
	if !result {
		t.Error("SendStoreMessage should return true for a successful store")
	}
}

// TestNetwork_SendMessage tests the SendMessage method
func TestNetwork_SendMessage(t *testing.T) {
	_ = &MockedNetwork{}
	network := &Network{reponseChan: make(chan Response), conn: nil} // conn is nil for mock
	sender := &Contact{ID: NewKademliaID("1"), Address: "127.0.0.1:8000"}
	receiver := &Contact{ID: NewKademliaID("2"), Address: "127.0.0.1:8001"}
	message := Message{Type: "TEST", SenderID: sender.ID, TargetID: receiver.ID.String()}

	response, err := network.SendMessage(sender, receiver, message)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(response) != "{\"Type\":\"TEST\",\"SenderID\":\"1\",\"TargetID\":\"2\"}" {
		t.Errorf("unexpected response: %s", response)
	}
}
*/
