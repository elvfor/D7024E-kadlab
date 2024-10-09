package kademlia

import (
	"encoding/json"
	"fmt"
	"net"
)

type Network struct {
	reponseChan chan Response
}

type Response struct {
	Data            []byte    `json:"data"`
	ClosestContacts []Contact `json:"closest_contacts"`
	Target          *Contact  `json:"target"`
}

// constructor for Network
func NewNetwork() *Network {
	return &Network{make(chan Response)}
}

// TODO check if can change to have "contacts in struct"
type Message struct {
	Type     string      // Type of message: "PING", "PONG", "FIND_NODE", etc.
	SenderID *KademliaID // ID of the node sending the message
	SenderIP string      // IP address of the node sending the message
	TargetID string      // ID of the target node
	TargetIP string      // IP address of the target node
	DataID   *KademliaID // ID of the data
	Data     []byte
}

func (network *Network) Listen(k *Kademlia) {
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
				SenderID: k.RoutingTable.Me.ID,
				SenderIP: k.RoutingTable.Me.Address,
			}
			data, _ := json.Marshal(pongMsg)
			_, err = conn.WriteToUDP(data, addr)
			if err != nil {
				fmt.Println("Error sending PONG:", err)
			} else {
				//TODO : Add Kademlia Routing Table Logic on receiving PING
				fmt.Println("Received PONG. Adding contact to routing table with ID: ", receivedMessage.SenderID.String()+" and IP: "+receivedMessage.SenderIP)
				//go k.UpdateRT(receivedMessage.SenderID, receivedMessage.SenderIP)
				action := Action{
					Action:   "UpdateRT",
					SenderId: receivedMessage.SenderID,
					SenderIp: receivedMessage.SenderIP,
				}
				k.ActionChannel <- action
			}
		case "STORE":
			storeOKMsg := Message{
				Type:     "STORE_OK",
				SenderID: k.RoutingTable.Me.ID,
				SenderIP: k.RoutingTable.Me.Address,
			}
			data, _ := json.Marshal(storeOKMsg)
			_, err = conn.WriteToUDP(data, addr)
			if err != nil {
				fmt.Println("Error sending STORE_OK:", err)
			} else {
				fmt.Println("Received STORE. Adding contact to routing table with ID: ", receivedMessage.SenderID.String()+" and IP: "+receivedMessage.SenderIP)
				//k.Store(receivedMessage.DataID.String(), receivedMessage.Data)
				//k.UpdateRT(receivedMessage.SenderID, receivedMessage.SenderIP)
				action := Action{
					Action:   "Store",
					Hash:     receivedMessage.DataID.String(),
					Data:     receivedMessage.Data,
					SenderId: receivedMessage.SenderID,
					SenderIp: receivedMessage.SenderIP,
				}
				k.ActionChannel <- action

			}
		case "FIND_NODE":
			fmt.Println("Received FIND_NODE. Adding contact to routing table with ID: ", receivedMessage.SenderID.String()+" and IP: "+receivedMessage.SenderIP)
			//k.UpdateRT(receivedMessage.SenderID, receivedMessage.SenderIP)
			//closestContacts := k.LookupContact(&Contact{ID: NewKademliaID(receivedMessage.TargetID), Address: receivedMessage.TargetIP})
			contact := Contact{ID: NewKademliaID(receivedMessage.TargetID), Address: receivedMessage.SenderIP}
			action := Action{
				Action:   "LookupContact",
				SenderId: NewKademliaID(receivedMessage.SenderID.String()),
				SenderIp: receivedMessage.SenderIP,
				Target:   &contact,
			}
			k.ActionChannel <- action
			responseChannel, _ := <-network.reponseChan
			response := Response{
				Data:            responseChannel.Data,
				ClosestContacts: responseChannel.ClosestContacts,
			}
			responseChannel.Data, _ = json.Marshal(response)
			_, err = conn.WriteToUDP(responseChannel.Data, addr)
			if err != nil {
				fmt.Println("Error sending closest contacts:", err)
			}

		case "FIND_DATA":
			//k.UpdateRT(receivedMessage.SenderID, receivedMessage.SenderIP)
			//data, closestContacts := k.LookupData(receivedMessage.TargetID)
			action := Action{
				Action:   "LookupData",
				SenderId: receivedMessage.SenderID,
				SenderIp: receivedMessage.SenderIP,
				Hash:     receivedMessage.TargetID,
			}
			k.ActionChannel <- action
			responseChannel, _ := <-network.reponseChan

			response := Response{
				Data:            responseChannel.Data,
				ClosestContacts: responseChannel.ClosestContacts,
			}
			responseChannel.Data, _ = json.Marshal(response)
			_, err = conn.WriteToUDP(responseChannel.Data, addr)
			if err != nil {
				fmt.Println("Error sending closest contacts:", err)
			}
		}

	}
}

