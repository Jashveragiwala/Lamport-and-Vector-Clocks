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
}

type Client struct {
	clientID      int
	clientChannel chan Message
	server        *Server
	closeChannel  chan bool
}

type Server struct {
	serverChannel chan Message
	clientsArray  []*Client
}

const MESSAGE_DELAY = 500

var NUM_CLIENTS int
var NUM_MESSAGES int

// Periodically each client sends a message to the server
func (c *Client) clientSender(wg *sync.WaitGroup) {
	defer wg.Done()

	msgCount := 1
	for {
		time.Sleep(MESSAGE_DELAY * time.Millisecond)
		fmt.Printf("\033[34mClient %d is sending Message %d\033[0m\n", c.clientID, msgCount)
		c.server.serverChannel <- Message{c.clientID, msgCount}
		// Stop when the client has sent the number of messages
		if NUM_MESSAGES != -1 && msgCount >= NUM_MESSAGES {
			break
		}
		msgCount++
	}
	c.closeChannel <- true
}

// Server listens for messages from clients and broadcasts them to rest of the clients based on a coin toss
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
			fmt.Printf("\033[32mServer received Message %d from Client %d\033[0m\n", clientMessage.messageID, clientMessage.senderID)
			// Server flips a coin to decide whether to broadcast the message or drop it
			coinToss := rand.Intn(2)
			if coinToss == 0 {
				for _, client := range s.clientsArray {
					if client.clientID != clientMessage.senderID {
						go s.serverSender(client, clientMessage)
					}
				}
			} else {
				fmt.Printf("\033[31mServer has dropped Message %d from Client %d\033[0m\n", clientMessage.messageID, clientMessage.senderID)
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

// Server broadcasts messages to rest of the clients
func (s *Server) serverSender(client *Client, msg Message) {
	client.clientChannel <- msg
}

// Clients listen for broadcast messages from the server
func (c *Client) clientListener(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case msg := <-c.clientChannel:
			fmt.Printf("Client %d received Message %d from Client %d\n", c.clientID, msg.messageID, msg.senderID)
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
	server := Server{serverChannel: make(chan Message, 10), clientsArray: make([]*Client, NUM_CLIENTS)}
	var wg sync.WaitGroup

	// Initialize clients and add them to the server's clients array
	for i := 0; i < NUM_CLIENTS; i++ {
		client := Client{
			clientID:      i + 1,
			clientChannel: make(chan Message),
			server:        &server,
			closeChannel:  make(chan bool),
		}
		server.clientsArray[i] = &client
	}

	// Start server listener in a goroutine
	wg.Add(1)
	go server.serverListener(&wg)

	// Small delay to ensure server listener starts before clients send messages
	time.Sleep(200 * time.Millisecond)

	// Each client sends and listens messages
	for _, client := range server.clientsArray {
		wg.Add(1)
		go client.clientListener(&wg)
		wg.Add(1)
		go client.clientSender(&wg)
	}

	// Wait for all goroutines of clients and server to complete
	wg.Wait()
	fmt.Println("All messages processed, program exiting.")
}
