package kademlia

import (
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

	bucketIsFull, lastContact := bucket.AddContact(contact)
	if bucketIsFull {
		t.Error("Expected bucket to not be full")
	}
	if lastContact != nil {
		t.Error("Expected lastContact to be nil")
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
