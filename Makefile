# Makefile for a Golang project

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GORUN=$(GOCMD) run


BINARY_NAME=High_Frequency_DNS_Monitoring

build:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o ./bin/$(BINARY_NAME)-linux-amd64 -v

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm ./bin/$(BINARY_NAME)-linux-amd64

run:
	$(GORUN) main.go

# Declare phony targets
.PHONY: build test clean run
