
.PHONY: proto
proto:
	buf dep update
	buf generate

.PHONY: run
run: proto
	go run main.go