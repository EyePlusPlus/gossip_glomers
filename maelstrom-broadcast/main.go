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

type BroadcastMessage struct {
	Type    string `json:"type"`
	Message int    `json:"message"`
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

func setValues(data map[int]struct{}, values []int, pending_messages []int) (map[int]struct{}, []int) {
	for _, v := range values {
		if _, exists := data[v]; !exists {
			data[v] = struct{}{}
			pending_messages = append(pending_messages, v)
		}

	}

	return data, pending_messages
}

func main() {

	n := maelstrom.NewNode()
	data := make(map[int]struct{})
	sync_ack := make(map[string]struct{})
	var neighbors []string
	pending_messages := make([]int, 0)

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var body BroadcastMessage

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		message := body.Message

		stateMutex.Lock()

		data[message] = struct{}{}
		pending_messages = append(pending_messages, message)
		stateMutex.Unlock()

		res := map[string]string{"type": "broadcast_ok"}

		return n.Reply(msg, res)
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		var body map[string]any

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		stateMutex.RLock()
		defer stateMutex.RUnlock()

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
		if _, exists := sync_ack[body.Id]; exists {
			return nil
		}

		log.Println("Length pending before", len(pending_messages))
		data, pending_messages = setValues(data, body.Values, pending_messages)
		log.Println("Length pending after", len(pending_messages))
		stateMutex.Unlock()

		return nil
	})

	go func() {
		for {
			time.Sleep(200 * time.Millisecond)

			if len(pending_messages) <= 0 {
				continue
			}

			sync_id, err := uuid.NewV7()
			if err != nil {
				continue
			}

			stateMutex.Lock()
			sync_ack[sync_id.String()] = struct{}{}
			result := make([]int, len(pending_messages))
			copy(result, pending_messages)
			pending_messages = make([]int, 0)
			stateMutex.Unlock()

			gossip := SyncMessage{Type: "sync", Values: result, Id: sync_id.String()}

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
