package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type InitMessage struct {
	Type    string   `json:"type"`
	NodeID  string   `json:"node_id"`
	NodeIDs []string `json:"node_ids"`
	MsgID   int      `json:"msg_id"`
}
type TxnMessage struct {
	Type  string          `json:"type"`
	MsgId int             `json:"msg_id"`
	Txn   [][]interface{} `json:"txn"`
}

var kvStore = make(map[int]int, 0)
var kvStoreMutex = sync.RWMutex{}

func processOperation(opInstruction []interface{}) []interface{} {
	key, keyOk := opInstruction[1].(float64)
	if !keyOk {
		panic("key is invalid")
	}
	intKey := int(key)

	switch opInstruction[0] {
	case "r":
		values, exists := kvStore[intKey]
		if exists {
			opInstruction[2] = values
		} else {
			opInstruction[2] = nil
		}

	case "w":
		value, valueOk := opInstruction[2].(float64)
		if !valueOk {
			panic("Value is invalid")
		}
		intValue := int(value)
		kvStore[intKey] = intValue

	default:
		panic("Invalid operation")
	}

	return []interface{}{opInstruction[0], opInstruction[1], opInstruction[2]}
}

func main() {
	n := maelstrom.NewNode()

	var neighbors []string

	n.Handle("init", func(msg maelstrom.Message) error {
		var body InitMessage

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		neighbors = body.NodeIDs

		return n.Reply(msg, map[string]any{"type": "init_ok"})

	})

	n.Handle("txn", func(msg maelstrom.Message) error {
		var body TxnMessage

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		operations := body.Txn

		kvStoreMutex.Lock()
		defer kvStoreMutex.Unlock()

		for idx, operation := range operations {
			retVal := processOperation(operation)
			body.Txn[idx] = retVal
		}

		res := map[string]interface{}{"type": "txn_ok", "txn": body.Txn}

		return n.Reply(msg, res)

	})

	n.Handle("sync", func(msg maelstrom.Message) error {
		var body map[string]interface{}

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		operations, ok := body["txn"].([]interface{})
		if !ok {
			return fmt.Errorf("invalid sync message body")
		}

		kvStoreMutex.Lock()
		defer kvStoreMutex.Unlock()

		for _, operation := range operations {
			if operation, ok := operation.([]interface{}); ok {
				processOperation(operation)
			}
		}

		return nil
	})

	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			kvStoreMutex.Lock()
			op := make([][]interface{}, 0, len(kvStore))
			for key, value := range kvStore {
				newOp := []interface{}{"w", key, value}
				op = append(op, newOp)
			}
			kvStoreMutex.Unlock()

			for _, neighbor := range neighbors {
				if n.ID() != neighbor {
					n.Send(neighbor, map[string]interface{}{"type": "sync", "txn": op})
				}
			}
		}
	}()

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}

}
