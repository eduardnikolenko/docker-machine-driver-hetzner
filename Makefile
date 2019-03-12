NAME    := docker-machine-driver-hetzner
VERSION ?= $(shell git describe --tags --abbrev=0)
LDFLAGS := -X main.Version=$(VERSION)

all: test

build: deps lint vet test
	@echo "+ $@"
	@go build -ldflags "$(LDFLAGS)" -o "$(NAME)" main.go

deps:
	@echo "+ $@"
	@dep ensure

lint:
	@echo "+ $@"
	@golint $(shell go list ./... 2>/dev/null | grep -v /vendor/)

vet:
	@echo "+ $@"
	@go vet $(shell go list ./... 2>/dev/null | grep -v /vendor/)

test:
	@echo "+ $@"
	@go test $(shell go list -v ./... 2>/dev/null | grep -v /vendor/)

clean:
	@echo "+ $@"
	@$(RM) -f "$(NAME)"

.PHONY: all build deps lint vet test clean
