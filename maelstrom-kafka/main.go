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

func main() {
	n := maelstrom.NewNode()

	log.Printf("@@@@@@@@@@ Started")

	n.Handle("send", func(msg maelstrom.Message) error {
		var body SendMessage

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		// key := body.Key
		// message := body.Msg

		// TODO

		res := make(map[string]any)

		res["type"] = "send_ok"
		res["offset"] = 0 // TODO

		return n.Reply(msg, res)
	})
	n.Handle("poll", func(msg maelstrom.Message) error {
		var body PollMessage

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		// offsets := body.Offsets

		res := make(map[string]any)

		res["type"] = "poll_ok"
		res["msgs"] = nil // TODO

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
