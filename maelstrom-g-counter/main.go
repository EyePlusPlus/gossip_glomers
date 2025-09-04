package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type TopologyResponse map[string][]string

func appendToFile(file *os.File, data string) {
	if _, err := file.WriteString(data + "\n"); err != nil {
		fmt.Printf("Error writing to file\n", err)
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

func getValue(kv *maelstrom.KV, ctx context.Context, topologyResponse TopologyResponse) (int, error) {
	result := 0
	for k, _ := range topologyResponse {
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

	node_id := n.ID()

	var topologyResponse TopologyResponse

	file, err := os.OpenFile("/tmp/maelstrom-g-counter.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening file", err)
		return
	}

	defer file.Close()

	n.Handle("topology", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		topologyResponse = body["topology"].(TopologyResponse)

		resp := make(map[string]any)
		resp["type"] = "topology_ok"

		return n.Reply(msg, resp)

	})

	n.Handle("add", func(msg maelstrom.Message) error {
		appendToFile(file, "Received add msg")

		var body map[string]any
		appendToFile(file, "addHandler:  Received msg: "+string(msg.Body))

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

		appendToFile(file, "addHandler:  delta value: "+strconv.Itoa(value))

		success, err := storeValue(kv, ctx, node_id, value)
		if err != nil {
			appendToFile(file, "addHandler: storeValue error"+err.Error())
		}
		if !success {
			appendToFile(file, "addHandler: storing value failed")
		}

		// Extract the "delta" value from the message
		// Read the current global counter value from kv
		// Add the values and compareAndSwap to store the new value
		// If the function throws an error, retry using the latest value

		res := map[string]any{
			"type": "add_ok",
		}

		appendToFile(file, fmt.Sprintf("addHandler:  Response: %s", res))
		return n.Reply(msg, res)
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		appendToFile(file, "Received read msg")
		body := make(map[string]any)
		appendToFile(file, "readHandler:  Received msg: "+string(msg.Body))

		body["type"] = "read_ok"
		val, err := getValue(kv, ctx, topologyResponse)
		body["value"] = val

		if err != nil {
			appendToFile(file, "readHandler: getValue error"+err.Error())
		} else {
			appendToFile(file, "readHandler: getValue value"+strconv.Itoa(val))
		}
		return n.Reply(msg, body)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}

}
