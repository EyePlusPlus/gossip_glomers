
.PHONY: echo unique broadcast gcounter kafka txn playground

export GOBIN=/Users/mansishah/go/bin

echo:
	@echo "Installing echo binary..."
	@(cd ./maelstrom-echo && go install)
	@echo "Running Maelstrom test..."
	@(cd ./maelstrom && ./maelstrom test -w echo --bin ~/go/bin/maelstrom-echo --node-count 1 --time-limit 10 &> /dev/null)
	@echo "Passed echo"


unique:
	@echo "Installing unique-id binary..."
	@(cd ./maelstrom-unique-ids && go install)
	@echo "Running Maelstrom test..."
	@(cd ./maelstrom && ./maelstrom test -w unique-ids --bin ~/go/bin/maelstrom-unique-ids --time-limit 30 --rate 1000 --node-count 3 --availability total --nemesis partition &> /dev/null)
	@echo "Passed unique-ids"


broadcast:
	@echo "Installing broadcast binary..."
	@(cd ./maelstrom-broadcast && go install)
	@echo "Running Maelstrom test..."
	@(cd ./maelstrom && ./maelstrom test -w broadcast --bin ~/go/bin/maelstrom-broadcast --node-count 1 --time-limit 20 --rate 10  &> /dev/null)
	@echo "Passed broadcast-a"
	@(cd ./maelstrom && ./maelstrom test -w broadcast --bin ~/go/bin/maelstrom-broadcast --node-count 5 --time-limit 20 --rate 10 &> /dev/null)
	@echo "Passed broadcast-b"
	@(cd ./maelstrom && ./maelstrom test -w broadcast --bin ~/go/bin/maelstrom-broadcast --node-count 5 --time-limit 20 --rate 10 --nemesis partition &> /dev/null)
	@echo "Passed broadcast-c"
	@(cd ./maelstrom && ./maelstrom test -w broadcast --bin ~/go/bin/maelstrom-broadcast --node-count 25 --time-limit 20 --rate 100 --latency 100)
	@echo "Passed broadcast-d"


gcounter:
	@echo "Installing g-counter binary..."
	@(cd ./maelstrom-g-counter && go install)
	@echo "Running Maelstrom test..."
	@(cd ./maelstrom && ./maelstrom test -w g-counter --bin ~/go/bin/maelstrom-g-counter --node-count 3 --rate 100 --time-limit 20 --nemesis partition)
	@echo "Passed g-counter"

kafka:
	@echo "Installing kafka binary..."
	@(cd ./maelstrom-kafka && go install)
	@echo "Running Maelstrom test..."
# 	@(cd ./maelstrom && ./maelstrom test -w kafka --bin ~/go/bin/maelstrom-kafka --node-count 1 --concurrency 2n --time-limit 20 --rate 1000)
# 	@echo "Passed kafka-a"
	@(cd ./maelstrom && ./maelstrom test -w kafka --bin ~/go/bin/maelstrom-kafka --node-count 2 --concurrency 2n --time-limit 20 --rate 1000)

txn:
	@echo "Installing txn binary..."
	@(cd ./maelstrom-txn && go install)
	@echo "Running Maelstrom test..."
# 	@(cd ./maelstrom && ./maelstrom test -w txn-rw-register --bin ~/go/bin/maelstrom-txn --node-count 1 --time-limit 20 --rate 1000 --concurrency 2n --consistency-models read-uncommitted --availability total)
# 	@echo "Passed txn-a"
	@(cd ./maelstrom && ./maelstrom test -w txn-rw-register --bin ~/go/bin/maelstrom-txn --node-count 2 --concurrency 2n --time-limit 20 --rate 1000 --consistency-models read-uncommitted)
	@echo "Passed txn-b"

playground:
	@echo "Installing playground binary..."
	@(cd ./maelstrom-playground && go install)
	@echo "Running Maelstrom test..."
	@(cd ./maelstrom && ./maelstrom test -w broadcast --bin ~/go/bin/maelstrom-playground --node-count 2 --time-limit 10)

