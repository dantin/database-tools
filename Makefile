
LDFLAGS += -X "github.com/dantin/database-tools/utils.Version=0.0.1~git.$(shell git rev-parse --short HEAD)"
LDFLAGS += -X "github.com/dantin/database-tools/utils.BuildTS=$(shell date -u '+%Y-%m-%d %I:%M:%S')"
LDFLAGS += -X "github.com/dantin/database-tools/utils.GitHash=$(shell git rev-parse HEAD)"

CURDIR := $(shell pwd)
GO := go

.PHONY: build importer bulk-ddl db-config

build: importer bulk-ddl db-config

importer:
	$(GO) build -ldflags '$(LDFLAGS)' -o bin/importer cmd/importer/main.go

bulk-ddl:
	$(GO) build -ldflags '$(LDFLAGS)' -o bin/bulk-ddl cmd/bulk-ddl/main.go

db-config:
	$(GO) build -ldflags '$(LDFLAGS)' -o bin/db-config cmd/db-config/main.go
