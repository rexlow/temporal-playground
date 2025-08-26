.PHONY: build clean run-worker run-client test

# Build the application
build:
	go build -o temporal-playground .

# Clean build artifacts
clean:
	rm -f temporal-playground

# Run the worker
run-worker: build
	./temporal-playground worker -n local

# Run a sample workflow
run-client: build
	./temporal-playground client start -n local

# Test the build
test:
	go test ./...

# Default target
all: clean build
