.PHONY: build test test-short lint fmt clean

build:
	go build -o cppenv .

test:
	go test ./... -v

test-short:
	go test ./... -v -short

lint:
	goimports -l .

fmt:
	goimports -w .

clean:
	rm -f cppenv