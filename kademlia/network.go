package kademlia

import (
	"encoding/json"
	"fmt"
	"net"
)

type Network struct {
}

// TODO check if can change to have "contacts in struct"
type Message struct {
	Type     string      // Type of message: "PING", "PONG", "FIND_NODE", etc.
	SenderID string      // ID of the node sending the message
	SenderIP string      // IP address of the node sending the message
	TargetID string      // ID of the target node
	TargetIP string      // IP address of the target node
	DataID   *KademliaID // ID of the data
	Data     []byte
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
				go k.UpdateRT(receivedMessage.SenderID, receivedMessage.SenderIP)

			}
		case "STORE":
			//TODO :
			// Call to actually store the data
		case "FIND_NODE":
			go func() {
				k.UpdateRT(receivedMessage.SenderID, receivedMessage.SenderIP)
				closestContacts := k.LookupContact(&Contact{ID: NewKademliaID(receivedMessage.TargetID), Address: receivedMessage.TargetIP})
				data, _ := json.Marshal(closestContacts)
				_, err = conn.WriteToUDP(data, addr)
				if err != nil {
					fmt.Println("Error sending closest contacts:", err)
				} else {
					fmt.Println("Sending K closest neighbours")
				}
			}()

		case "FIND_DATA":
			//TODO : Add FIND_DATA logic
			// If routing table contains data: return
			// If not call to check for k closest contacts
		}

	}
}

func (network *Network) SendPingMessage(sender *Contact, receiver *Contact) bool {
	resultChan := make(chan bool)
	go func() {
		defer close(resultChan)
		pingMsg := Message{
			Type:     "PING",
			SenderID: sender.ID.String(),
			SenderIP: sender.Address,
		}

		response, err := network.SendMessage(sender, receiver, pingMsg)
		if err != nil {
			fmt.Println("Error sending Ping:", err)
			resultChan <- false
		}

		var receivedMessage Message
		err = json.Unmarshal(response, &receivedMessage)
		if err != nil {
			fmt.Println("Error unmarshalling response:", err)
			resultChan <- false
		}

		if receivedMessage.Type == "PONG" {
			fmt.Println("Received PONG from ", receiver.Address)
			resultChan <- true
		} else {
			fmt.Println("Received unexpected message:", receivedMessage)
			resultChan <- false
		}
	}()
	return <-resultChan
}

func (network *Network) SendFindContactMessage(sender *Contact, receiver *Contact, target *Contact) ([]Contact, error) {
	contactsChan := make(chan []Contact)
	errChan := make(chan error)
	go func() {
		defer close(contactsChan)
		defer close(errChan)
		findNodeMsg := Message{
			Type:     "FIND_NODE",
			SenderID: sender.ID.String(),
			SenderIP: sender.Address,
			TargetID: target.ID.String(),
			TargetIP: target.Address,
		}

		response, err := network.SendMessage(sender, receiver, findNodeMsg)
		if err != nil {
			errChan <- err
			fmt.Errorf("error sending FIND_NODE message: %v", err)
			return
		}

		var closestContacts []Contact
		err = json.Unmarshal(response, &closestContacts)
		if err != nil {
			errChan <- err
			fmt.Errorf("error unmarshalling contacts: %v", err)
			return
		}

		fmt.Println("Closest contacts:", closestContacts)
		contactsChan <- closestContacts
		return
	}()
	return <-contactsChan, <-errChan
}

func (network *Network) SendFindDataMessage(hash string) {

}

func (network *Network) SendStoreMessage(sender *Contact, receiver *Contact, dataID *KademliaID, data []byte) ([]byte, error) {
	dataChan := make(chan []byte)
	errChan := make(chan error)
	go func() {
		defer close(dataChan)
		defer close(errChan)
		storeMsg := Message{
			Type:     "STORE",
			SenderID: sender.ID.String(),
			SenderIP: sender.Address,
			DataID:   dataID,
			Data:     data,
		}

		response, err := network.SendMessage(sender, receiver, storeMsg)
		if err != nil {
			errChan <- err
			fmt.Errorf("error sending STORE message: %v", err)
			return
		}

		var data []byte
		err = json.Unmarshal(response, &data)
		if err != nil {
			errChan <- err
			fmt.Errorf("error unmarshalling data: %v", err)
			return
		}

		fmt.Println("Data:", data)
		dataChan <- data
		return
	}()
	return <-dataChan, <-errChan
}

// SendMessage is a generalized function to send and receive UDP messages.
func (network *Network) SendMessage(sender *Contact, receiver *Contact, message interface{}) ([]byte, error) {
	// Resolve the string address to a UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", receiver.Address)
	if err != nil {
		return nil, fmt.Errorf("error resolving UDP address: %v", err)
	}

	// Dial to the address with UDP
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, fmt.Errorf("error dialing UDP: %v", err)
	}
	defer conn.Close()

	// Serialize the message
	data, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("error serializing message: %v", err)
	}

	// Send the message
	_, err = conn.Write(data)
	if err != nil {
		return nil, fmt.Errorf("error sending message: %v", err)
	}

	// Receive the response
	var buf [512]byte
	n, _, err := conn.ReadFromUDP(buf[0:])
	if err != nil {
		return nil, fmt.Errorf("error receiving response: %v", err)
	}

	// Return the received raw data for further processing
	return buf[:n], nil
}
