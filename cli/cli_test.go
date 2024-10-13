package cli_test

import (
	"bytes"
	"d7024e/cli" // Change this to the correct import path for your project structure
	"strings"
	"testing"
)

func TestReadUserInput(t *testing.T) {
	tests := []struct {
		input          string
		expectedCmd    string
		expectedArg    string
		expectedOutput string
	}{
		{"PING 192.168.1.1\n", "PING", "192.168.1.1", ">"},
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
