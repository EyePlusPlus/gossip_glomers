
.PHONY: playground

export GOBIN=/Users/mansishah/go/bin

broadcast:
	@echo "Installing playground binary..."
	@(cd ./maelstrom-broadcast && go install)
	@echo "Running Maelstrom test..."
	@(cd ./maelstrom && ./maelstrom test -w broadcast --bin ~/go/bin/maelstrom-broadcast --node-count 5 --time-limit 20 --rate 10)
	@(code /Users/mansishah/personal/recurse/gossip-glomers/maelstrom/store/broadcast/latest/node-logs)

playground:
	@echo "Installing playground binary..."
	@(cd ./maelstrom-playground && go install)
	@echo "Running Maelstrom test..."
	@(cd ./maelstrom && ./maelstrom test -w broadcast --bin ~/go/bin/maelstrom-playground --node-count 2 --time-limit 10)

