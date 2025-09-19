package main

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type TopologyMessage struct {
	Topology map[string][]string `json:"topology"`
}

type BroadcastMessage struct {
	Type    string `json:"type"`
	Message int    `json:"message"`
}

type SyncMessage struct {
	Type   string `json:"type"`
	Values []int  `json:"values"`
}

var stateMutex = sync.RWMutex{}

func getValues(obj map[int]struct{}) []int {
	stateMutex.RLock()
	defer stateMutex.RUnlock()

	retVal := make([]int, 0, len(obj))
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

	n := maelstrom.NewNode()
	data := make(map[int]struct{})
	var neighbors []string

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var body BroadcastMessage

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		message := body.Message

		stateMutex.Lock()
		if _, exists := data[message]; exists {
			stateMutex.Unlock()
			return nil
		}

		data[message] = struct{}{}
		stateMutex.Unlock()

		for _, neighbor := range neighbors {
			if neighbor == msg.Src {
				continue
			}
			n.Send(neighbor, msg.Body)
		}

		res := map[string]string{"type": "broadcast_ok"}

		return n.Reply(msg, res)
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		var body map[string]any

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		result := getValues(data)
		body["type"] = "read_ok"
		body["messages"] = result

		log.Println("Read count", len(result))
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

		stateMutex.Lock()
		data = setValues(data, body.Values)
		stateMutex.Unlock()

		return nil
	})

	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			// time.Sleep(200 * time.Millisecond)

			if len(data) <= 0 {
				continue
			}

			gossip := SyncMessage{Type: "sync", Values: getValues(data)}

			for _, nid := range neighbors {
				if nid != n.ID() {
					n.Send(nid, gossip)
				}
			}
		}
	}()

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
