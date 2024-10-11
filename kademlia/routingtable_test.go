package kademlia

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestAddContactRT(t *testing.T) {
	rt := NewRoutingTable(
		NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000"),
	)

	// Add some contacts
	contact1 := NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000001"), "localhost:8001")
	contact2 := NewContact(NewKademliaID("1111111100000000000000000000000000000001"), "localhost:8002")

	bucketIsFull, _ := rt.AddContact(contact1)
	if bucketIsFull {
		t.Fatalf("Bucket should not be full after adding the first contact")
	}

	bucketIsFull, _ = rt.AddContact(contact2)
	if bucketIsFull {
		t.Fatalf("Bucket should not be full after adding the second contact")
	}
}

func TestRemoveContactRT(t *testing.T) {
	rt := NewRoutingTable(
		NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000"),
	)

	contact1 := NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000001"), "localhost:8001")
	rt.AddContact(contact1)

	// Remove the contact and ensure it's no longer in the routing table
	rt.RemoveContact(&contact1)
	closestContacts := rt.FindClosestContacts(contact1.ID, 1)
	if len(closestContacts) != 0 {
		t.Fatalf("Contact should have been removed, but was still found")
	}
}

func TestFindClosestContacts(t *testing.T) {
	// Create a new routing table with a local contact
	rt := NewRoutingTable(
		NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000"),
	)

	// Add multiple contacts to different buckets
	rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000001"), "localhost:8001"))
	rt.AddContact(NewContact(NewKademliaID("1111111100000000000000000000000000000001"), "localhost:8002"))
	rt.AddContact(NewContact(NewKademliaID("1111111200000000000000000000000000000002"), "localhost:8003"))
	rt.AddContact(NewContact(NewKademliaID("1111111300000000000000000000000000000003"), "localhost:8004"))
	rt.AddContact(NewContact(NewKademliaID("1111111400000000000000000000000000000004"), "localhost:8005"))
	rt.AddContact(NewContact(NewKademliaID("1111111500000000000000000000000000000005"), "localhost:8006"))

	targetID := NewKademliaID("1111111600000000000000000000000000000006")
	kClosest := 3

	closestContacts := rt.FindClosestContacts(targetID, kClosest)

	expectedContacts := []Contact{
		NewContact(NewKademliaID("1111111500000000000000000000000000000005"), "localhost:8006"),
		NewContact(NewKademliaID("1111111400000000000000000000000000000004"), "localhost:8005"),
		NewContact(NewKademliaID("1111111200000000000000000000000000000002"), "localhost:8003"),
	}

	for _, expected := range expectedContacts {
		found := false
		for _, contact := range closestContacts {
			if contact.ID.Equals(expected.ID) && contact.Address == expected.Address {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("Expected contact %s not found in closestContacts", expected.String())
		}
	}
}

func TestFewerContactsThanK(t *testing.T) {
	rt := NewRoutingTable(
		NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000"),
	)

	// Add multiple contacts to different buckets
	rt.AddContact(NewContact(NewKademliaID("1111111300000000000000000000000000000003"), "localhost:8004"))
	rt.AddContact(NewContact(NewKademliaID("1111111400000000000000000000000000000004"), "localhost:8005"))
	rt.AddContact(NewContact(NewKademliaID("1111111500000000000000000000000000000005"), "localhost:8006"))

	targetID := NewKademliaID("1111111600000000000000000000000000000006")
	kClosest := 4

	closestContacts := rt.FindClosestContacts(targetID, kClosest)

	expectedContacts := []Contact{
		NewContact(NewKademliaID("1111111500000000000000000000000000000005"), "localhost:8006"),
		NewContact(NewKademliaID("1111111400000000000000000000000000000004"), "localhost:8005"),
		NewContact(NewKademliaID("1111111300000000000000000000000000000003"), "localhost:8004"),
	}

	// Check if all expected contacts are in the closest contacts
	if len(closestContacts) != len(expectedContacts) {
		t.Fatalf("Expected %d contacts, but got %d", len(expectedContacts), len(closestContacts))
	}

	for _, expected := range expectedContacts {
		found := false
		for _, contact := range closestContacts {
			if contact.ID.Equals(expected.ID) && contact.Address == expected.Address {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("Expected contact %s with address %s not found in closestContacts", expected.ID.String(), expected.Address)
		}
	}
}

func TestEmptyRoutingTable(t *testing.T) {
	rt := NewRoutingTable(
		NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000"),
	)

	targetID := NewKademliaID("1111111600000000000000000000000000000006")
	kClosest := 3

	closestContacts := rt.FindClosestContacts(targetID, kClosest)

	if len(closestContacts) != 0 {
		t.Fatalf("Expected 0 contacts, but got %d", len(closestContacts))
	}
}

func TestSingleContact(t *testing.T) {
	rt := NewRoutingTable(
		NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000"),
	)

	// Add a single contact
	rt.AddContact(NewContact(NewKademliaID("1111111300000000000000000000000000000003"), "localhost:8004"))

	targetID := NewKademliaID("1111111600000000000000000000000000000006")
	kClosest := 3

	closestContacts := rt.FindClosestContacts(targetID, kClosest)

	expectedContact := NewContact(NewKademliaID("1111111300000000000000000000000000000003"), "localhost:8004")

	if len(closestContacts) != 1 {
		t.Fatalf("Expected 1 contact, but got %d", len(closestContacts))
	}

	// Check if the ID comparison method is correctly implemented
	if !expectedContact.ID.Equals(closestContacts[0].ID) || closestContacts[0].Address != expectedContact.Address {
		t.Fatalf("Expected contact %s with address %s not found in closestContacts. Got %s with address %s.",
			expectedContact.ID.String(), expectedContact.Address,
			closestContacts[0].ID.String(), closestContacts[0].Address)
	}
}

func TestPrintAllIP(t *testing.T) {
	rt := NewRoutingTable(
		NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000"),
	)

	// Add a few contacts
	rt.AddContact(NewContact(NewKademliaID("FFFFFFFF00000000000000000000000000000001"), "localhost:8001"))
	rt.AddContact(NewContact(NewKademliaID("1111111100000000000000000000000000000001"), "localhost:8002"))

	// Capture the output of PrintAllIP
	actualOutput := captureOutput(func() {
		rt.PrintAllIP()
	})

	// Check if the output contains the expected contacts
	expectedContacts := []string{
		"Address: localhost:8001 ID: ffffffff00000000000000000000000000000001",
		"Address: localhost:8002 ID: 1111111100000000000000000000000000000001",
	}

	for _, expected := range expectedContacts {
		if !strings.Contains(actualOutput, expected) {
			t.Fatalf("Expected output to contain:\n%s\nbut it did not.\nActual output:\n%s", expected, actualOutput)
		}
	}
}

func TestPrintRoutingTable(t *testing.T) {
	rt := NewRoutingTable(
		NewContact(NewKademliaID("ffffffff00000000000000000000000000000000"), "localhost:8000"),
	)

	// Capture the output of PrintRoutingTable
	actualOutput := captureOutput(func() {
		rt.PrintRoutingTable()
	})

	// Check if the output contains the expected contacts
	expectedContacts := []string{
		"ffffffff00000000000000000000000000000000",
	}

	// Validate that the output contains the expected contacts
	for _, expected := range expectedContacts {
		if !strings.Contains(actualOutput, expected) {
			t.Fatalf("Expected output to contain:\n%s\nbut it did not.\nActual output:\n%s", expected, actualOutput)
		}
	}
}

func captureOutput(f func()) string {
	var buf bytes.Buffer
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = stdout
	buf.ReadFrom(r)
	return buf.String()
}
