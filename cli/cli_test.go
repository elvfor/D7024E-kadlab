package cli_test

import (
	"bytes"
	"crypto/sha1"
	"d7024e/cli" // Change this to the correct import path for your project structure
	"d7024e/kademlia"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"testing"
)

type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("simulated read error")
}

func TestReadUserInput_ErrorReadingInput(t *testing.T) {
	reader := &errorReader{}
	writer := &strings.Builder{}

	_, _, err := cli.ReadUserInput(reader, writer)
	if err == nil {
		t.Fatal("Expected an error, but got nil")
	}

	expectedError := "error reading input: simulated read error"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', but got '%s'", expectedError, err.Error())
	}
}

func TestReadUserInput(t *testing.T) {
	tests := []struct {
		input          string
		expectedCmd    string
		expectedArg    string
		expectedOutput string
	}{
		{"GET ABCD1234\n", "GET", "ABCD1234", ">"},
		{"EXIT\n", "EXIT", "", ">"},
		{"UNKNOWNCOMMAND\n", "UNKNOWNCOMMAND", "", ">"},
	}

	for _, test := range tests {
		reader := strings.NewReader(test.input)
		writer := &bytes.Buffer{}
		cmd, arg, err := cli.ReadUserInput(reader, writer)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if cmd != test.expectedCmd {
			t.Errorf("Expected command %s, got %s", test.expectedCmd, cmd)
		}

		if arg != test.expectedArg {
			t.Errorf("Expected argument %s, got %s", test.expectedArg, arg)
		}

		if writer.String() != test.expectedOutput {
			t.Errorf("Expected output %s, got %s", test.expectedOutput, writer.String())
		}
	}
}

func TestValidateGetArg(t *testing.T) {
	tests := []struct {
		arg           string
		expectedError string
	}{
		{"", "error: No argument provided for GET"},                                          // Empty argument
		{"123", "error: Invalid Kademlia ID length"},                                         // Argument too short
		{"1234567890123456789012345678901234567890123", "error: Invalid Kademlia ID length"}, // Argument too long
		{"1234567890123456789012345678901234567890", ""},                                     // Valid argument
	}

	for _, test := range tests {
		err := cli.ValidateGetArg(test.arg)

		if err != nil {
			if err.Error() != test.expectedError {
				t.Errorf("Expected error '%s', got '%s'", test.expectedError, err.Error())
			}
		} else if test.expectedError != "" {
			t.Errorf("Expected error '%s', got nil", test.expectedError)
		}
	}
}

func TestCreateTargetContact(t *testing.T) {
	tests := []struct {
		arg          string
		expectedID   string
		expectedAddr string
	}{
		{"1234567890123456789012345678901234567890", "1234567890123456789012345678901234567890", ""}, // Valid 40-character Kademlia ID
		{"abcdefabcdefabcdefabcdefabcdefabcdefabcd", "abcdefabcdefabcdefabcdefabcdefabcdefabcd", ""}, // Another valid ID
	}

	for _, test := range tests {
		contact := cli.CreateTargetContact(test.arg)

		// Check if the Kademlia ID is created correctly
		if contact.ID.String() != test.expectedID {
			t.Errorf("Expected Kademlia ID %s, got %s", test.expectedID, contact.ID.String())
		}

		// Check if the contact address is empty as expected
		if contact.Address != test.expectedAddr {
			t.Errorf("Expected address %s, got %s", test.expectedAddr, contact.Address)
		}
	}
}

func TestHandleLookupResult_DataFound(t *testing.T) {
	// Setup mock contact and found data
	mockContact := kademlia.NewContact(kademlia.NewKademliaID("1234567890123456789012345678901234567890"), "127.0.0.1")
	foundData := []byte("mock data")

	// Capture the output printed by fmt.Println using a buffer
	var buf bytes.Buffer
	fmt.Fprint(&buf, "Test output start\n") // Add any expected prefix output here

	// Call the function, passing the buffer as the writer
	cli.HandleLookupResult(mockContact, foundData)

	// Check if output matches expected result
	output := buf.String()
	if !strings.Contains(output, "Data found on contact:") {
		t.Errorf("Expected 'Data found on contact:' but got: %s", output)
	}
	if !strings.Contains(output, "mock data") {
		t.Errorf("Expected data 'mock data' but got: %s", output)
	}
}

