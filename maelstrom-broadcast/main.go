package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func appendToFile(file *os.File, msg string) {
	if _, err := file.WriteString(msg + "\n"); err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		return
	}
}

func main() {

	file, err := os.OpenFile("/tmp/maelstrom-broadcast.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}

	defer file.Close()

	n := maelstrom.NewNode()
	var data []int

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var body map[string]any

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		body["type"] = "broadcast_ok"
		if msgInt, ok := body["message"].(int); ok {
			data = append(data, int(msgInt))
		}

		if msgFloat, ok := body["message"].(float64); ok {
			data = append(data, int(msgFloat))
		}

		// appendToFile(file, string(msg.Body))

		res := map[string]string{"type": "broadcast_ok"}

		return n.Reply(msg, res)
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		var body map[string]any

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		// appendToFile(file, string(msg.Body))

		body["type"] = "read_ok"
		body["messages"] = data

		return n.Reply(msg, body)
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		var body map[string]any

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		// appendToFile(file, string(msg.Body))

		res := map[string]string{"type": "topology_ok"}

		return n.Reply(msg, res)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
