package main

import (
	"log"
	"net"
	//"os"
)

const PORT = "8080"

type Message struct {
	Type    string
	Conn    net.Conn
	Message string
}

func server(outgoing chan Message) {
	clients := map[string]net.Conn{}
	for {
		msg := <-outgoing
		switch msg.Type {
		case "join":
			clients[msg.Conn.RemoteAddr().String()] = msg.Conn
		case "exit":
			{
				delete(clients, msg.Conn.RemoteAddr().String())
				log.Printf("%s Disconnected", msg.Conn.RemoteAddr().String())
			}
		case "msg":
			for _, c := range clients {
				if c == msg.Conn {
					continue
				}
				//content := fmt.Sprintf("[%s]:%s",msg.Conn.RemoteAddr().String(),msg.Message)
				_, err := c.Write([]byte(msg.Message))
				if err != nil {
					log.Printf("coudnot write message to %s", c.RemoteAddr().String())
				}
			}
		}

	}
}
func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Could not listen to epic port: %s: %s\n", PORT, err.Error())
		//os.Exit(1) //Fatalf use printf asn os.Exit(1)
	}

	log.Printf("Listening to TCP Connection on port %s ...\n", PORT)
	outgoing := make(chan Message)
	go server(outgoing)

	for {
		conn, err := ln.Accept()
		if err != nil {
			//handle err
			log.Printf("ERROR: could not accept a connection: %s\n", err)
			continue
		}

		log.Printf("Accpeted Connection from %s", conn.RemoteAddr())

		go handleConnection(conn, outgoing)
	}
}

func handleConnection(conn net.Conn, outgoing chan Message) {
	defer conn.Close()
	joinMsg := Message{Type: "join", Conn: conn}
	outgoing <- joinMsg

	message := "\033[31mThis text is red\033[0m"

	n, err := conn.Write([]byte(message))
	if err != nil {
		log.Printf("Could not write message to %s: %s\n", conn.RemoteAddr(), err.Error())
	}

	if n < len(message) {
		log.Printf("message was not fully written!  %d:%d \n", n, len(message))
	}

	buffer := make([]byte, 512)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			conn.Close()
			outgoing <- Message{Type: "exit", Conn: conn}
			return
		}

		msg := Message{
			Type:    "msg",
			Conn:    conn,
			Message: string(buffer[0:n]),
		}

		log.Printf("msg: %s",msg.Message)

		outgoing <- msg

	}
}
