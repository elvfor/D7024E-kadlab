package kademlia

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
)

type Network struct {
}
type PingMessage struct {
	Type     string // Type of message: "PING", "PONG", "FIND_NODE", etc.
	senderID string // ID of the node sending the message
	senderIP string // IP address of the node sending the message}
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
		fmt.Println("Received ", string(buf[0:n]), " from ", addr)

		//switch on the message
		switch strings.TrimSpace(string(buf[0:n])) {
		case "PING":
			// Send "PONG" message back to the client
			_, err := conn.WriteToUDP([]byte("PONG"), addr)
			if err != nil {
				fmt.Println("Error sending PONG:", err)
			} else {
				//TODO : Add Kademlia Routing Table Logic on receiving PING
				//call for k.HandlePing with the correct contact
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
func (network *Network) SendPingMessage(source *Contact, target *Contact) {
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
	pingMsg := PingMessage{
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
	n, _, err := conn.ReadFromUDP(buffer)
	if err != nil {
		fmt.Println(err)
		return
	}
	if strings.TrimSpace(string(buffer[0:n])) == "PONG" {
		fmt.Println("Received PONG from ", source.Address)
		//TODO : Add Kademlia Routing Table Logic
	} else {
		fmt.Println("Received unexpected message: ", string(buffer[0:n]))
	}
	fmt.Printf("Reply: %s\n", string(buffer[0:n]))
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
