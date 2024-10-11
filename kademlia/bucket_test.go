package kademlia

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

const bucketTestSize = 5

func TestNewBucket(t *testing.T) {
	bucket := newBucket()
	if bucket == nil {
		t.Error("Expected new bucket to be created")
	}
	if bucket.list == nil {
		t.Error("Expected bucket to contain a list")
	}
}

func TestAddContact(t *testing.T) {
	bucket := newBucket()
	contact := NewContact(NewRandomKademliaID(), "127.0.0.1:8000")

	// Test adding a new contact
	bucketIsFull, lastContact := bucket.AddContact(contact)
	if bucketIsFull {
		t.Error("Expected bucket to not be full")
	}
	if lastContact != nil {
		t.Error("Expected lastContact to be nil")
	}

	// Test moving an existing contact to the front
	bucket.AddContact(contact)
	if bucket.list.Front().Value.(Contact).ID != contact.ID {
		t.Error("Expected existing contact to be moved to the front")
	}

	// Fill the bucket
	for i := 0; i < bucketTestSize; i++ {
		bucket.AddContact(NewContact(NewRandomKademliaID(), "127.0.0.1:8000"))
	}

	// Add one more contact to fill the bucket
	bucketIsFull, lastContact = bucket.AddContact(NewContact(NewRandomKademliaID(), "127.0.0.1:8000"))
	if !bucketIsFull {
		t.Error("Expected bucket to be full")
	}
	if lastContact == nil {
		t.Error("Expected lastContact to not be nil")
	}
}

func TestRemoveContact(t *testing.T) {
	bucket := newBucket()
	contact := NewContact(NewRandomKademliaID(), "127.0.0.1:8000")
	bucket.AddContact(contact)

	bucket.RemoveContact(&contact)
	if bucket.list.Len() != 0 {
		t.Error("Expected contact to be removed from bucket")
	}
}

func TestGetContactAndCalcDistance(t *testing.T) {
	bucket := newBucket()
	target := NewRandomKademliaID()
	contact := NewContact(NewRandomKademliaID(), "127.0.0.1:8000")
	bucket.AddContact(contact)

	contacts := bucket.GetContactAndCalcDistance(target)
	if len(contacts) != 1 {
		t.Error("Expected one contact in the bucket")
	}
}

func TestLen(t *testing.T) {
	bucket := newBucket()
	if bucket.Len() != 0 {
		t.Error("Expected bucket length to be 0")
	}
	contact := NewContact(NewRandomKademliaID(), "127.0.0.1:8000")
	bucket.AddContact(contact)
	if bucket.Len() != 1 {
		t.Error("Expected bucket length to be 1")
	}
}

func TestPrintAllIPBucket(t *testing.T) {
	// Create a new bucket and add some contacts
	bucket := newBucket()
	contact1 := NewContact(NewRandomKademliaID(), "127.0.0.1:8000")
	contact2 := NewContact(NewRandomKademliaID(), "127.0.0.1:8001")
	bucket.AddContact(contact1)
	bucket.AddContact(contact2)

	// Capture the output of PrintAllIP
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w

	// Call PrintAllIP
	bucket.PrintAllIP()

	// Restore os.Stdout and close the writer
	w.Close()
	os.Stdout = old

	// Read the captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	r.Close()

	// Verify the output
	output := buf.String()
	if !strings.Contains(output, contact1.Address) || !strings.Contains(output, contact1.ID.String()) {
		t.Errorf("Expected output to contain contact1's address and ID, got: %s", output)
	}
	if !strings.Contains(output, contact2.Address) || !strings.Contains(output, contact2.ID.String()) {
		t.Errorf("Expected output to contain contact2's address and ID, got: %s", output)
	}
}
