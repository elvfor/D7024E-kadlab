package cli

/*func TestUserInputHandler_HandlesPingCommand(t *testing.T) {
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
}*/
