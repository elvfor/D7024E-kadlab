package kademlia

type MockNetwork struct {
	PingResponse        bool
	FindContactResponse []Contact
	FindContactError    error
	FindDataResponse    []byte
	FindDataContacts    []Contact
	FindDataError       error
	StoreResponse       bool
	SendMessageResponse []byte
	SendMessageError    error
}

func (m *MockNetwork) Listen(k *Kademlia) {}

func (m *MockNetwork) SendPingMessage(sender *Contact, receiver *Contact) bool {
	return m.PingResponse
}

func (m *MockNetwork) SendFindContactMessage(sender *Contact, receiver *Contact, target *Contact) ([]Contact, error) {
	return m.FindContactResponse, m.FindContactError
}

func (m *MockNetwork) SendFindDataMessage(sender *Contact, receiver *Contact, hash string) ([]Contact, []byte, error) {
	return m.FindDataContacts, m.FindDataResponse, m.FindDataError
}

func (m *MockNetwork) SendStoreMessage(sender *Contact, receiver *Contact, dataID *KademliaID, data []byte) bool {
	return m.StoreResponse
}

func (m *MockNetwork) SendMessage(sender *Contact, receiver *Contact, message interface{}) ([]byte, error) {
	return m.SendMessageResponse, m.SendMessageError
}
