package kademlia

import "fmt"

const bucketSize = 5

// RoutingTable definition
// keeps a refrence contact of me and an array of buckets
type RoutingTable struct {
	Me      Contact
	buckets [IDLength * 8]*bucket
}

// NewRoutingTable returns a new instance of a RoutingTable
func NewRoutingTable(me Contact) *RoutingTable {
	routingTable := &RoutingTable{}
	for i := 0; i < IDLength*8; i++ {
		routingTable.buckets[i] = newBucket()
	}
	routingTable.Me = me
	return routingTable
}

// AddContact add a new contact to the correct Bucket
func (routingTable *RoutingTable) AddContact(contact Contact) (bool, *Contact) {
	bucketIndex := routingTable.getBucketIndex(contact.ID)
	bucket := routingTable.buckets[bucketIndex]
	bucketIsFull, lastContact := bucket.AddContact(contact)
	return bucketIsFull, lastContact
}

// RemoveContact removes a contact from the correct Bucket
func (routingTable *RoutingTable) RemoveContact(contact *Contact) {
	bucketIndex := routingTable.getBucketIndex(contact.ID)
	bucket := routingTable.buckets[bucketIndex]
	bucket.RemoveContact(contact)
}

// FindClosestContacts finds the count closest Contacts to the target in the RoutingTable
func (routingTable *RoutingTable) FindClosestContacts(target *KademliaID, count int) []Contact {
	var candidates ContactCandidates
	bucketIndex := routingTable.getBucketIndex(target)
	bucket := routingTable.buckets[bucketIndex]

	candidates.Append(bucket.GetContactAndCalcDistance(target))

	for i := 1; (bucketIndex-i >= 0 || bucketIndex+i < IDLength*8) && candidates.Len() < count; i++ {
		if bucketIndex-i >= 0 {
			bucket = routingTable.buckets[bucketIndex-i]
			candidates.Append(bucket.GetContactAndCalcDistance(target))
		}
		if bucketIndex+i < IDLength*8 {
			bucket = routingTable.buckets[bucketIndex+i]
			candidates.Append(bucket.GetContactAndCalcDistance(target))
		}
	}

	candidates.Sort()

	if count > candidates.Len() {
		count = candidates.Len()
	}

	return candidates.GetContacts(count)
}

// getBucketIndex get the correct Bucket index for the KademliaID
func (routingTable *RoutingTable) getBucketIndex(id *KademliaID) int {
	distance := id.CalcDistance(routingTable.Me.ID)
	for i := 0; i < IDLength; i++ {
		for j := 0; j < 8; j++ {
			if (distance[i]>>uint8(7-j))&0x1 != 0 {
				return i*8 + j
			}
		}
	}

	return IDLength*8 - 1
}

// PrintAllIP prints all the IP addresses in the RoutingTable
func (routingTable *RoutingTable) PrintAllIP() {
	for i := 0; i < IDLength*8; i++ {
		bucket := routingTable.buckets[i]
		if bucket.Len() > 0 {
			//PRINT IP and ID
			fmt.Printf("Bucket %d: %v\n", i, bucket)
			bucket.PrintAllIP()
		}
	}
}

// PrintRoutingTable prints the RoutingTable
func (routingTable *RoutingTable) PrintRoutingTable() {
	fmt.Println("Routing Table:")
	fmt.Println("Me: ", routingTable.Me)
	for i := 0; i < IDLength*8; i++ {
		bucket := routingTable.buckets[i]
		if bucket.Len() > 0 {
			fmt.Printf("Bucket %d: %v\n", i, bucket)
		}
	}
}
