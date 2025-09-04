package main

import (
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

func appendToFile(file, data) {
	if _, err := file.WriteString(data + "\n"); err != nil {
		fmt.Printf("Error writing to file\n", err)
		return
	}
}

func storeValue(kv maelstrom.KV, value int) bool {
	counter, err := kv.Read(key)
	if err != nil {
		log.Fatal("Reading from kv failed", err)
		return false
	}

	newValue := counter + value

	if err := kv.CompareAndSwap(key, counter, newValue, true); err != nil {
		log.Fatal("Failed to compare and swap the value", err)
		return false
	}

	return true

}

func main() {
	n := maelstrom.NewNode()
	kv := maelstrom.NewKV(n)
	file, err := os.OpenFile("/tmp/maelstrom-g-counter.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening file", err)
		return
	}

	n.Handle("add", func(msg maelstrom.Message) error {
		var body map[string]any
		appendToFile(file, "addHandler:  Received msg: "+msg)

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

		success := storeValue(kv, value)
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

		appendToFile(file, fmt.Sprintf("addHandler:  Response: ", res))
		return n.Reply(msg, res)
	})
}
