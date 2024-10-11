// TODO: Add package documentation for `main`, like this:
// Package main something something...
package main

import (
	"d7024e/cli"
	"d7024e/kademlia"
	"fmt"
	"log"
	"net"
	"time"
)

func main() {
	fmt.Println("Starting the kademlia app...")
	ip := GetOutboundIP().String()
	if ip == "172.20.0.6" {
		k := JoinNetworkBootstrap(ip)
		go k.ListenActionChannel()
		//wait for the network to be ready
		time.Sleep(1 * time.Second)
		go k.Network.Listen(k)
		go cli.UserInputHandler(k)
	} else {
		k := JoinNetwork(GetOutboundIP().String() + ":8000")
		go k.ListenActionChannel()
		go k.Network.Listen(k)
		time.Sleep(10 * time.Second)
		DoLookUpOnSelf(k)
		go cli.UserInputHandler(k)
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

	conn, err := net.ListenPacket("udp", ":8000")
	if err != nil {
		log.Fatal(err)
	}

	return kademlia.NewKademlia(routingTable, conn)
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
	//k.ActionChannel <- kademlia.Action{Action: "NODELOOKUP", Target: &k.RoutingTable.Me}
	contacts, _, _ := k.NodeLookup(&k.RoutingTable.Me, "")
	for _, contact := range contacts {
		action := kademlia.Action{
			Action:   "UpdateRT",
			SenderId: contact.ID,
			SenderIp: contact.Address,
		}
		k.ActionChannel <- action
	}
}

func JoinNetworkBootstrap(ip string) *kademlia.Kademlia {
	bootStrapContact := kademlia.NewContact(kademlia.NewKademliaID("FFFFFFFFF0000000000000000000000000000000)"), ip+":8000")
	bootStrapContact.CalcDistance(bootStrapContact.ID)
	routingTable := kademlia.NewRoutingTable(bootStrapContact)

	conn, err := net.ListenPacket("udp", ":8000")
	if err != nil {
		log.Fatal(err)
	}

	return kademlia.NewKademlia(routingTable, conn)
}
