package kademlia

/*func TestNetwork_SendPingMessage(t *testing.T) {

	idSender := NewRandomKademliaID()
	contactSender := NewContact(idSender, "172.20.0.10:8000")
	contactSender.CalcDistance(idSender)
	rt := NewRoutingTable(contactSender)
	networkSender := &Network{}
	kSender := &Kademlia{RoutingTable: rt, Network: networkSender}
	//go Listen(kSender)

	idReceiver := NewRandomKademliaID()
	contactReceiver := NewContact(idReceiver, "172.20.0.11:8000")
	contactReceiver.CalcDistance(idReceiver)
	rtReceiver := NewRoutingTable(contactReceiver)
	networkReceiver := &Network{}
	kReceiver := &Kademlia{RoutingTable: rtReceiver, Network: networkReceiver}


}*/
