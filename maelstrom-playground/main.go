package main

import (
	"encoding/json"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()

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

		log.Printf("Node %s received 'test' message: %v", n.ID(), body)

		body["type"] = "test_ok"

		return n.Reply(msg, body)
	})

	// This handler is called when the node is initialized.
	n.Handle("init", func(msg maelstrom.Message) error {
		log.Printf("Node %s initialized with node IDs: %v", n.ID(), n.NodeIDs())

		// Find another node to send a message to.
		var dest string
		for _, id := range n.NodeIDs() {
			if id != n.ID() {
				dest = id
				break
			}
		}

		// If there is another node, send it a message.
		if dest != "" {
			log.Printf("Node %s is sending a message to node %s", n.ID(), dest)
			msgBody := map[string]any{
				"type":   "test",
				"value":  "Hello from " + n.ID(),
				"msg_id": 1,
			}
			// Using RPC to send and get a reply.
			err := n.RPC(dest, msgBody, func(reply maelstrom.Message) error {
				log.Printf("Node %s received reply from %s: %s", n.ID(), dest, reply.Body)
				return nil
			})

			if err != nil {
				log.Printf("Error calling RPC: %s", err)
			}
		}

		return nil
	})

	if err := n.Run(); err != nil {
		log.Fatal(err.Error())
	}
}
