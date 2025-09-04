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

const (
	key = "GLOBAL_COUNTER"
)

func appendToFile(file *os.File, data string) {
	if _, err := file.WriteString(data + "\n"); err != nil {
		fmt.Printf("Error writing to file\n", err)
		return
	}
}

func storeValue(kv *maelstrom.KV, ctx context.Context, value int) (bool, error) {
	counter, err := kv.ReadInt(ctx, key)
	if err != nil {
		return false, err
	}

	newValue := counter + value

	if err := kv.CompareAndSwap(ctx, key, counter, newValue, true); err != nil {
		return false, err
	}

	return true, nil

}

func getValue(kv *maelstrom.KV, ctx context.Context) (int, error) {
	counter, err := kv.ReadInt(ctx, key)
	if err != nil {
		return 0, err
	}

	return counter, nil
}

func main() {
	n := maelstrom.NewNode()
	kv := maelstrom.NewSeqKV(n)
	ctx := context.Background()

	file, err := os.OpenFile("/tmp/maelstrom-g-counter.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening file", err)
		return
	}

	defer file.Close()

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

		success, err := storeValue(kv, ctx, value)
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
		var body map[string]any
		appendToFile(file, "readHandler:  Received msg: "+string(msg.Body))

		body["type"] = "read_ok"
		body["value"], err = getValue(kv, ctx)
		appendToFile(file, "readHandler: getValue error"+err.Error())
		if err != nil {
		}

		return n.Reply(msg, body)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}

}
