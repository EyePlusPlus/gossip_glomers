
.PHONY: playground

export GOBIN=/Users/mansishah/go/bin

playground:
	@echo "Installing playground binary..."
	@(cd ./maelstrom-playground && go install)
	@echo "Running Maelstrom test..."
	@(cd ./maelstrom && ./maelstrom test -w broadcast --bin ~/go/bin/maelstrom-playground --node-count 2 --time-limit 10)

