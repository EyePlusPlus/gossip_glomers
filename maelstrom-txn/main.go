package main

import (
	"encoding/json"
	"log"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type TxnMessage struct {
	Type  string          `json:"type"`
	MsgId int             `json:"msg_id"`
	Txn   [][]interface{} `json:"txn"`
}

var kvStore = make(map[int][]int, 0)
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
			opInstruction[2] = values[len(values)-1]
		} else {
			opInstruction[2] = nil
		}

	case "w":
		value, valueOk := opInstruction[2].(float64)
		if !valueOk {
			panic("Value is invalid")
		}
		intValue := int(value)
		values, exists := kvStore[intKey]
		if exists {
			kvStore[intKey] = append(values, intValue)
		} else {
			kvStore[intKey] = []int{intValue}
		}

	default:
		panic("Invalid operation")
	}

	return []interface{}{opInstruction[0], opInstruction[1], opInstruction[2]}
}

func main() {
	n := maelstrom.NewNode()

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

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}

}
