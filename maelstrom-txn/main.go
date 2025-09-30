package main

import (
	"encoding/json"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type TxnOperation struct {
	Op    string
	Key   int
	Value int
}

type TxnMessage struct {
	Type  string         `json:"type"`
	MsgId int            `json:"msg_id"`
	Txn   []TxnOperation `json:"txn"`
}

var kvStore = make(map[int][]int, 0)

func processOperation(opInstruction TxnOperation) TxnOperation {
	switch opInstruction.Op {
	case "r":
		// Read the value from kv store, update opInstruction, return it
		values, exists := kvStore[opInstruction.Key]
		if exists {
			opInstruction.Value = values[len(values)-1]
		}

	case "w":
		// Write the value to kv store, return opInstruction
		values, exists := kvStore[opInstruction.Key]
		if exists {
			kvStore[opInstruction.Key] = append(values, opInstruction.Value)
		} else {
			kvStore[opInstruction.Key] = []int{opInstruction.Value}
		}

	default:
		panic("Invalid operation")
	}

	return opInstruction
}

func main() {
	n := maelstrom.NewNode()

	n.Handle("txn", func(msg maelstrom.Message) error {
		var body TxnMessage

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		operations := body.Txn

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
