PKG := ./cmd/worker
BIN := worker

.PHONY: all build install test vet tidy clean

all: build

build:
	go build -o $(BIN) $(PKG)

install:
	go install $(PKG)

test:
	go test ./...

vet:
	go vet ./...

tidy:
	go mod tidy

clean:
	rm -f $(BIN)
