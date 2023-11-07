package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
)

const Listen = ":8080" // Listen address and port
const MaxClients = 256 // Maximum number of clients

// Client represents a connected client (TCP connection).
type Client struct {
	id   int      // client ID
	conn net.Conn // client connection
	nick string   // nickname of the client
}

// Server holds the list of connected clients and a mutex for concurrent access.
type Server struct {
	clients []*Client
	mutex   sync.RWMutex
}

// AddClient adds a new client to the server.
func (s *Server) AddClient(client *Client) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.clients[client.id] = client
}

// FreeClient removes a client from the server.
func (s *Server) FreeClient(client *Client) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.clients[client.id] = nil
	client.conn.Close()
}

// FindAFreeID returns the first free ID in the server
func (s *Server) FindAFreeID() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for i, v := range s.clients {
		if v == nil {
			return i
		}
	}
	// We have no free ID
	return -1
}

// BroadcastMessage sends the message to all connected clients except the sender.
func (s *Server) BroadcastMessage(sender *Client, message []byte) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for _, client := range s.clients {
		if client == sender || client == nil {
			continue
		}
		s.SendMessage(client, message)
	}
}

// SendMessage sends the message to the client
func (s *Server) SendMessage(client *Client, message []byte) error {
	_, err := client.conn.Write(message)
	if err != nil {
		fmt.Printf("Send messages to client %s error: %v", client.nick, err)
		s.FreeClient(client)
	}

	return err
}

func main() {
	server := &Server{
		clients: make([]*Client, MaxClients),
	}

	listener, err := net.Listen("tcp", Listen)
	if err != nil {
		panic(fmt.Sprintf("Error starting server: %v", err))
	}
	defer listener.Close()

	fmt.Println("Server started on ", Listen)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err)
			continue
		}
		freeID := server.FindAFreeID()
		if freeID == -1 {
			_, _ = conn.Write([]byte("Server is full\n"))
			_ = conn.Close()
			continue
		}

		client := &Client{
			id:   freeID,
			conn: conn,
			nick: fmt.Sprintf("user:%d", freeID),
		}
		server.AddClient(client)

		go handleConnection(client, server)
	}
}

// handleConnection handles a client connection
func handleConnection(client *Client, server *Server) {
	fmt.Println("Connected client", client.id)
	welcomeMsg := "Welcome to Simple Chat! Use /nick <nick> to set your nick.\n"
	err := server.SendMessage(client, []byte(welcomeMsg))
	if err != nil {
		return
	}

	buffer := make([]byte, 256)
	for {
		n, err := client.conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Disconnected client", client.nick)
			} else {
				fmt.Printf("Error reading from client %s: %v\n", client.nick, err)
			}
			server.FreeClient(client)
			return
		}

		msg := bytes.TrimRight(buffer[:n], "\r\n")
		if len(msg) > 0 && msg[0] == '/' {
			command := strings.SplitN(string(msg), " ", 2)
			switch command[0] {
			case "/nick":
				if len(command) > 1 {
					client.nick = command[1]
				} else {
					server.SendMessage(client, []byte("Unsupported command\n"))
				}
			case "/quit":
				fmt.Println("Disconnected client", client.nick)
				server.FreeClient(client)
				return
			default:
				server.SendMessage(client, []byte("Unsupported command\n"))
			}
			continue
		}
		sendMsg := fmt.Sprintf("%s> %s\n", client.nick, msg)
		fmt.Printf("%s> %s\n", client.nick, string(msg))
		server.BroadcastMessage(client, []byte(sendMsg))
	}
}
