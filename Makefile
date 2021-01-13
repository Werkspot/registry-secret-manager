test:
	go test -race ./...

lint:
	fgt goimports -w .
	fgt golint ./...
	fgt go vet ./...
	fgt go fmt ./...
	fgt errcheck -ignore Close  ./...

install-tools:
	GO111MODULE=off go get -u golang.org/x/lint/golint
	GO111MODULE=off go get -u golang.org/x/tools/cmd/goimports
	GO111MODULE=off go get -u github.com/GeertJohan/fgt
	GO111MODULE=off go get -u github.com/kisielk/errcheck

go-mod:
	go mod tidy
	go mod verify

verify: go-mod test install-tools lint
