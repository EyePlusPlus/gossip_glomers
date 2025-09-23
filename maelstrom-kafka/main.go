package main

import (
	"context"
	"encoding/json"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type SendMessage struct {
	Type string `json:"type"`
	Key  string `json:"key"`
	Msg  int    `json:"msg"`
}
type PollMessage struct {
	Type    string         `json:"type"`
	Offsets map[string]int `json:"offsets"`
}
type CommitOffsetsMessage struct {
	Type    string         `json:"type"`
	Offsets map[string]int `json:"offsets"`
}
type ListCommitesOffsetsMessage struct {
	Type string   `json:"type"`
	Keys []string `json:"keys"`
}

func appendToLog(kv *maelstrom.KV, ctx context.Context, key string, value int) int {
	success := false
	var newValue []int

	for !success {
		oldValue, err := kv.Read(ctx, key)
		if err != nil || oldValue == nil {
			newValue = []int{value}
		} else {
			z, ok := oldValue.([]interface{})
			if !ok {
				panic("Type is wrong, expected a slice")
			}

			newValue = make([]int, 0, len(z)+1)
			for _, v := range z {
				if f, ok := v.(float64); ok {
					newValue = append(newValue, int(f))
				}
			}
			newValue = append(newValue, value)
		}

		if err := kv.CompareAndSwap(ctx, key, oldValue, newValue, true); err == nil {
			success = true
		}
	}

	return len(newValue) - 1
}

func appendCommitOffsets(kv *maelstrom.KV, ctx context.Context, key string, value int) bool {
	success := false

	for !success {
		oldValue, err := kv.ReadInt(ctx, key)
		if err != nil {
			if maelstrom.ErrorCode(err) != 20 {
				panic("Error reading")
			}
		}

		if err := kv.CompareAndSwap(ctx, key, oldValue, value, true); err == nil {
			success = true
		}
	}

	return true
}

func readFromLog(kv *maelstrom.KV, ctx context.Context, key string, offset int) [][]int {
	returnValue := make([][]int, 0)

	storedValue, err := kv.Read(ctx, key)
	if err != nil {
		if maelstrom.ErrorCode(err) != 20 {
			panic("Failed to read log 1")
		}

		return returnValue
	}

	z, ok := storedValue.([]interface{})
	if !ok {
		panic("Type is wrong")
	}

	for i := offset; i < len(z); i++ {
		v := z[i]
		if f, ok := v.(float64); ok {
			returnValue = append(returnValue, []int{i, int(f)})
		}
	}

	return returnValue
}

func readCommittedOffsets(kv *maelstrom.KV, ctx context.Context, key string) (int, bool) {
	storedValue, err := kv.ReadInt(ctx, key)
	if err != nil {
		if maelstrom.ErrorCode(err) != 20 {
			panic("Failed to read log 2")
		}
		return 0, false
	}

	return storedValue, true
}

func main() {
	n := maelstrom.NewNode()
	kv := maelstrom.NewLinKV(n)
	ctx := context.Background()

	n.Handle("send", func(msg maelstrom.Message) error {
		var body SendMessage

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		key := body.Key
		message := body.Msg

		offset := appendToLog(kv, ctx, key, message)

		res := map[string]any{"type": "send_ok", "offset": offset}

		return n.Reply(msg, res)
	})
	n.Handle("poll", func(msg maelstrom.Message) error {
		var body PollMessage

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		offsets := body.Offsets
		msgs := make(map[string][][]int)

		for k, v := range offsets {
			msgs[k] = readFromLog(kv, ctx, k, v)
		}

		res := make(map[string]any)

		res["type"] = "poll_ok"
		res["msgs"] = msgs

		return n.Reply(msg, res)
	})
	n.Handle("commit_offsets", func(msg maelstrom.Message) error {
		var body CommitOffsetsMessage

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		offsets := body.Offsets

		for k, v := range offsets {
			key := "committed_" + k
			appendCommitOffsets(kv, ctx, key, v)
		}

		res := make(map[string]any)

		res["type"] = "commit_offsets_ok"

		return n.Reply(msg, res)
	})
	n.Handle("list_committed_offsets", func(msg maelstrom.Message) error {
		var body ListCommitesOffsetsMessage

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		keys := body.Keys
		offsets := make(map[string]int)

		for _, k := range keys {
			key := "committed_" + k
			v, ok := readCommittedOffsets(kv, ctx, key)
			if ok {
				offsets[k] = v
			}
		}

		res := make(map[string]any)

		res["type"] = "list_committed_offsets_ok"
		res["offsets"] = offsets

		return n.Reply(msg, res)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