func (network *Network) SendPingMessage(sender *Contact, receiver *Contact) bool {
	pingMsg := Message{
		Type:     "PING",
		SenderID: sender.ID,
		SenderIP: sender.Address,
	}

	response, err := network.SendMessage(sender, receiver, pingMsg)
	if err != nil {
		fmt.Println("Error sending PING message:", err)
		return false
	}

	var receivedMessage Message
	err = json.Unmarshal(response, &receivedMessage)
	if err != nil {
		fmt.Println("Error unmarshalling response:", err)
		return false
	}

	if receivedMessage.Type == "PONG" {
		fmt.Println("Received PONG from ", receiver.Address)
		return true
	} else {
		fmt.Println("Received unexpected message:", receivedMessage)
		return false
	}
}

// TODO : Add error handling + double chech to return err if something goes wrong
func (network *Network) SendFindContactMessage(sender *Contact, receiver *Contact, target *Contact) ([]Contact, error) {

	findNodeMsg := Message{
		Type:     "FIND_NODE",
		SenderID: sender.ID,
		SenderIP: sender.Address,
		TargetID: target.ID.String(),
		TargetIP: target.Address,
	}

	response, err := network.SendMessage(sender, receiver, findNodeMsg)
	if err != nil {
		return nil, fmt.Errorf("error sending FIND_NODE message: %v", err)
	}

	var resp Response
	err = json.Unmarshal(response, &resp)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling contacts: %v", err)
	}
	closestContacts := resp.ClosestContacts
	fmt.Println("Closest contacts:", closestContacts)
	return closestContacts, nil
}

func (network *Network) SendFindDataMessage(sender *Contact, receiver *Contact, hash string) ([]Contact, []byte, error) {
	findNodeMsg := Message{
		Type:     "FIND_DATA",
		SenderID: sender.ID,
		SenderIP: sender.Address,
		TargetID: hash,
	}

	response, err := network.SendMessage(sender, receiver, findNodeMsg)
	if err != nil {
		return nil, nil, fmt.Errorf("error sending FIND_DATA message: %v", err)
	}
	type Response struct {
		Data            []byte    `json:"data"`
		ClosestContacts []Contact `json:"closest_contacts"`
	}

	var resp Response
	err = json.Unmarshal(response, &resp)
	if err != nil {
		return nil, nil, fmt.Errorf("error unmarshalling data: %v", err)
	}
	data := resp.Data
	closestContacts := resp.ClosestContacts

	return closestContacts, data, nil

}

func (network *Network) SendStoreMessage(sender *Contact, receiver *Contact, dataID *KademliaID, data []byte) bool {
	storeMsg := Message{
		Type:     "STORE",
		SenderID: sender.ID,
		SenderIP: sender.Address,
		DataID:   dataID,
		Data:     data,
	}

	response, err := network.SendMessage(sender, receiver, storeMsg)
	if err != nil {
		fmt.Println("Error sending STORE message:", err)
		return false
	}

	var responseMsg Message
	err = json.Unmarshal(response, &responseMsg)
	if err != nil {
		fmt.Println("Error unmarshalling response:", err)
		return false
	}
	fmt.Println("Response message:", responseMsg.Type)
	if responseMsg.Type == "STORE_OK" {
		fmt.Println("Received STORE_OK from ", receiver.Address)
		return true
	} else {
		fmt.Println("Received unexpected message:", responseMsg)
		return false
	}
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
