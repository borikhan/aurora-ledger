.PHONY: run test verify container
run:
	go run ./cmd/api
test:
	go test -race ./...
verify:
	test -z "$$(gofmt -l .)"
	go test -race ./...
container:
	docker build -t aurora-ledger .
