package kademlia

import (
	"fmt"
	"net"
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
	//send ping over udp

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
