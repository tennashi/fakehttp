GOBIN ?= $(shell go env GOPATH)/bin

.PHONY: lint
lint: $(GOBIN)/golint
	go vet ./...
	golint -set_exit_status ./...

$(GOBIN)/golint:
	cd && go get golang.org/x/lint/golint

.PHONY: test
test:
	go test -v ./...

