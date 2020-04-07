GOBIN ?= $(shell go env GOPATH)/bin

.PHONY: lint
lint: $(GOBIN)/golint
	go vet ./...
	golint -set_exit_status ./...

.PHONY: test
test:
	go test -v ./...

