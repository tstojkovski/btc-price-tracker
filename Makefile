# Makefile for Bitcoin Price Tracker

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet
GOLINT=golangci-lint

# Binary name
BINARY_NAME=btc-price-tracker
SERVER_MAIN=./cmd/server/main.go

# Build directory
BUILD_DIR=build

# Docker parameters
DOCKER_IMAGE=btc-price-tracker
DOCKER_TAG=latest

# Coverage output
COVERAGE_OUT=coverage.out
COVERAGE_HTML=coverage.html

.PHONY: all build clean run test test-verbose test-coverage fmt lint vet docker-build docker-run help tidy

all: test build

# Build the application
build:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(SERVER_MAIN)

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f $(COVERAGE_OUT) $(COVERAGE_HTML)

# Run the application
run:
	$(GORUN) $(SERVER_MAIN)

# Run all tests
test:
	$(GOTEST) -v ./...

# Run tests with verbose output
test-verbose:
	$(GOTEST) -v -count=1 ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -count=1 -coverprofile=$(COVERAGE_OUT) ./...
	$(GOCMD) tool cover -html=$(COVERAGE_OUT) -o $(COVERAGE_HTML)
	open $(COVERAGE_HTML)

# Format code
fmt:
	$(GOFMT) ./...

# Run go vet
vet:
	$(GOVET) ./...

# Run linter (requires golangci-lint to be installed)
lint:
	$(GOLINT) run

# Build docker image
docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

# Run the application in docker
docker-run:
	docker run -p 8082:8082 $(DOCKER_IMAGE):$(DOCKER_TAG)

# Update go.mod and go.sum
tidy:
	$(GOMOD) tidy

# Display help
help:
	@echo "Bitcoin Price Tracker Makefile"
	@echo "Usage:"
	@echo "  make build        - Build the application"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make run          - Run the application"
	@echo "  make test         - Run all tests"
	@echo "  make test-verbose - Run tests with verbose output"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo "  make fmt          - Format code"
	@echo "  make vet          - Run go vet"
	@echo "  make lint         - Run linter"
	@echo "  make docker-build - Build docker image"
	@echo "  make docker-run   - Run in docker"
	@echo "  make tidy         - Update dependencies"