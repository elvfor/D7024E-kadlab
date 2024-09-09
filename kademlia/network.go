package kademlia

import (
	"fmt"
	"net"
	"os"
)

type Network struct {
}

func Listen(ip string, port int) {
	fmt.Println("Listening on ", ip, ":", port)
	// Resolve the given address
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(ip),
	}
	// Start listening for UDP packages on the given address
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Println(err)
	}
	for {
		var buf [512]byte
		_, addr, err := conn.ReadFromUDP(buf[0:])
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Print("> ", string(buf[0:]))

		// Write back the message over UPD
		conn.WriteToUDP([]byte("Hello UDP Client\n"), addr)
	}
}

// PING
func (network *Network) SendPingMessage(contact *Contact) {
	// Resolve the string address to a UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", contact.Address)
	// Dial to the address with UDP
	conn, err := net.DialUDP("udp", nil, udpAddr)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Send a message to the server
	_, err = conn.Write([]byte("Hello UDP Server\n"))
	fmt.Println("send...")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
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

func handlePingMessage() {

}
