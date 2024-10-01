// TODO: Add package documentation for `main`, like this:
// Package main something something...
package main

import (
	"d7024e/cli"
	"d7024e/kademlia"
	"fmt"
	"log"
	"net"
)

func main() {
	fmt.Println("Starting the kademlia app...")
	ip := GetOutboundIP().String()
	if ip == "172.20.0.6" {
		k := JoinNetworkBootstrap(ip)
		n := kademlia.Network{}
		go n.Listen(k)
		go cli.UserInputHandler(k)
		go k.ListenActionChannel()
	} else {
		k := JoinNetwork(GetOutboundIP().String() + ":8000")
		n := kademlia.Network{}
		go n.Listen(k)
		go DoLookUpOnSelf(k)
		go cli.UserInputHandler(k)
		go k.ListenActionChannel()
	}

	// Keep the main function running to prevent container exit
	select {}
}

func JoinNetwork(ip string) *kademlia.Kademlia {
	id := kademlia.NewRandomKademliaID()
	contact := kademlia.NewContact(id, ip)
	contact.CalcDistance(id)
	routingTable := kademlia.NewRoutingTable(contact)
	bootStrapContact := kademlia.NewContact(kademlia.NewKademliaID("FFFFFFFFF0000000000000000000000000000000)"), "172.20.0.6:8000")
	routingTable.AddContact(bootStrapContact)
	network := &kademlia.Network{}
	data := make(map[string][]byte)
	return &kademlia.Kademlia{RoutingTable: routingTable, Network: network, Data: &data}
}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}

func DoLookUpOnSelf(k *kademlia.Kademlia) {
	fmt.Println("Doing lookup on self")
	kClosest := k.NodeLookup(&k.RoutingTable.Me)
	//bootStrapContact := kademlia.NewContact(kademlia.NewKademliaID("FFFFFFFFF0000000000000000000000000000000)"), "172.20.0.6:8000")
	////kClosest, _ := k.Network.SendFindContactMessage(&k.RoutingTable.Me, &bootStrapContact, &k.RoutingTable.Me)
	fmt.Println("Length of kClosest: ", len(kClosest))
	for _, contact := range kClosest {
		k.UpdateRT(contact.ID, contact.Address)
	}
}

func JoinNetworkBootstrap(ip string) *kademlia.Kademlia {
	bootStrapContact := kademlia.NewContact(kademlia.NewKademliaID("FFFFFFFFF0000000000000000000000000000000)"), ip)
	bootStrapContact.CalcDistance(bootStrapContact.ID)
	routingTable := kademlia.NewRoutingTable(bootStrapContact)
	network := &kademlia.Network{}
	data := make(map[string][]byte)
	return &kademlia.Kademlia{RoutingTable: routingTable, Network: network, Data: &data}
}
