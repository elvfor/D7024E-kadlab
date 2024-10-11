package kademlia

import (
	"encoding/json"
	"fmt"
	"net"
)

// NetworkInterface defines the methods for network operations
type NetworkInterface interface {
	Listen(k *Kademlia)
	SendPingMessage(sender *Contact, receiver *Contact) bool
	SendFindContactMessage(sender *Contact, receiver *Contact, target *Contact) ([]Contact, error)
	SendFindDataMessage(sender *Contact, receiver *Contact, hash string) ([]Contact, []byte, error)
	SendStoreMessage(sender *Contact, receiver *Contact, dataID *KademliaID, data []byte) bool
	SendMessage(sender *Contact, receiver *Contact, message interface{}) ([]byte, error)
}

// Network struct implements NetworkInterface
type Network struct {
	reponseChan chan Response
	conn        net.PacketConn
}

// Response struct for network responses
type Response struct {
	Data            []byte    `json:"data"`
	ClosestContacts []Contact `json:"closest_contacts"`
	Target          *Contact  `json:"target"`
}

// NewNetwork constructor for Network
func NewNetwork(conn net.PacketConn) *Network {
	return &Network{make(chan Response), conn}
}

// Message struct for network messages
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
	defer network.conn.Close()

	for {
		var buf [8192]byte
		n, addr, err := network.conn.ReadFrom(buf[0:])
		if err != nil {
			fmt.Println(err)
			return
		}
		var receivedMessage Message
		err = json.Unmarshal(buf[:n], &receivedMessage)
		if err != nil {
			fmt.Println("Error unmarshalling message:", err)
			continue
		}
		network.handleMessage(k, receivedMessage, addr)
	}
}

func (network *Network) handleMessage(k *Kademlia, receivedMessage Message, addr net.Addr) {
	switch receivedMessage.Type {
	case "PING":
		network.handlePing(k, receivedMessage, addr)
	case "STORE":
		network.handleStore(k, receivedMessage, addr)
	case "FIND_NODE":
		network.handleFindNode(k, receivedMessage, addr)
	case "FIND_DATA":
		network.handleFindData(k, receivedMessage, addr)
	}
}

func (network *Network) handlePing(k *Kademlia, receivedMessage Message, addr net.Addr) {
	pongMsg := Message{
		Type:     "PONG",
		SenderID: k.RoutingTable.Me.ID,
		SenderIP: k.RoutingTable.Me.Address,
	}
	data, _ := json.Marshal(pongMsg)
	_, err := network.conn.WriteTo(data, addr)
	if err != nil {
		fmt.Println("Error sending PONG:", err)
	} else {
		fmt.Println("Received PING. Adding contact with ID:", receivedMessage.SenderID.String(), "and IP:", receivedMessage.SenderIP)
		action := Action{
			Action:   "UpdateRT",
			SenderId: receivedMessage.SenderID,
			SenderIp: receivedMessage.SenderIP,
		}
		k.ActionChannel <- action
	}
}

func (network *Network) handleStore(k *Kademlia, receivedMessage Message, addr net.Addr) {
	storeOKMsg := Message{
		Type:     "STORE_OK",
		SenderID: k.RoutingTable.Me.ID,
		SenderIP: k.RoutingTable.Me.Address,
	}
	data, _ := json.Marshal(storeOKMsg)
	_, err := network.conn.WriteTo(data, addr)
	if err != nil {
		fmt.Println("Error sending STORE_OK:", err)
	} else {
		fmt.Println("Received STORE. Added contact to routing table with ID:", receivedMessage.SenderID.String(), "and IP:", receivedMessage.SenderIP)
		action := Action{
			Action:   "Store",
			Hash:     receivedMessage.DataID.String(),
			Data:     receivedMessage.Data,
			SenderId: receivedMessage.SenderID,
			SenderIp: receivedMessage.SenderIP,
		}
		k.ActionChannel <- action
	}
}

func (network *Network) handleFindNode(k *Kademlia, receivedMessage Message, addr net.Addr) {
	fmt.Println("Received FIND_NODE")
	if network.SendPingMessage(&k.RoutingTable.Me, &Contact{ID: receivedMessage.SenderID, Address: receivedMessage.SenderIP}) {
		action := Action{
			Action:   "UpdateRT",
			SenderId: receivedMessage.SenderID,
			SenderIp: receivedMessage.SenderIP,
		}
		k.ActionChannel <- action
	} else {
		fmt.Println("Error receiving PONG in FIND_NODE")
	}
	contact := Contact{ID: NewKademliaID(receivedMessage.TargetID), Address: receivedMessage.SenderIP}
	action := Action{
		Action:   "LookupContact",
		SenderId: NewKademliaID(receivedMessage.SenderID.String()),
		SenderIp: receivedMessage.SenderIP,
		Target:   &contact,
	}
	k.ActionChannel <- action
	responseChannel := <-network.reponseChan
	response := Response{
		Data:            responseChannel.Data,
		ClosestContacts: responseChannel.ClosestContacts,
	}
	responseChannel.Data, _ = json.Marshal(response)
	_, err := network.conn.WriteTo(responseChannel.Data, addr)
	if err != nil {
		fmt.Println("Error sending closest contacts:", err)
	}
}

func (network *Network) handleFindData(k *Kademlia, receivedMessage Message, addr net.Addr) {
	if network.SendPingMessage(&k.RoutingTable.Me, &Contact{ID: receivedMessage.SenderID, Address: receivedMessage.SenderIP}) {
		action := Action{
			Action:   "UpdateRT",
			SenderId: receivedMessage.SenderID,
			SenderIp: receivedMessage.SenderIP,
		}
		k.ActionChannel <- action
	}
	action := Action{
		Action:   "LookupData",
		SenderId: receivedMessage.SenderID,
		SenderIp: receivedMessage.SenderIP,
		Hash:     receivedMessage.TargetID,
	}
	k.ActionChannel <- action
	responseChannel := <-network.reponseChan

	response := Response{
		Data:            responseChannel.Data,
		ClosestContacts: responseChannel.ClosestContacts,
	}
	responseChannel.Data, _ = json.Marshal(response)
	_, err := network.conn.WriteTo(responseChannel.Data, addr)
	if err != nil {
		fmt.Println("Error sending closest contacts:", err)
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
		fmt.Println("Received PONG from", receiver.Address)
		return true
	} else {
		fmt.Println("Received unexpected message:", receivedMessage)
		return false
	}
}

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

	fmt.Printf("Raw response: %s\n", response) // Debug the raw response

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
		fmt.Println("Received STORE_OK from", receiver.Address)
		return true
	} else {
		fmt.Println("Received unexpected message:", responseMsg)
		return false
	}
}

func (network *Network) SendMessage(sender *Contact, receiver *Contact, message interface{}) ([]byte, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", receiver.Address)
	if err != nil {
		return nil, fmt.Errorf("error resolving UDP address: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, fmt.Errorf("error dialing UDP: %v", err)
	}
	defer conn.Close()

	data, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("error serializing message: %v", err)
	}

	_, err = conn.Write(data)
	if err != nil {
		return nil, fmt.Errorf("error sending message: %v", err)
	}

	var buf [8192]byte
	n, _, err := conn.ReadFromUDP(buf[0:])
	if err != nil {
		return nil, fmt.Errorf("error receiving response: %v", err)
	}

	return buf[:n], nil
}
