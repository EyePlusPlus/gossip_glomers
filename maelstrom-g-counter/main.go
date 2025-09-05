package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type TopologyResponse map[string][]string

type InitMessage struct {
	Type    string   `json:"type"`
	NodeID  string   `json:"node_id"`
	NodeIDs []string `json:"node_ids"`
	MsgID   int      `json:"msg_id"`
}

func appendToFile(file *os.File, data string) {
	if _, err := file.WriteString(data + "\n"); err != nil {
		fmt.Printf("Error writing to file: %s\n", err.Error())
		return
	}
}

func storeValue(kv *maelstrom.KV, ctx context.Context, key string, value int) (bool, error) {
	counter, err := kv.ReadInt(ctx, key)
	if err != nil {
		counter = 0
	}

	newValue := counter + value

	if err := kv.CompareAndSwap(ctx, key, counter, newValue, true); err != nil {
		return false, err
	}

	return true, nil

}

func getValue(kv *maelstrom.KV, ctx context.Context, node_ids []string) (int, error) {
	result := 0
	for _, k := range node_ids {
		counter, err := kv.ReadInt(ctx, k)
		if err != nil {
			result += 0
		} else {
			result += counter
		}
	}

	return result, nil
}

func main() {
	n := maelstrom.NewNode()
	kv := maelstrom.NewSeqKV(n)
	ctx := context.Background()
	var node_id string

	var node_ids []string

	file, err := os.OpenFile("/tmp/maelstrom-g-counter.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening file %s", err.Error())
		return
	}

	defer file.Close()

	n.Handle("init", func(msg maelstrom.Message) error {
		var initMsg InitMessage
		if err := json.Unmarshal(msg.Body, &initMsg); err != nil {
			return err
		}

		// appendToFile(file, "initHandler: response"+string(msg.Body))

		node_id = initMsg.NodeID
		node_ids = initMsg.NodeIDs

		if len(node_ids) == 0 {
			appendToFile(file, "initHandler: No nodeids found")
		}

		body := map[string]any{
			"type": "init_ok",
		}

		return n.Reply(msg, body)
	})

	n.Handle("add", func(msg maelstrom.Message) error {
		var body map[string]any

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		var value int

		if delta, ok := body["delta"].(int); ok {
			value = int(delta)
		}

		if delta, ok := body["delta"].(float64); ok {
			value = int(delta)
		}

		_, err := storeValue(kv, ctx, node_id, value)
		if err != nil {
			appendToFile(file, "addHandler: ("+string(node_id)+")storeValue error"+err.Error())
		}
		// if success {
		// 	appendToFile(file, "addHandler: storing value succeeded, "+node_id+", "+strconv.Itoa(value))
		// }

		res := map[string]any{
			"type": "add_ok",
		}

		// appendToFile(file, fmt.Sprintf("addHandler:  Response: %s", res))
		return n.Reply(msg, res)
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		body := make(map[string]any)

		body["type"] = "read_ok"
		val, err := getValue(kv, ctx, node_ids)
		body["value"] = val

		if err != nil {
			appendToFile(file, "readHandler: getValue error"+err.Error())
		}
		return n.Reply(msg, body)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}

}
