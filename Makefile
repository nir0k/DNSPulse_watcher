# Makefile for a Golang project

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GORUN=$(GOCMD) run

BINARY_NAME=HighFrequencyDNSChecker

build:
	@CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=x86_64-linux-gnu-gcc $(GOBUILD) -o ./bin/$(BINARY_NAME)-linux-amd64 ./cmd/$(BINARY_NAME)

test:
	@$(GOTEST) -v ./...

clean:
	@$(GOCLEAN)
	@rm ./bin/$(BINARY_NAME)-linux-amd64

run:
	@env GOOS=darwin $(GOBUILD) -o bin/$(BINARY_NAME) ./cmd/$(BINARY_NAME)

# Declare phony targets
.PHONY: build test clean run
