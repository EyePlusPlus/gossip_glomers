package main

import (
	"encoding/json"
	"fmt"
	"log"
	"maps"
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
	var neighbors []string

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var body map[string]any

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		gossip := maps.Clone(body)
		
		if msgFloat, ok := body["message"].(float64); ok {
			data = append(data, int(msgFloat))
		}

		for _, nid := range neighbors {
			if nid != msg.Src {
				n.RPC(nid, gossip, nil)
			}
		}

		res := map[string]string{"type": "broadcast_ok"}

		return n.Reply(msg, res)
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		var body map[string]any

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		body["type"] = "read_ok"
		body["messages"] = data

		return n.Reply(msg, body)
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		var body map[string]any

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		topology, ok := body["topology"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("malformed topology message: 'topology' field is not a map")
		}

		if neighbors_any, ok := topology[n.ID()]; ok {
			if neighbors_interface, ok := neighbors_any.([]interface{}); ok {
				var tempNeighbors []string
				for _, neighbor_any := range neighbors_interface {
					if neighbor_str, ok := neighbor_any.(string); ok {
						tempNeighbors = append(tempNeighbors, neighbor_str)
					}
				}
				neighbors = tempNeighbors
			}
		}

		res := map[string]string{"type": "topology_ok"}

		return n.Reply(msg, res)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
