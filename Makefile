PHONY: tools
tools:
	go install github.com/daixiang0/gci@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin

PHONY: tidy
tidy:
	@go mod tidy -v

PHONY: install
install:
	@go mod download

PHONY: format
format:
	@golangci-lint run --fix ./...

PHONY: lint
lint:
	golangci-lint run -v ./...

PHONY: test
test:
	go test -race -v -shuffle on ./...

test/%:
	go vet ./$(@:test/%=%)
	go test -race -v -shuffle on ./$(@:test/%=%)
