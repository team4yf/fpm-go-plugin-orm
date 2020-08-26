PROJECTNAME=$(shell basename "$(PWD)")
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin

all: install build docker-build docker-push

install:
	go mod download

dev:
	go build -race -o $(GOBIN)/app ./main.go && $(GOBIN)/app
