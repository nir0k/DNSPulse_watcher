# Makefile for a Golang project

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GORUN=$(GOCMD) run

BINARY_NAME=HighFrequencyDNSChecker

build:
	GOOS=linux GOARCH=amd64 CC=x86_64-linux-gnu-gcc $(GOBUILD) -o ./bin/$(BINARY_NAME)-linux-amd64 .

test:
	@$(GOTEST) -v ./...

clean:
	@$(GOCLEAN)
	@rm ./bin/$(BINARY_NAME)-linux-amd64

run:
	@env GOOS=darwin $(GOBUILD) -o bin/$(BINARY_NAME) .

# Declare phony targets
.PHONY: build test clean run
