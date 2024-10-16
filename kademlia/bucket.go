package kademlia

import (
	"container/list"
	"fmt"
)

// bucket definition
// contains a List
type bucket struct {
	list *list.List
}

// newBucket returns a new instance of a bucket
func newBucket() *bucket {
	bucket := &bucket{}
	bucket.list = list.New()
	return bucket
}

// AddContact adds the Contact to the front of the bucket
// or moves it to the front of the bucket if it already existed
func (bucket *bucket) AddContact(contact Contact) (bool, *Contact) {
	var element *list.Element
	for e := bucket.list.Front(); e != nil; e = e.Next() {
		nodeID := e.Value.(Contact).ID

		if (contact).ID.Equals(nodeID) {
			element = e
		}
	}

	if element == nil {
		//element non existing in bucket
		if bucket.list.Len() < bucketSize {
			bucket.list.PushFront(contact)
			return false, nil
		} else {
			// bucket is full
			lastContact := bucket.list.Back().Value.(Contact)
			return true, &lastContact
		}
	} else {
		//item already exists in bucket
		bucket.list.MoveToFront(element)
		return false, nil
	}
}

// RemoveContact removes the Contact from the bucket
func (bucket *bucket) RemoveContact(contact *Contact) {
	for e := bucket.list.Front(); e != nil; e = e.Next() {
		if e.Value.(Contact).ID.Equals(contact.ID) {
			bucket.list.Remove(e)
			break
		}
	}
}

// GetContactAndCalcDistance returns an array of Contacts where
// the distance has already been calculated
func (bucket *bucket) GetContactAndCalcDistance(target *KademliaID) []Contact {
	var contacts []Contact

	for elt := bucket.list.Front(); elt != nil; elt = elt.Next() {
		contact := elt.Value.(Contact)
		contact.CalcDistance(target)
		contacts = append(contacts, contact)
	}

	return contacts
}

// Len return the size of the bucket
func (bucket *bucket) Len() int {
	return bucket.list.Len()
}

func (bucket *bucket) PrintAllIP() {
	for elt := bucket.list.Front(); elt != nil; elt = elt.Next() {
		contact := elt.Value.(Contact)
		fmt.Println("Address: " + contact.Address + " ID: " + contact.ID.String())
	}
}
