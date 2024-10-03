package kademlia

import (
	"encoding/json"
	"testing"
)

func TestMarshalContactList(t *testing.T) {
	id := NewRandomKademliaID()
	contact := NewContact(id, "172.20.0.6:8000")
	contact.CalcDistance(id)

	// Create a slice of Contact objects
	contactList := []Contact{contact}

	// Marshal the slice of Contact objects
	data, err := json.Marshal(contactList)
	if err != nil {
		t.Errorf("Error marshalling contact list: %v", err)
	}

	// Unmarshal the JSON data back into a slice of Contact objects
	var unmarshalledContactList []Contact
	err = json.Unmarshal(data, &unmarshalledContactList)
	if err != nil {
		t.Errorf("Error unmarshalling contact list: %v", err)
	}

	// Verify the unmarshalled data
	if contact.ID.String() != unmarshalledContactList[0].ID.String() {
		t.Errorf("Expected %s, got %s", contact.ID.String(), unmarshalledContactList[0].ID.String())
	}
}
