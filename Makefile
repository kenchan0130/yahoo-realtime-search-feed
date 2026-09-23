PHONY: tools
tools:
	go install github.com/daixiang0/gci@latest
	@set -eu; archive=$$(mktemp); \
	trap 'rm -f "$$archive"' EXIT; \
	curl -fsSL -o "$$archive" https://github.com/golangci/golangci-lint/releases/download/v2.11.3/golangci-lint-2.11.3-linux-amd64.tar.gz; \
	echo "87bb8cddbcc825d5778b64e8a91b46c0526b247f4e2f2904dea74ec7450475d1  $$archive" | shasum -a 256 -c -; \
	mkdir -p "$$(go env GOPATH)/bin"; \
	tar -xz -C "$$(go env GOPATH)/bin" --strip-components=1 -f "$$archive" golangci-lint-2.11.3-linux-amd64/golangci-lint

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
