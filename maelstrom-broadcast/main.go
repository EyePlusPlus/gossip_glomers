package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"slices"
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

func main() {

	n := maelstrom.NewNode()
	var neighbors []string
	var pending_queue []int

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var body BroadcastMessage

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		message := body.Message

		if !slices.Contains(pending_queue, message) {
			stateMutex.Lock()
			pending_queue = append(pending_queue, message)
			stateMutex.Unlock()

			for _, neighbor := range neighbors {
				if neighbor == msg.Src {
					continue
				}
				n.Send(neighbor, msg.Body)
			}
		}

		res := map[string]string{"type": "broadcast_ok"}

		return n.Reply(msg, res)
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		// var body map[string]any
		body := map[string]any{}

		// if err := json.Unmarshal(msg.Body, &body); err != nil {
		// 	return err
		// }

		body["type"] = "read_ok"
		body["messages"] = pending_queue

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

		for _, v := range body.Values {
			if !slices.Contains(pending_queue, v) {
				stateMutex.Lock()
				pending_queue = append(pending_queue, v)
				stateMutex.Unlock()
			}
		}

		return nil
	})

	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			// time.Sleep(200 * time.Millisecond)

			if len(pending_queue) <= 0 {
				continue
			}

			stateMutex.Lock()
			gossip := SyncMessage{Type: "sync", Values: pending_queue}
			stateMutex.Unlock()

			neighbor := neighbors[rand.Intn(len(neighbors))]

			// for _, neighbor := range neighbors {
			// 	if neighbor != n.ID() {
			n.Send(neighbor, gossip)
			// 	}
			// }
		}
	}()

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
