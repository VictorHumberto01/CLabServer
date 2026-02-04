package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

type WSMsg struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

const complexCode = `
#include <stdio.h>
#include <unistd.h>

int main() {
    printf("Small Test: Started\n");
    for(int i=0; i<5; i++) {
        printf("Tick %d\n", i);
        fflush(stdout);
        usleep(100000);
    }
    printf("Small Test: Done\n");
    return 0;
}
`

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}
	log.Printf("Connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	// Read Loop
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			// Just print as string
			fmt.Printf("%s", message)
		}
	}()

	// Send Run Code
	log.Println("Sending Code...")
	msg := WSMsg{
		Type:    "run_code",
		Payload: complexCode,
	}
	jsonMsg, _ := json.Marshal(msg)
	err = c.WriteMessage(websocket.TextMessage, jsonMsg)
	if err != nil {
		log.Println("write:", err)
		return
	}

	// Wait for interrupt or completion (manual)
	select {
	case <-interrupt:
		log.Println("interrupt")
		// Cleanly close the connection by sending a close message and then
		// waiting (with timeout) for the server to close the connection.
		err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Println("write close:", err)
			return
		}
		select {
		case <-done:
		case <-time.After(time.Second):
		}
	case <-time.After(5 * time.Second):
		log.Println("Test finished (timeout)")
		// In real usage we might wait for "Program exited" message closing done
		// for now just exit after enough time
	}
}