func TestHandleLookupResult_DataNotFound(t *testing.T) {
	// Setup mock contact and no data (nil)
	mockContact := kademlia.NewContact(kademlia.NewKademliaID("1234567890123456789012345678901234567890"), "127.0.0.1")
	var foundData []byte = nil

	// Capture the output printed by fmt.Println using a buffer
	var buf bytes.Buffer
	fmt.Fprint(&buf, "Test output start\n") // Add any expected prefix output here

	// Call the function, passing the buffer as the writer
	cli.HandleLookupResult(mockContact, foundData)

	// Check if output matches expected result
	output := buf.String()
	if !strings.Contains(output, "Data not found.") {
		t.Errorf("Expected 'Data not found.' but got: %s", output)
	}
}

func TestValidatePutArg(t *testing.T) {
	tests := []struct {
		arg         string
		expectedErr string
	}{
		{"", "Error: No argument provided for PUT."},
		{"somevalue", ""},
	}

	for _, test := range tests {
		err := cli.ValidatePutArg(test.arg)
		if err != nil && err.Error() != test.expectedErr {
			t.Errorf("Expected error '%s', got '%s'", test.expectedErr, err.Error())
		}
		if err == nil && test.expectedErr != "" {
			t.Errorf("Expected error '%s', got nil", test.expectedErr)
		}
	}
}

func TestCreatePutTargetContact(t *testing.T) {
	data := []byte("somevalue")
	kadId, contact := cli.CreatePutTargetContact(data)

	hasher := sha1.New()
	hasher.Write([]byte("hash1"))
	expectedHash := hex.EncodeToString(hasher.Sum(nil))

	if kadId.String() != expectedHash {
		t.Errorf("Expected Kademlia ID '%s', got '%s'", expectedHash, kadId.String())
	}

	if contact.ID.String() != expectedHash {
		t.Errorf("Expected contact ID '%s', got '%s'", expectedHash, contact.ID.String())
	}
}

func TestHandleStoreResult(t *testing.T) {
	tests := []struct {
		successCount   int
		totalContacts  int
		expectedOutput string
	}{
		{3, 5, "Data stored successfully."},
		{2, 5, "Failed to store data."},
	}

	for _, test := range tests {
		writer := &strings.Builder{}
		cli.HandleStoreResult(test.successCount, test.totalContacts)
		output := writer.String()

		if !strings.Contains(output, test.expectedOutput) {
			t.Errorf("Expected output to contain '%s', got '%s'", test.expectedOutput, output)
		}
	}
}

