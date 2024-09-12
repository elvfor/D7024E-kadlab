package kademlia

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
)

type Network struct {
}
type Message struct {
	Type     string // Type of message: "PING", "PONG", "FIND_NODE", etc.
	senderID string // ID of the node sending the message
	senderIP string // IP address of the node sending the message
}

func Listen(k *Kademlia) {
	fmt.Println("Listening on all interfaces on port 8000")
	// Resolve the given address
	addr := net.UDPAddr{
		Port: 8000,
		IP:   net.ParseIP("0.0.0.0"),
	}
	// Start listening for UDP packages on the given address
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()

	for {
		var buf [512]byte
		n, addr, err := conn.ReadFromUDP(buf[0:])
		if err != nil {
			fmt.Println(err)
			return
		}
		//print receiving message
		var receivedMessage Message
		err = json.Unmarshal(buf[:n], &receivedMessage)
		//switch on the message
		switch receivedMessage.Type {
		case "PING":
			// Send "PONG" message back to the client
			pongMsg := Message{
				Type:     "PONG",
				senderID: k.RoutingTable.Me.ID.String(),
				senderIP: k.RoutingTable.Me.Address,
			}
			data, _ := json.Marshal(pongMsg)
			_, err = conn.WriteToUDP(data, addr)
			if err != nil {
				fmt.Println("Error sending PONG:", err)
			} else {
				//TODO : Add Kademlia Routing Table Logic on receiving PING
				k.HandlePingOrPong(k.RoutingTable.Me.ID.String(), k.RoutingTable.Me.Address)

			}
		case "STORE":
			//TODO : Add STORE logic
		case "FIND_NODE":
			//TODO : Add FIND_NODE logic
		case "FIND_DATA":
			//TODO : Add FIND_DATA logic
		}

	}
}

// PING
func (network *Network) SendPingMessage(source *Contact, target *Contact) bool {
	// Resolve the string address to a UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", target.Address)
	// Dial to the address with UDP
	conn, err := net.DialUDP("udp", nil, udpAddr)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Send a message to the server
	//_, err = conn.Write([]byte("PING"))
	pingMsg := Message{
		Type:     "PING",
		senderID: source.ID.String(),
		senderIP: source.Address,
	}
	fmt.Println("Sending Ping to ", target.Address+"\n")
	data, _ := json.Marshal(pingMsg)
	_, err = conn.Write(data)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Read the reply from the server, expected PONG
	buffer := make([]byte, 1024)
	var buf [512]byte
	n, _, err := conn.ReadFromUDP(buf[0:])
	if err != nil {
		fmt.Println(err)
		return false
	}
	//print receiving message
	var receivedMessage Message
	err = json.Unmarshal(buf[:n], &receivedMessage)
	if receivedMessage.Type == "PONG" && receivedMessage.senderIP == target.Address {
		fmt.Println("Received PONG from ", target.Address)
		//TODO : Add Kademlia Routing Table Logic
		//k.HandlePingOrPong(receivedMessage.senderID, receivedMessage.senderIP)
		return true
	} else {
		fmt.Println("Received unexpected message: ", string(buffer[0:n]))
		return false
	}

}

func (network *Network) SendFindContactMessage(contact *Contact) {
	// TODO
}

func (network *Network) SendFindDataMessage(hash string) {
	// TODO
}

func (network *Network) SendStoreMessage(data []byte) {
	// TODO
}
