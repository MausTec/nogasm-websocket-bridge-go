package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
)
import "log"
import "github.com/jacobsa/go-serial/serial"
import "github.com/gorilla/websocket"

// Globals
var addr = flag.String("addr", "localhost:80", "websocket address")
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func serialReadLoop(reader *bufio.Reader, options serial.OpenOptions, done chan string) {
	for {
		data, err := reader.ReadBytes('\n')
		if err != nil {
			fmt.Printf("error reading: %v", err)
			break
		}

		fmt.Print(options.PortName, "> ", string(data))
		//done <- string(data)
	}
	done <- ""
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("upgrade: %v\n", err)
		return
	}

	defer c.Close()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			fmt.Printf("read: %v", err)
			break
		}
		log.Printf("WS> %s", message)
	}
}

func main() {
	// Set up options.
	options := serial.OpenOptions{
		PortName:        "COM32",
		BaudRate:        115200,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
	}

	fmt.Println("Opening port", options.PortName, "at", options.BaudRate, "baud...")
	// Open the port.
	port, err := serial.Open(options)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}
	fmt.Println("Open.")

	// Make sure to close it later.
	defer port.Close()

	// Write 4 bytes to the port.
	//b := []byte{0x00, 0x01, 0x02, 0x03}
	//n, err := port.Write(b)
	//if err != nil {
	//	log.Fatalf("port.Write: %v", err)
	//}

	// Read
	reader := bufio.NewReader(port)
	serialData := make(chan string)
	go serialReadLoop(reader, options, serialData)

	//for {
	//	data := <-serialData // Wait for read loop to return
	//	if data == "" {
	//		break
	//	}
	//	fmt.Print("FWD> ", data)
	//}

	// websocket stuff

	flag.Parse()
	http.HandleFunc("/", websocketHandler)
	err = http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatalf("can't start http server: %v", err)
	}

	fmt.Println("Good bye.")
}
