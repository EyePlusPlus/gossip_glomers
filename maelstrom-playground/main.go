package main

import (
	"encoding/json"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type Message struct {
	Type      string `json:"type"`
	Value     any    `json:"value"`
	MessageID int    `json:"msg_id"`
}

func main() {
	counter := 0
	n := maelstrom.NewNode()

	node_ids := n.NodeIDs()

	for _, k := range node_ids {
		msg := Message{
			Type:      "test",
			Value:     "Hello",
			MessageID: counter,
		}
		jsonBody, err := json.Marshal(msg)

		if err != nil {
			log.Fatal(err.Error())
		}

		n.Send(k, jsonBody)
		counter = counter + 1
	}

	n.Handle("echo", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		body["type"] = "echo_ok"

		return n.Reply(msg, body)
	})

	n.Handle("test", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		body["type"] = "test_ok"

		return n.Reply(msg, body)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err.Error())
	}
}
