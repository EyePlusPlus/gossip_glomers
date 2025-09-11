package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type TopologyMessage struct {
	Topology map[string][]string `json:"topology"`
}

type SyncMessage struct {
	Type   string `json:"type"`
	Values []int  `json:"values"`
}

func getValues(obj map[int]struct{}) []int {
	var retVal []int
	for key := range obj {
		retVal = append(retVal, key)
	}
	return retVal
}

func setValues(data map[int]struct{}, values []int) map[int]struct{} {
	for _, v := range values {
		data[v] = struct{}{}
	}

	return data
}

func main() {

	file, err := os.OpenFile("/tmp/maelstrom-broadcast.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}

	defer file.Close()

	n := maelstrom.NewNode()
	data := make(map[int]struct{})
	var neighbors []string

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var body map[string]any

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		if msgFloat, ok := body["message"].(float64); ok {
			data[int(msgFloat)] = struct{}{}
		}

		gossip := SyncMessage{Type: "sync", Values: getValues(data)}

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
		body["messages"] = getValues(data)

		return n.Reply(msg, body)
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		var body TopologyMessage

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		neighbors = body.Topology[n.ID()]

		res := map[string]string{"type": "topology_ok"}

		return n.Reply(msg, res)
	})

	n.Handle("sync", func(msg maelstrom.Message) error {
		var body SyncMessage

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		data = setValues(data, body.Values)

		return n.Reply(msg, map[string]any{"type": "sync_ok"})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
