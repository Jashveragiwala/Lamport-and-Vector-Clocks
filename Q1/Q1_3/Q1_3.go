package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var wg sync.WaitGroup

type Message struct {
	senderID        int
	messageID       int
	vectorTimeStamp []int
}

type Client struct {
	pID             int
	clientChannel   chan Message
	server          *Server
	closeChannel    chan bool
	readyChannel    chan int
	vectorTimeStamp []int
}

type Server struct {
	pID             int
	serverChannel   chan Message
	clientsArray    []*Client
	vectorTimeStamp []int
}

type Event struct {
	senderID        int
	receiverID      int
	messageID       int
	vectorTimeStamp []int
	eventType       int
}

var (
	NUM_CLIENTS   int
	NUM_MESSAGES  int
	NUM_EVENTS    int
	NUM_CLOCKS    int
	MESSAGE_DELAY = 100
)

const (
	CLIENT_SEND_EVENT      = 1
	SERVER_RECEIVE_EVENT   = 2
	SERVER_BROADCAST_EVENT = 3
	CLIENT_RECEIVE_EVENT   = 4
)

// Function to find the greaterimum of two integers
func greater(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

// Function to merge two vector timestamps
func mergeVectorClock(x []int, y []int, receiverID int) []int {
	z := make([]int, len(x))
	for i := range x {
		z[i] = greater(x[i], y[i])
	}
	z[receiverID] += 1
	return z
}

func (c Client) prepMsgs() {
	if NUM_MESSAGES == -1 {
		// Infinite loop for unlimited messages
		for messageID := 1; ; messageID++ {
			time.Sleep(time.Duration(MESSAGE_DELAY) * time.Millisecond)
			c.readyChannel <- messageID
		}
	} else {
		// Send the specified number of messages
		for i := 1; i <= NUM_MESSAGES; i++ {
			time.Sleep(time.Duration(MESSAGE_DELAY) * time.Millisecond)
			c.readyChannel <- i
		}
	}
}

func (s Server) serverListener(eventsChannel chan Event, pcvChannel chan string) {
	fmt.Println("Server is ready to receive messages...")

	doneClients := 0
	for {
		clientMessage := <-s.serverChannel

		if lesserVectorClock(clientMessage.vectorTimeStamp, s.vectorTimeStamp) {
			pcv := fmt.Sprintf("\033[31m[Causality Violation] Server received Message %d from Client %d. Server VC: %v; Message VC: %v\033[0m\n", clientMessage.messageID, clientMessage.senderID, s.vectorTimeStamp, clientMessage.vectorTimeStamp)
			pcvChannel <- pcv
		}

		s.vectorTimeStamp = mergeVectorClock(s.vectorTimeStamp, clientMessage.vectorTimeStamp, s.pID)
		fmt.Printf("\033[32m(Vector Clock of Server: %v) Server receives Message %d from Client %d\033[0m\n", s.vectorTimeStamp, clientMessage.messageID, clientMessage.senderID)

		event := Event{clientMessage.senderID, 0, clientMessage.messageID, s.vectorTimeStamp, SERVER_RECEIVE_EVENT}
		eventsChannel <- event

		s.vectorTimeStamp[s.pID] += 1

		if rand.Intn(2) == 0 {
			for receiverID := 1; receiverID <= NUM_CLIENTS; receiverID++ {
				if receiverID != clientMessage.senderID {
					serverBroadcastMessage := Message{clientMessage.senderID, clientMessage.messageID, s.vectorTimeStamp}
					go s.serverSender(eventsChannel, serverBroadcastMessage, receiverID)
				}
			}
		} else {
			fmt.Printf("\033[31m(Vector Clock of Server: %v) Server has dropped Message %d from Client %d\033[0m\n", s.vectorTimeStamp, clientMessage.messageID, clientMessage.senderID)
		}

		if clientMessage.messageID == NUM_MESSAGES {
			doneClients++
			if doneClients == NUM_CLIENTS {
				fmt.Println("Server has received all messages.")
				return
			}
		}
	}
}

func (s Server) serverSender(eventsChannel chan Event, serverBroadcastMessage Message, receiverID int) {
	fmt.Printf("\033[38;5;208m(Vector Clock of Server: %v) Server broadcasts Message %d from Client %d to Client %d\033[0m\n", serverBroadcastMessage.vectorTimeStamp, serverBroadcastMessage.messageID, serverBroadcastMessage.senderID, receiverID)
	s.clientsArray[receiverID-1].clientChannel <- serverBroadcastMessage
	event := Event{serverBroadcastMessage.senderID, receiverID, serverBroadcastMessage.messageID, s.vectorTimeStamp, SERVER_BROADCAST_EVENT}
	eventsChannel <- event
}

func (c Client) clientListenerSender(eventsChannel chan Event, pcvChannel chan string) {
	for {
		select {
		case messageID := <-c.readyChannel:
			clientMessage := Message{c.pID, messageID, c.vectorTimeStamp}
			fmt.Printf("\033[34m(Vector Clock of Client %d: %v) Client %d is sending Message %d to Server\033[0m\n", c.pID, c.vectorTimeStamp, c.pID, messageID)
			c.server.serverChannel <- clientMessage
			event := Event{clientMessage.senderID, 0, clientMessage.messageID, clientMessage.vectorTimeStamp, CLIENT_SEND_EVENT}
			eventsChannel <- event

		case serverBroadcastMessage := <-c.clientChannel:
			if lesserVectorClock(serverBroadcastMessage.vectorTimeStamp, c.vectorTimeStamp) {
				pcv := fmt.Sprintf("\033[31m[Causality Violation] Client %d receives Message %d from %d. Client %d's VC: %v; Message VC: %v\033[0m\n", c.pID, serverBroadcastMessage.messageID, serverBroadcastMessage.senderID, c.pID, c.vectorTimeStamp, serverBroadcastMessage.vectorTimeStamp)
				fmt.Println(pcv)
				pcvChannel <- pcv
			}
			c.vectorTimeStamp = mergeVectorClock(c.vectorTimeStamp, serverBroadcastMessage.vectorTimeStamp, c.pID)
			fmt.Printf("(Vector Clock of Client %d: %v) Client %d receives Message %d from Client %d\n", c.pID, c.vectorTimeStamp, c.pID, serverBroadcastMessage.messageID, serverBroadcastMessage.senderID)
			event := Event{serverBroadcastMessage.senderID, c.pID, serverBroadcastMessage.messageID, c.vectorTimeStamp, CLIENT_RECEIVE_EVENT}
			eventsChannel <- event

		case <-c.closeChannel:
			fmt.Printf("Client %d has finished listening for messages.\n", c.pID)
			return

		default:
		}
	}
}

// Function to check if vector timestamp x is less than y
func lesserVectorClock(x []int, y []int) bool {
	for i := range x {
		if x[i] < y[i] {
			return true
		} else if x[i] > y[i] {
			return false
		}
	}
	return false
}

func main() {
	var err error
	// Prompt for number of clients
	for {
		fmt.Print("Enter the number of clients (minimum 2): ")
		_, err = fmt.Scan(&NUM_CLIENTS)
		if err != nil || NUM_CLIENTS < 2 {
			fmt.Println("Invalid input. Please enter at least 2 clients.")
			continue
		}
		break
	}

	// Prompt for number of messages
	for {
		fmt.Print("Enter the number of messages per client (-1 for infinte number of messages): ")
		_, err = fmt.Scan(&NUM_MESSAGES)
		if err != nil || NUM_MESSAGES < -1 {
			fmt.Println("Invalid input. Please enter at least 1 or -1 for infinite messages.")
			continue
		}
		break
	}

	NUM_CLOCKS = NUM_CLIENTS + 1 // One for each client and one for the server

	if NUM_MESSAGES != -1 {
		NUM_EVENTS = (2 + 2*(NUM_CLIENTS-1)) * NUM_CLIENTS * NUM_MESSAGES // Number of events
	} else {
		NUM_EVENTS = (2 + 2*(NUM_CLIENTS-1)) * NUM_CLIENTS * 100 // Assuming a large constant for infinite messages
	}

	clientArray := []*Client{}
	vectorClock := make([]int, NUM_CLOCKS)
	server := Server{0, make(chan Message), clientArray, vectorClock}

	for i := 1; i <= NUM_CLIENTS; i++ {
		client := Client{i, make(chan Message), &server, make(chan bool), make(chan int), make([]int, NUM_CLOCKS)} // Initialize client's vector clock as a slice
		server.clientsArray = append(server.clientsArray, &client)
	}

	eventsChannel := make(chan Event, NUM_EVENTS)
	pcvChannel := make(chan string, NUM_EVENTS)

	// Start all client and server goroutines with wait group
	for _, client := range server.clientsArray {
		wg.Add(1)
		go func(c *Client) {
			defer wg.Done()
			c.clientListenerSender(eventsChannel, pcvChannel)
		}(client)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		server.serverListener(eventsChannel, pcvChannel)
	}()

	for _, client := range server.clientsArray {
		wg.Add(1)
		go func(c *Client) {
			defer wg.Done()
			c.prepMsgs()
		}(client)
	}

	wg.Wait()
	close(pcvChannel)
	close(eventsChannel)

	fmt.Println("Program has finished execution.")
}
