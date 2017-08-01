
LDFLAGS += -X "github.com/dantin/database-tools/utils.Version=0.0.1~git.$(shell git rev-parse --short HEAD)"
LDFLAGS += -X "github.com/dantin/database-tools/utils.BuildTS=$(shell date -u '+%Y-%m-%d %I:%M:%S')"
LDFLAGS += -X "github.com/dantin/database-tools/utils.GitHash=$(shell git rev-parse HEAD)"

CURDIR := $(shell pwd)
GO := go

.PHONY: build importer bulk_ddl

build: importer bulk_ddl

importer:
	$(GO) build -ldflags '$(LDFLAGS)' -o bin/importer cmd/importer/main.go

bulk_ddl:
	$(GO) build -ldflags '$(LDFLAGS)' -o bin/bulk_ddl cmd/bulk_ddl/main.go
