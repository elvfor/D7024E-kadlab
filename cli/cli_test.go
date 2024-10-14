package cli

import (
	"d7024e/kademlia"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestNewCLI_ReturnsCLIInstance(t *testing.T) {
	k := &kademlia.Kademlia{}
	cli := NewCLI(k)

	if cli == nil {
		t.Fatal("Expected non-nil CLI instance")
	}
	if cli.kademlia != k {
		t.Errorf("Expected kademlia instance to be set, got %v", cli.kademlia)
	}
	if cli.reader != os.Stdin {
		t.Errorf("Expected reader to be os.Stdin, got %v", cli.reader)
	}
	if cli.writer != os.Stdout {
		t.Errorf("Expected writer to be os.Stdout, got %v", cli.writer)
	}
}

func TestNewCLI_WithCustomReaderWriter(t *testing.T) {
	k := &kademlia.Kademlia{}
	reader := strings.NewReader("")
	writer := &strings.Builder{}
	cli := &CLI{
		kademlia: k,
		reader:   reader,
		writer:   writer,
	}

	if cli.reader != reader {
		t.Errorf("Expected reader to be set, got %v", cli.reader)
	}
	if cli.writer != writer {
		t.Errorf("Expected writer to be set, got %v", cli.writer)
	}
}

func TestHandleGet_InvalidArgumentLength(t *testing.T) {
	k := &kademlia.Kademlia{}
	writer := &strings.Builder{}
	cli := &CLI{
		kademlia: k,
		reader:   strings.NewReader(""),
		writer:   writer,
	}

	arg := "invalid_length"
	cli.handleGet(arg)

	expectedOutput := "error: Invalid Kademlia ID length"
	if !strings.Contains(writer.String(), expectedOutput) {
		t.Errorf("Expected output to contain '%s', got '%s'", expectedOutput, writer.String())
	}
}

func TestHandleGet_EmptyArgument(t *testing.T) {
	k := &kademlia.Kademlia{}
	writer := &strings.Builder{}
	cli := &CLI{
		kademlia: k,
		reader:   strings.NewReader(""),
		writer:   writer,
	}

	arg := ""
	cli.handleGet(arg)

	expectedOutput := "error: No argument provided for GET"
	if !strings.Contains(writer.String(), expectedOutput) {
		t.Errorf("Expected output to contain '%s', got '%s'", expectedOutput, writer.String())
	}
}
func TestCreateTargetContact_ValidArgument(t *testing.T) {
	cli := &CLI{}
	arg := "a94a8fe5ccb19ba61c4c0873d391e987982fbbd3"
	contact := cli.CreateTargetContact(arg)

	if contact.ID.String() != arg {
		t.Errorf("Expected contact ID to be '%s', got '%s'", arg, contact.ID.String())
	}
}
func TestHandleLookupResult_DataFound(t *testing.T) {
	writer := &strings.Builder{}
	cli := &CLI{writer: writer}
	contact := kademlia.NewContact(kademlia.NewKademliaID("a94a8fe5ccb19ba61c4c0873d391e987982fbbd3"), "127.0.0.1")
	data := []byte("some data")

	cli.HandleLookupResult(contact, data)

	expectedOutput := "Data found on contact: " + contact.String() + "\nData: some data\n"
	if writer.String() != expectedOutput {
		t.Errorf("Expected output to be '%s', got '%s'", expectedOutput, writer.String())
	}
}

func TestHandleLookupResult_DataNotFound(t *testing.T) {
	writer := &strings.Builder{}
	cli := &CLI{writer: writer}
	contact := kademlia.NewContact(kademlia.NewKademliaID("a94a8fe5ccb19ba61c4c0873d391e987982fbbd3"), "127.0.0.1")

	cli.HandleLookupResult(contact, nil)

	expectedOutput := "Data not found.\n"
	if writer.String() != expectedOutput {
		t.Errorf("Expected output to be '%s', got '%s'", expectedOutput, writer.String())
	}
}
func TestValidatePutArg_EmptyArgument(t *testing.T) {
	cli := &CLI{}
	arg := ""
	err := cli.ValidatePutArg(arg)

	if err == nil || err.Error() != "Error: No argument provided for PUT." {
		t.Errorf("Expected error 'Error: No argument provided for PUT.', got '%v'", err)
	}
}

func TestValidatePutArg_ValidArgument(t *testing.T) {
	cli := &CLI{}
	arg := "some data"
	err := cli.ValidatePutArg(arg)

	if err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
}

/*func TestCreatePutTargetContact_ValidData(t *testing.T) {
	cli := &CLI{}
	data := []byte("some data")
	kadId, contact := cli.CreatePutTargetContact(data)

	expectedKadId := "a94a8fe5ccb19ba61c4c0873d391e987982fbbd3"
	if kadId.String() != expectedKadId {
		t.Errorf("Expected Kademlia ID to be '%s', got '%s'", expectedKadId, kadId.String())
	}
	if contact.ID.String() != expectedKadId {
		t.Errorf("Expected contact ID to be '%s', got '%s'", expectedKadId, contact.ID.String())
	}
}*/

func TestCreatePutTargetContact_EmptyData(t *testing.T) {
	cli := &CLI{}
	data := []byte("")
	kadId, contact := cli.CreatePutTargetContact(data)

	expectedKadId := "da39a3ee5e6b4b0d3255bfef95601890afd80709"
	if kadId.String() != expectedKadId {
		t.Errorf("Expected Kademlia ID to be '%s', got '%s'", expectedKadId, kadId.String())
	}
	if contact.ID.String() != expectedKadId {
		t.Errorf("Expected contact ID to be '%s', got '%s'", expectedKadId, contact.ID.String())
	}
}

/*
	func TestCreatePutTargetContact_SpecialCharacters(t *testing.T) {
		cli := &CLI{}
		data := []byte("!@#$%^&*()")
		kadId, contact := cli.CreatePutTargetContact(data)

		expectedKadId := "d3486ae9136e7856bc42212385ea797094475802"
		if kadId.String() != expectedKadId {
			t.Errorf("Expected Kademlia ID to be '%s', got '%s'", expectedKadId, kadId.String())
		}
		if contact.ID.String() != expectedKadId {
			t.Errorf("Expected contact ID to be '%s', got '%s'", expectedKadId, contact.ID.String())
		}
	}
*/
func TestReadUserInput_ValidCommandAndArgument(t *testing.T) {
	reader := strings.NewReader("GET some_data\n")
	writer := &strings.Builder{}
	cli := &CLI{reader: reader, writer: writer}

	command, arg, err := cli.ReadUserInput()

	if err != nil {
		t.Fatalf("Expected no error, got '%v'", err)
	}
	if command != "GET" {
		t.Errorf("Expected command to be 'GET', got '%s'", command)
	}
	if arg != "some_data" {
		t.Errorf("Expected argument to be 'some_data', got '%s'", arg)
	}
}

func TestReadUserInput_ValidCommandNoArgument(t *testing.T) {
	reader := strings.NewReader("EXIT\n")
	writer := &strings.Builder{}
	cli := &CLI{reader: reader, writer: writer}

	command, arg, err := cli.ReadUserInput()

	if err != nil {
		t.Fatalf("Expected no error, got '%v'", err)
	}
	if command != "EXIT" {
		t.Errorf("Expected command to be 'EXIT', got '%s'", command)
	}
	if arg != "" {
		t.Errorf("Expected argument to be empty, got '%s'", arg)
	}
}

func TestReadUserInput_EmptyInput(t *testing.T) {
	reader := strings.NewReader("\n")
	writer := &strings.Builder{}
	cli := &CLI{reader: reader, writer: writer}

	command, arg, err := cli.ReadUserInput()

	if err != nil {
		t.Fatalf("Expected no error, got '%v'", err)
	}
	if command != "" {
		t.Errorf("Expected command to be empty, got '%s'", command)
	}
	if arg != "" {
		t.Errorf("Expected argument to be empty, got '%s'", arg)
	}
}

func TestReadUserInput_ErrorReadingInput(t *testing.T) {
	reader := &errorReader{}
	writer := &strings.Builder{}
	cli := &CLI{reader: reader, writer: writer}

	_, _, err := cli.ReadUserInput()

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("read error")
}
func TestUserInputHandler_ExitCommand(t *testing.T) {
	reader := strings.NewReader("EXIT\n")
	writer := &strings.Builder{}
	cli := &CLI{reader: reader, writer: writer}

	cli.UserInputHandler()

	expectedOutput := ">You entered: command=EXIT, argument=\nExiting program.\n"
	if writer.String() != expectedOutput {
		t.Errorf("Expected output to be '%s', got '%s'", expectedOutput, writer.String())
	}
}

func TestHandleStoreResult_SuccessfulStorage(t *testing.T) {
	writer := &strings.Builder{}
	cli := &CLI{writer: writer}

	cli.HandleStoreResult(3, 4, "a94a8fe5ccb19ba61c4c0873d391e987982fbbd3")

	expectedOutput := "Data stored successfully. Hash: a94a8fe5ccb19ba61c4c0873d391e987982fbbd3\n"
	if writer.String() != expectedOutput {
		t.Errorf("Expected output to be '%s', got '%s'", expectedOutput, writer.String())
	}
}

func TestHandleStoreResult_FailedStorage(t *testing.T) {
	writer := &strings.Builder{}
	cli := &CLI{writer: writer}

	cli.HandleStoreResult(1, 4, "a94a8fe5ccb19ba61c4c0873d391e987982fbbd3")

	expectedOutput := "Failed to store data.\n"
	if writer.String() != expectedOutput {
		t.Errorf("Expected output to be '%s', got '%s'", expectedOutput, writer.String())
	}
}

func TestHandleStoreResult_ExactHalfSuccess(t *testing.T) {
	writer := &strings.Builder{}
	cli := &CLI{writer: writer}

	cli.HandleStoreResult(2, 4, "a94a8fe5ccb19ba61c4c0873d391e987982fbbd3")

	expectedOutput := "Failed to store data.\n"
	if writer.String() != expectedOutput {
		t.Errorf("Expected output to be '%s', got '%s'", expectedOutput, writer.String())
	}
}
