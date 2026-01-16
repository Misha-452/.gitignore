package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

func main() {

	serverAddr := "localhost:8045"
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Printf("Помилка підключення до сервера %s: %v\n", serverAddr, err)
		return
	}
	defer conn.Close()

	fmt.Println("Підключено до сервера. Введіть повідомлення (або 'exit' для виходу):")

	reader := bufio.NewReader(os.Stdin)
	serverReader := bufio.NewReader(conn)

	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "exit" {
			break
		}

		startTime := time.Now()

		_, err := fmt.Fprintln(conn, input)
		if err != nil {
			fmt.Println("Помилка відправки:", err)
			break
		}

		message, err := serverReader.ReadString('\n')
		if err != nil {
			fmt.Println("Помилка отримання відповіді:", err)
			break
		}

		rtt := time.Since(startTime).Milliseconds()

		fmt.Printf("Відповідь сервера: %s", message)
		fmt.Printf("Дані доставлені. RTT: %d ms\n", rtt)
	}
}
