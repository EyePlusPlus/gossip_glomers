package main

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type TopologyMessage struct {
	Topology map[string][]string `json:"topology"`
}

type SyncMessage struct {
	Type   string `json:"type"`
	Values []int  `json:"values"`
	Id     string `json:"id"`
}

var stateMutex = sync.RWMutex{}

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

	n := maelstrom.NewNode()
	data := make(map[int]struct{})
	sync_ack := make(map[string]struct{})
	var neighbors []string

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var body map[string]any

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		stateMutex.Lock()
		defer stateMutex.Unlock()

		if msgFloat, ok := body["message"].(float64); ok {
			data[int(msgFloat)] = struct{}{}
		}

		sync_id, err := uuid.NewV7()
		if err != nil {
			return err
		}

		sync_ack[sync_id.String()] = struct{}{}

		gossip := SyncMessage{Type: "sync", Values: getValues(data), Id: sync_id.String()}

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
		// log.Printf("** %s can talk to %v\n", n.ID(), neighbors)

		res := map[string]string{"type": "topology_ok"}

		return n.Reply(msg, res)
	})

	n.Handle("sync", func(msg maelstrom.Message) error {
		stateMutex.Lock()
		defer stateMutex.Unlock()
		var body SyncMessage

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		if _, exists := sync_ack[body.Id]; exists {
			return n.Reply(msg, map[string]any{"type": "sync_ok"})
		}

		data = setValues(data, body.Values)

		sync_ack[body.Id] = struct{}{}

		for _, nid := range neighbors {
			if nid != msg.Src {
				n.RPC(nid, body, nil)
			}
		}

		return n.Reply(msg, map[string]any{"type": "sync_ok"})
	})

	go func() {
		for {
			time.Sleep(1 * time.Second)

			sync_id, err := uuid.NewV7()
			if err != nil {
				continue
			}

			stateMutex.Lock()
			sync_ack[sync_id.String()] = struct{}{}
			stateMutex.Unlock()

			gossip := SyncMessage{Type: "sync", Values: getValues(data), Id: sync_id.String()}

			for _, nid := range neighbors {
				if nid != n.ID() {
					n.RPC(nid, gossip, nil)
				}
			}
		}
	}()

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
