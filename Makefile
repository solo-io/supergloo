#----------------------------------------------------------------------------------
# Base
#----------------------------------------------------------------------------------

ROOTDIR := $(shell pwd)
OUTPUT_DIR ?= $(ROOTDIR)/_output
SOURCES := $(shell find . -name "*.go" | grep -v test.go)
#VERSION ?= $(shell git describe --tags)
# TODO: use the above instead
VERSION ?= dev

#----------------------------------------------------------------------------------
# Repo init
#----------------------------------------------------------------------------------
# https://www.viget.com/articles/two-ways-to-share-git-hooks-with-your-team/
.PHONY: init
init:
	git config core.hooksPath .githooks

#----------------------------------------------------------------------------------
# Generated Code
#----------------------------------------------------------------------------------


.PHONY: generated-code
generated-code:
	go generate ./... && gofmt -w pkg && goimports -w pkg

#################
#################
#               #
#     Build     #
#               #
#               #
#################
#################
#################


#----------------------------------------------------------------------------------
# SuperGloo
#----------------------------------------------------------------------------------

SOURCES=$(shell find . -name "*.go" | grep -v test )

$(OUTPUT_DIR)/supergloo-linux-amd64: $(SOURCES)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o $@ cmd/main.go


.PHONY: supergloo
supergloo: $(OUTPUT_DIR)/supergloo-linux-amd64

$(OUTPUT_DIR)/Dockerfile.supergloo: cmd/Dockerfile
	cp $< $@

supergloo-docker: $(OUTPUT_DIR)/supergloo-linux-amd64 $(OUTPUT_DIR)/Dockerfile.supergloo
	docker build -t soloio/supergloo:$(VERSION)  $(OUTPUT_DIR) -f $(OUTPUT_DIR)/Dockerfile.supergloo

supergloo-docker-push: supergloo-docker
	docker push soloio/supergloo:$(VERSION)

#----------------------------------------------------------------------------------
# SuperGloo Server (for local testing)
#----------------------------------------------------------------------------------

.PHONY: supergloo-server
supergloo-server:
	cd cmd && go build -o $(GOPATH)/bin/supergloo-server

#----------------------------------------------------------------------------------
# SuperGloo CLI
#----------------------------------------------------------------------------------

.PHONY: install-cli
install-cli:
	cd cli/cmd && go build -o $(GOPATH)/bin/supergloo