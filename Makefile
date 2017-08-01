
CURDIR := $(shell pwd)
GO := go

.PHONY: build

build: 
	$(GO) build -o bin/importer cmd/importer/main.go
