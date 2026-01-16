package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

type server struct {
	portToUsername map[string]string
	history        []string
}

const (
	HOST                  = "localhost"
	PORT                  = "8080"
	TYPE                  = "tcp"
	CLIENT_CHANNEL_BUFFER = 10
)

var (
	messages = make(chan string)
	joining  = make(chan chan<- string)
	leaving  = make(chan chan<- string)
)

func (s *server) manager() {
	clients := make(map[chan<- string]bool)
	for {
		select {
		case msg := <-messages:
			s.history = append(s.history, msg)
			for clientChan := range clients {

				select {
				case clientChan <- msg:
				default:
					leaving <- clientChan

					delete(clients, clientChan)
					fmt.Printf("Клієнт відключений через переповнення буфера.\n")
				}
			}

		case clientChan := <-joining:
			clients[clientChan] = true
			fmt.Printf("ЧАТ: Новий клієнт приєднався. Активних:%d\n", len(clients))

		case clientChan := <-leaving:
			delete(clients, clientChan)
			close(clientChan)
			fmt.Printf("ЧАТ: Клієнт відключився. Активних:%d\n", len(clients))
		}
	}
}

func (s *server) clientWriter(conn net.Conn, out <-chan string) {
	defer conn.Close()
	for msg := range out {
		_, err := conn.Write([]byte(msg))
		if err != nil {
			break
		}
	}
}

func trimFirstLine(s string, maxLines int) string {
	lines := strings.Split(s, "\n")

	if len(lines) > maxLines {
		lines = lines[1:]
	}

	return strings.Join(lines, "\n")
}

func (s *server) handleConnection(conn net.Conn) {
	outgoing := make(chan string, CLIENT_CHANNEL_BUFFER)
	go s.clientWriter(conn, outgoing)
	joining <- outgoing
	welcome := "Добро пожаловать на сервер! enter your nikname\n"
	conn.Write([]byte(welcome))

	input := bufio.NewScanner(conn)
	userName := ""

	_ = input.Scan()
	userName = input.Text()

	for input.Scan() {
		text := input.Text()

		if text == "history" {
			for _, msg := range s.history {
				conn.Write([]byte(msg))
			}
			continue
		}
		messages <- fmt.Sprintf("[%s]: %s\n", userName, input.Text())

	}

	leaving <- outgoing
}
func main() {

	s := server{
		portToUsername: make(map[string]string),
	}
	go s.manager()
	listener, err := net.Listen(TYPE, HOST+":"+PORT)
	if err != nil {

		fmt.Fprintf(os.Stderr, "Помилка запуску сервера: %v\n", err)

		os.Exit(1)
	}
	defer listener.Close()
	fmt.Printf("TCP Chat Сервер запущено на %s:%s. Очікування клієнтів...\n", HOST, PORT)
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Помилка Accept: %v\n", err)
			continue
		}
		go s.handleConnection(conn)
	}
}
