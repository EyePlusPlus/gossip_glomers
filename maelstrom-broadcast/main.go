package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

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

var repl_log_mutex = sync.RWMutex{}
var ack_mutex = sync.RWMutex{}

func getValues(obj map[int]struct{}) []int {
	repl_log_mutex.RLock()
	defer repl_log_mutex.RUnlock()

	var retVal []int
	for key := range obj {
		retVal = append(retVal, key)
	}
	return retVal
}

func setValues(data map[int]struct{}, values []int) map[int]struct{} {
	repl_log_mutex.Lock()
	defer repl_log_mutex.Unlock()

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
	sync_ack := make(map[string]struct{})
	var neighbors []string

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var body map[string]any

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		repl_log_mutex.Lock()
		if msgFloat, ok := body["message"].(float64); ok {
			data[int(msgFloat)] = struct{}{}
		}
		repl_log_mutex.Unlock()

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

		res := map[string]string{"type": "topology_ok"}

		return n.Reply(msg, res)
	})

	n.Handle("sync", func(msg maelstrom.Message) error {
		var body SyncMessage

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		ack_mutex.RLock()
		_, exists := sync_ack[body.Id]
		ack_mutex.RUnlock()

		if exists {
			return n.Reply(msg, map[string]any{"type": "sync_ok"})
		}

		data = setValues(data, body.Values)

		ack_mutex.Lock()
		sync_ack[body.Id] = struct{}{}
		ack_mutex.Unlock()

		for _, nid := range neighbors {
			if nid != msg.Src {
				n.RPC(nid, body, nil)
			}
		}

		return n.Reply(msg, map[string]any{"type": "sync_ok"})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