/*
type KademliaInterface interface {
	Ping(arg string)
	Get(arg string)
	Put(arg string)
	Lookup(arg string)
	Print()
	GetActionChannel() chan Action
}

func NewMockKademlia() *MockKademlia {
	return &MockKademlia{
		ActionChannel: make(chan kademlia.Action, 10),
	}
}

// Mock functions for command handlers
func handlePing(k *kademlia.Kademlia, arg string) {
	k.(*MockKademlia).PingCalled = true
}

func handleGet(k *kademlia.Kademlia, arg string) {
	k.(*MockKademlia).GetCalled = true
}

func handlePut(k *kademlia.Kademlia, arg string) {
	k.(*MockKademlia).PutCalled = true
}

func handleLookup(k *kademlia.Kademlia, arg string) {
	k.(*MockKademlia).LookupCalled = true
}

func TestUserInputHandler_PingCommand(t *testing.T) {
	mockK := NewMockKademlia()
	input := "PING 127.0.0.1\n"
	reader := strings.NewReader(input)
	writer := &bytes.Buffer{}

	// Call UserInputHandler in a goroutine to simulate input handling
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		UserInputHandler(mockK, reader, writer)
	}()
	wg.Wait()

	if !mockK.PingCalled {
		t.Error("Expected PingCalled to be true, but it was false")
	}
}

func TestUserInputHandler_GetCommand(t *testing.T) {
	mockK := NewMockKademlia()
	input := "GET ABC123\n"
	reader := strings.NewReader(input)
	writer := &bytes.Buffer{}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		UserInputHandler(mockK, reader, writer)
	}()
	wg.Wait()

	if !mockK.GetCalled {
		t.Error("Expected GetCalled to be true, but it was false")
	}
}

func TestUserInputHandler_PutCommand(t *testing.T) {
	mockK := NewMockKademlia()
	input := "PUT some_data\n"
	reader := strings.NewReader(input)
	writer := &bytes.Buffer{}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		UserInputHandler(mockK, reader, writer)
	}()
	wg.Wait()

	if !mockK.PutCalled {
		t.Error("Expected PutCalled to be true, but it was false")
	}
}

func TestUserInputHandler_ExitCommand(t *testing.T) {
	mockK := NewMockKademlia()
	input := "EXIT\n"
	reader := strings.NewReader(input)
	writer := &bytes.Buffer{}

	// Simulate the input handling
	UserInputHandler(mockK, reader, writer)

	// Check that the program exited correctly by looking at the output
	if !strings.Contains(writer.String(), "Exiting program.") {
		t.Error("Expected output to contain 'Exiting program.', but it did not")
	}
}

func TestUserInputHandler_UnknownCommand(t *testing.T) {
	mockK := NewMockKademlia()
	input := "UNKNOWN\n"
	reader := strings.NewReader(input)
	writer := &bytes.Buffer{}

	// Simulate the input handling
	UserInputHandler(mockK, reader, writer)

	// Check that the unknown command is handled properly
	if !strings.Contains(writer.String(), "Error: Unknown command.") {
		t.Error("Expected output to contain 'Error: Unknown command.', but it did not")
	}
}

func TestUserInputHandler_PrintCommand(t *testing.T) {
	mockK := NewMockKademlia()
	input := "PRINT\n"
	reader := strings.NewReader(input)
	writer := &bytes.Buffer{}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		UserInputHandler(mockK, reader, writer)
	}()
	wg.Wait()

	select {
	case action := <-mockK.ActionChannel:
		if action.Action != "PRINT" {
			t.Errorf("Expected ActionChannel to receive 'PRINT', got '%s'", action.Action)
		}
	default:
		t.Error("Expected ActionChannel to receive 'PRINT', but nothing was sent")
	}
}

/*
func TestUserInputHandler_HandlesPingCommand(t *testing.T) {
	k := &kademlia.Kademlia{}
	input := "PING 172.20.0.2:8000\n"
	reader := strings.NewReader(input)
	writer := &strings.Builder{}
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		UserInputHandler(k, reader, writer)
	}()

	wg.Wait()

	output := writer.String()
	if !strings.Contains(output, "You entered: command=PING, argument=172.20.0.2:8000") {
		t.Errorf("Expected output to contain 'You entered: command=PING, argument=172.20.0.2:8000', got %s", output)
	}
}

func TestUserInputHandler_HandlesGetCommand(t *testing.T) {
	k := &kademlia.Kademlia{}
	input := "GET somehash\n"
	reader := strings.NewReader(input)
	writer := &strings.Builder{}

	go UserInputHandler(k, reader, writer)

	// Add assertions to verify the behavior
	output := writer.String()
	if !strings.Contains(output, "You entered: command=GET, argument=somehash") {
		t.Errorf("Expected output to contain 'You entered: command=GET, argument=somehash', got %s", output)
	}
}

func TestUserInputHandler_HandlesPutCommand(t *testing.T) {
	k := &kademlia.Kademlia{}
	input := "PUT somevalue\n"
	reader := strings.NewReader(input)
	writer := &strings.Builder{}

	go UserInputHandler(k, reader, writer)

	// Add assertions to verify the behavior
	output := writer.String()
	if !strings.Contains(output, "You entered: command=PUT, argument=somevalue") {
		t.Errorf("Expected output to contain 'You entered: command=PUT, argument=somevalue', got %s", output)
	}
}

func TestUserInputHandler_HandlesExitCommand(t *testing.T) {
	k := &kademlia.Kademlia{}
	input := "EXIT\n"
	reader := strings.NewReader(input)
	writer := &strings.Builder{}

	go UserInputHandler(k, reader, writer)

	// Add assertions to verify the behavior
	output := writer.String()
	if !strings.Contains(output, "Exiting program.") {
		t.Errorf("Expected output to contain 'Exiting program.', got %s", output)
	}
}

func TestUserInputHandler_HandlesUnknownCommand(t *testing.T) {
	k := &kademlia.Kademlia{}
	input := "UNKNOWN\n"
	reader := strings.NewReader(input)
	writer := &strings.Builder{}

	go UserInputHandler(k, reader, writer)

	// Add assertions to verify the behavior
	output := writer.String()
	if !strings.Contains(output, "Error: Unknown command.") {
		t.Errorf("Expected output to contain 'Error: Unknown command.', got %s", output)
	}
}
*/
