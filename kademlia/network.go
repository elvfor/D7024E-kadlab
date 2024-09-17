package kademlia

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
)

type Network struct {
}

// TODO check if can change to have "contacts in struct"
type Message struct {
	Type     string // Type of message: "PING", "PONG", "FIND_NODE", etc.
	SenderID string // ID of the node sending the message
	SenderIP string // IP address of the node sending the message
	TargetID string // ID of the target node
	TargetIP string // IP address of the target node
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
				SenderID: k.RoutingTable.Me.ID.String(),
				SenderIP: k.RoutingTable.Me.Address,
			}
			data, _ := json.Marshal(pongMsg)
			_, err = conn.WriteToUDP(data, addr)
			if err != nil {
				fmt.Println("Error sending PONG:", err)
			} else {
				//TODO : Add Kademlia Routing Table Logic on receiving PING
				fmt.Println("Adding contact to routing table with ID: ", receivedMessage.SenderID+" and IP: "+receivedMessage.SenderIP)
				k.UpdateRT(receivedMessage.SenderID, receivedMessage.SenderIP)

			}
		case "STORE":
			//TODO : Add STORE logic
		case "FIND_NODE":
			k.UpdateRT(receivedMessage.SenderID, receivedMessage.SenderIP)
			closestContacts := k.LookupContact(&Contact{ID: NewKademliaID(receivedMessage.TargetID), Address: receivedMessage.TargetIP})
			data, _ := json.Marshal(closestContacts)
			_, err = conn.WriteToUDP(data, addr)
			if err != nil {
				fmt.Println("Error sending closest contacts:", err)
			} else {
				fmt.Println("Sending K closest neighbours")
			}

		case "FIND_DATA":
			//TODO : Add FIND_DATA logic
		}

	}
}

// PING
// TODO BReakout and generalize sending/receiving messages
func (network *Network) SendPingMessage(sender *Contact, receiver *Contact) bool {
	// Resolve the string address to a UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", receiver.Address)
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
		SenderID: sender.ID.String(),
		SenderIP: sender.Address,
	}
	fmt.Println("Sending Ping to ", receiver.Address+"\n"+"with source ID: "+pingMsg.SenderID+" and source IP: "+pingMsg.SenderIP)
	data, _ := json.Marshal(pingMsg)
	_, err = conn.Write(data)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Read the reply from the server, expected PONG
	buffer := make([]byte, 1024)
	var buf [512]byte
	n, receivedAddr, err := conn.ReadFromUDP(buf[0:])
	if err != nil {
		fmt.Println(err)
		return false
	}
	//print receiving message
	var receivedMessage Message
	err = json.Unmarshal(buf[:n], &receivedMessage)
	if receivedMessage.Type == "PONG" {
		fmt.Println("Received PONG from ", receivedAddr)
		return true
	} else {
		fmt.Println("Received unexpected message: ", string(buffer[0:n]))
		return false
	}
}

func (network *Network) SendFindContactMessage(sender *Contact, receiver *Contact, target *Contact) ([]Contact, error) {
	// Resolve the string address to a UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", receiver.Address)
	// Dial to the address with UDP
	conn, err := net.DialUDP("udp", nil, udpAddr)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	findNodeMsg := Message{
		Type:     "FIND_NODE",
		SenderID: sender.ID.String(),
		SenderIP: sender.Address,
		TargetID: target.ID.String(),
		TargetIP: sender.Address,
	}

	// Serialize
	data, _ := json.Marshal(findNodeMsg)
	// Send the message
	_, err = conn.Write(data)
	if err != nil {
		return nil, fmt.Errorf("error sending FIND_NODE message: %v", err)
	}

	// Receive the message
	// Read the reply from the server, expected list of Contacts
	var buf [512]byte
	n, _, err := conn.ReadFromUDP(buf[0:])
	if err != nil {
		return nil, fmt.Errorf("error receiving response: %v", err)
	}

	// Deserialize the received message (expected to be a list of contacts)
	var closestContacts []Contact
	err = json.Unmarshal(buf[:n], &closestContacts)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling contacts: %v", err)
	}
	fmt.Print("Closest contacts:", closestContacts)
	return closestContacts, nil
}

func (network *Network) SendFindDataMessage(hash string) {

}

func (network *Network) SendStoreMessage(data []byte) {
	// TODO
}

/* SendMsg(msg Message, target *Contact) {
	// Resolve the string address to a UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", target.Address)
	// Dial to the address with UDP
	conn, err := net.DialUDP("udp", nil, udpAddr)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}*/
