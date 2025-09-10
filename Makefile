
.PHONY: playground

export GOBIN=/Users/mansishah/go/bin

playground:
	@echo "Installing playground binary..."
	@(cd ./maelstrom-playground && go install)
	@echo "Running Maelstrom test..."
	@(cd ./maelstrom && ./maelstrom test -w echo --bin ~/go/bin/maelstrom-playground --node-count 1 --time-limit 10)

