package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Message struct {
	senderID  int
	messageID int
	clock     int
}

type Client struct {
	clientID      int
	clientChannel chan Message
	server        *Server
	closeChannel  chan bool
	lamportClock  int
}

type Server struct {
	serverChannel chan Message
	clientsArray  []*Client
	lamportClock  int
}

const MESSAGE_DELAY = 1000

var NUM_CLIENTS int
var NUM_MESSAGES int

// Update and return Lamport clock
func updateLamportClock(localClock, remoteClock int) int {
	return max(localClock, remoteClock) + 1
}

// Helper to get max of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Periodically each client sends a message to the server
func (c *Client) clientSender(wg *sync.WaitGroup) {
	defer wg.Done()

	msgCount := 1
	for {
		time.Sleep(MESSAGE_DELAY * time.Millisecond)
		c.lamportClock++
		fmt.Printf("\033[34m(Lamport Clock of Client %d: %d) Client %d  is sending Message %d to Server\033[0m\n", c.clientID, c.lamportClock, c.clientID, msgCount)
		c.server.serverChannel <- Message{c.clientID, msgCount, c.lamportClock}
		// Stop when the client has sent the number of messages
		if NUM_MESSAGES != -1 && msgCount >= NUM_MESSAGES {
			break
		}
		msgCount++
	}
	c.closeChannel <- true
}

// Server broadcasts messages to clients concurrently
func (s *Server) serverSender(client *Client, msg Message) {
	// Clients update their Lamport clock correctly using the server Lamport clock
	client.lamportClock = updateLamportClock(client.lamportClock, msg.clock)
	// Send the message with the server updated clock to the client
	client.clientChannel <- Message{msg.senderID, msg.messageID, client.lamportClock}
}

// Server listens for messages from clients and broadcasts them concurrently
func (s *Server) serverListener(wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("Server is listening for Messages...")

	doneClients := 0
	for {
		select {
		case clientMessage, ok := <-s.serverChannel:
			if !ok {
				return
			}
			// Update server Lamport clock when receiving a message
			s.lamportClock = updateLamportClock(s.lamportClock, clientMessage.clock)
			fmt.Printf("\033[32m(Lamport Clock of Server: %d) Server received Message %d from Client %d\033[0m\n", s.lamportClock, clientMessage.messageID, clientMessage.senderID)
			// Server flips a coin to decide whether to broadcast the message or drop it
			coinToss := rand.Intn(2)
			s.lamportClock++
			if coinToss == 0 {
				fmt.Printf("\033[38;5;214m(Lamport Clock of Server: %d) Server is forwarding message %d from Client %d\033[0m\n", s.lamportClock, clientMessage.messageID, clientMessage.senderID)
				for _, client := range s.clientsArray {
					if client.clientID != clientMessage.senderID {
						go s.serverSender(client, Message{clientMessage.senderID, clientMessage.messageID, s.lamportClock})
					}
				}
			} else {
				fmt.Printf("\033[31m(Lamport Clock of Server: %d) Server has dropped Message %d from Client %d\033[0m\n", s.lamportClock, clientMessage.messageID, clientMessage.senderID)
			}
			if clientMessage.messageID == NUM_MESSAGES {
				doneClients++
				if doneClients == NUM_CLIENTS {
					close(s.serverChannel) // Close server channel when all clients are done
				}
			}
		}
	}
}

// Clients listen for broadcast messages from the server
func (c *Client) clientListener(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case msg := <-c.clientChannel:
			// Update client Lamport clock when receiving a message
			c.lamportClock = updateLamportClock(c.lamportClock, msg.clock)
			fmt.Printf("(Lamport Clock of Client %d: %d) Client %d received Message %d from Client %d\n", c.clientID, c.lamportClock, c.clientID, msg.messageID, msg.senderID)
		case <-c.closeChannel:
			fmt.Printf("Client %d is done listening for Messages...\n", c.clientID)
			return
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	// Get user input for number of clients
	for {
		fmt.Print("Enter the number of clients (at least 2): ")
		fmt.Scan(&NUM_CLIENTS)
		if NUM_CLIENTS < 2 {
			fmt.Println("You need at least 2 clients to communicate.")
		} else {
			break
		}
	}

	// Get user input for number of messages per client
	for {
		fmt.Print("Enter the number of messages each client will send (enter -1 for infinite messages): ")
		fmt.Scan(&NUM_MESSAGES)
		if NUM_MESSAGES == 0 {
			fmt.Println("There is nothing to communicate. Please enter a valid number of messages.")
		} else {
			break
		}
	}
	// Initialize server
	server := Server{serverChannel: make(chan Message, 10), clientsArray: make([]*Client, NUM_CLIENTS), lamportClock: 0}
	var wg sync.WaitGroup

	// Initialize clients and add them to the server's clients array
	for i := 0; i < NUM_CLIENTS; i++ {
		client := Client{
			clientID:      i + 1,
			clientChannel: make(chan Message),
			server:        &server,
			closeChannel:  make(chan bool),
			lamportClock:  0,
		}
		server.clientsArray[i] = &client
	}

	// Start server listener in a goroutine
	wg.Add(1)
	go server.serverListener(&wg)

	for _, client := range server.clientsArray {
		wg.Add(1)
		go client.clientListener(&wg)
		wg.Add(1)
		go client.clientSender(&wg)
	}

	// Wait for all goroutines (clients and server) to complete
	wg.Wait()
	fmt.Println("All messages processed, program exiting.")
}
