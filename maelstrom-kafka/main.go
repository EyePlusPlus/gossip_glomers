package main

import (
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

var repl_log = make(map[string][]int)

func appendToLog(key string, value int) int {
	repl_log[key] = append(repl_log[key], value)
	return len(repl_log[key]) - 1
}

func readFromLog(key string, offset int) [][]int {
	returnValue := make([][]int, 0)
	for i := offset; i < len(repl_log[key]); i++ {
		returnValue = append(returnValue, []int{i, repl_log[key][i]})
	}
	return returnValue
}

func main() {
	n := maelstrom.NewNode()

	log.Printf("@@@@@@@@@@ Started")

	n.Handle("send", func(msg maelstrom.Message) error {
		var body SendMessage

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		key := body.Key
		message := body.Msg

		offset := appendToLog(key, message)

		res := make(map[string]any)

		res["type"] = "send_ok"
		res["offset"] = offset

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
			msgs[k] = readFromLog(k, v)
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

		// offsets := body.Offsets

		res := make(map[string]any)

		res["type"] = "commit_offsets_ok"

		return n.Reply(msg, res)
	})
	n.Handle("list_committed_offsets", func(msg maelstrom.Message) error {
		var body ListCommitesOffsetsMessage

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		// keys := body.Keys

		res := make(map[string]any)

		res["type"] = "list_committed_offsets_ok"
		res["offsets"] = nil // TODO

		return n.Reply(msg, res)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
