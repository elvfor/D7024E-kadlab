package kademlia

import (
	"testing"
)

// TestNewContact tests the creation of a new Contact
func TestNewContact(t *testing.T) {
	id := NewKademliaID("FFFFFFFF00000000000000000000000000000000")
	address := "127.0.0.1:8000"
	contact := NewContact(id, address)

	if contact.ID.String() != id.String() {
		t.Errorf("Expected ID %s, got %s", id, contact.ID)
	}
	if contact.Address != address {
		t.Errorf("Expected address %s, got %s", address, contact.Address)
	}
}

// TestCalcDistance tests calculating the distance between two contacts
func TestCalcDistance(t *testing.T) {
	id1 := NewKademliaID("FFFFFFFF00000000000000000000000000000000")
	id2 := NewKademliaID("0000000000000000000000000000000000000001")
	contact1 := NewContact(id1, "127.0.0.1:8000")
	contact2 := NewContact(id2, "127.0.0.2:8000")

	contact1.CalcDistance(id2)
	expectedDistance := id1.CalcDistance(id2)

	if !contact1.distance.Equals(expectedDistance) {
		t.Errorf("Expected distance %s, got %s", expectedDistance, contact1.distance)
	}

	contact2.CalcDistance(id1)
	expectedDistance2 := id2.CalcDistance(id1)

	if !contact2.distance.Equals(expectedDistance2) {
		t.Errorf("Expected distance %s, got %s", expectedDistance2, contact2.distance)
	}
}

// TestLess tests the comparison between two contacts based on distance
func TestLess(t *testing.T) {
	contact1 := Contact{ID: NewKademliaID("abcdefabcdefabcdefabcdefabcdefabcdefabcd"), Address: "192.168.1.1"}
	contact2 := Contact{ID: NewKademliaID("abcdefabcdefabcdefabcdefabcdefabcdefabcc"), Address: "192.168.1.2"}
	target := NewKademliaID("0000000000000000000000000000000000000001")

	contact1.CalcDistance(target)
	contact2.CalcDistance(target)

	// Expect contact1 to be less (closer) than contact2
	if !contact1.Less(&contact2) {
		t.Errorf("Expected contact1 to be less than contact2")
	}

	// Test when both KademliaID instances are equal
	equalID1 := NewKademliaID("abcdefabcdefabcdefabcdefabcdefabcdefabcd")
	equalID2 := NewKademliaID("abcdefabcdefabcdefabcdefabcdefabcdefabcd")

	if equalID1.Less(equalID2) {
		t.Errorf("Expected equal KademliaID instances to return false for Less")
	}
}

func TestKademliaID_Less(t *testing.T) {
	// Test when kademliaID is less than otherKademliaID
	id1 := NewKademliaID("0000000000000000000000000000000000000001")
	id2 := NewKademliaID("0000000000000000000000000000000000000002")
	if !id1.Less(id2) {
		t.Errorf("Expected id1 to be less than id2")
	}

	// Test when kademliaID is greater than otherKademliaID
	if id2.Less(id1) {
		t.Errorf("Expected id2 to not be less than id1")
	}

	// Test when both KademliaID instances are equal
	equalID1 := NewKademliaID("abcdefabcdefabcdefabcdefabcdefabcdefabcd")
	equalID2 := NewKademliaID("abcdefabcdefabcdefabcdefabcdefabcdefabcd")
	if equalID1.Less(equalID2) {
		t.Errorf("Expected equal KademliaID instances to return false for Less")
	}
}

// TestContactCandidatesAppend tests the appending of contacts to candidates
func TestContactCandidatesAppend(t *testing.T) {
	id1 := NewKademliaID("FFFFFFFF00000000000000000000000000000000")
	id2 := NewKademliaID("0000000000000000000000000000000000000001")
	contact1 := NewContact(id1, "127.0.0.1:8000")
	contact2 := NewContact(id2, "127.0.0.2:8000")

	candidates := &ContactCandidates{}
	candidates.Append([]Contact{contact1, contact2})

	if len(candidates.contacts) != 2 {
		t.Errorf("Expected 2 contacts, got %d", len(candidates.contacts))
	}
}

// TestContactCandidatesGetContacts tests retrieving the first N contacts
func TestContactCandidatesGetContacts(t *testing.T) {
	id1 := NewKademliaID("FFFFFFFF00000000000000000000000000000000")
	id2 := NewKademliaID("0000000000000000000000000000000000000001")
	contact1 := NewContact(id1, "127.0.0.1:8000")
	contact2 := NewContact(id2, "127.0.0.2:8000")

	candidates := &ContactCandidates{}
	candidates.Append([]Contact{contact1, contact2})

	contacts := candidates.GetContacts(1)
	if len(contacts) != 1 {
		t.Errorf("Expected 1 contact, got %d", len(contacts))
	}
	if contacts[0].ID.String() != contact1.ID.String() {
		t.Errorf("Expected contact ID %s, got %s", contact1.ID, contacts[0].ID)
	}
}

// TestContactCandidatesSort tests sorting contacts by distance
func TestContactCandidatesSort(t *testing.T) {
	target := NewKademliaID("0000000000000000000000000000000000000001")
	id1 := NewKademliaID("abcdefabcdefabcdefabcdefabcdefabcdefabcd")
	id2 := NewKademliaID("abcdefabcdefabcdefabcdefabcdefabcdefabcc")

	contact1 := NewContact(id1, "127.0.0.1:8000")
	contact2 := NewContact(id2, "127.0.0.2:8000")

	contact1.CalcDistance(target)
	contact2.CalcDistance(target)

	candidates := &ContactCandidates{}
	candidates.Append([]Contact{contact1, contact2})

	candidates.Sort()

	if !candidates.contacts[0].ID.Equals(id1) {
		t.Errorf("Expected contact1 to be first after sorting")
	}
}

// TestContactCandidatesLen tests the length of contact candidates
func TestContactCandidatesLen(t *testing.T) {
	id1 := NewKademliaID("FFFFFFFF00000000000000000000000000000000")
	id2 := NewKademliaID("0000000000000000000000000000000000000001")
	contact1 := NewContact(id1, "127.0.0.1:8000")
	contact2 := NewContact(id2, "127.0.0.2:8000")

	candidates := &ContactCandidates{}
	candidates.Append([]Contact{contact1, contact2})

	if candidates.Len() != 2 {
		t.Errorf("Expected length 2, got %d", candidates.Len())
	}
}

// TestContactCandidatesSwap tests swapping contacts
func TestContactCandidatesSwap(t *testing.T) {
	id1 := NewKademliaID("FFFFFFFF00000000000000000000000000000000")
	id2 := NewKademliaID("0000000000000000000000000000000000000001")
	contact1 := NewContact(id1, "127.0.0.1:8000")
	contact2 := NewContact(id2, "127.0.0.2:8000")

	candidates := &ContactCandidates{}
	candidates.Append([]Contact{contact1, contact2})

	candidates.Swap(0, 1)

	if candidates.contacts[0].ID.String() != id2.String() {
		t.Errorf("Expected contact2 to be first after swap")
	}
}
