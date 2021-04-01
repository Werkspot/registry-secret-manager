update:
	go get -u && go mod tidy

test:
	go test -race ./...

lint:
	golangci-lint run

go-mod:
	go mod tidy
	go mod verify

verify: go-mod test lint