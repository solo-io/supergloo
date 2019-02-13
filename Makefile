#----------------------------------------------------------------------------------
# Base
#----------------------------------------------------------------------------------


ROOTDIR := $(shell pwd)
OUTPUT_DIR ?= $(ROOTDIR)/_output
SOURCES := $(shell find . -name "*.go" | grep -v test.go)
VERSION ?= $(shell git describe --tags )
REPOSITORY ?= $(basename `git rev-parse --show-toplevel`)
LDFLAGS := "-X github.com/solo-io/supergloo/pkg2/version.Version=$(VERSION)"

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


fmt:
	gofmt -w pkg test && goimports -w pkg test

.PHONY: generated-code
generated-code:
	go generate ./...
	gofmt -w pkg
	goimports -w pkg

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

SOURCES=$(shell find . -name "*.go" | grep -v test | grep -v mock)

$(OUTPUT_DIR)/supergloo-linux-amd64: $(SOURCES)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags=$(LDFLAGS) -o $@ cmd/main.go
	shasum -a 256 $@ > $@.sha256

$(OUTPUT_DIR)/supergloo-darwin-amd64: $(SOURCES)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=darwin go build -ldflags=$(LDFLAGS) -o $@ cmd/main.go
	shasum -a 256 $@ > $@.sha256

supergloo: $(SOURCES)
	go build -ldflags=$(LDFLAGS) -o $(OUTPUT_DIR)/$@ cmd/main.go

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
	cd cmd && go build -ldflags=$(LDFLAGS) -o $(GOPATH)/bin/supergloo-server

#----------------------------------------------------------------------------------
# SuperGloo CLI
#----------------------------------------------------------------------------------

SOURCES=$(shell find . -name "*.go" | grep -v test | grep -v mock)

.PHONY: install-cli
install-cli:
	cd cli/cmd && go build -ldflags=$(LDFLAGS) -o $(GOPATH)/bin/supergloo


$(OUTPUT_DIR)/supergloo-cli-linux-amd64: $(SOURCES)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags=$(LDFLAGS) -o $@ cli/cmd/main.go
	shasum -a 256 $@ > $@.sha256

$(OUTPUT_DIR)/supergloo-cli-darwin-amd64: $(SOURCES)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=darwin go build -ldflags=$(LDFLAGS) -o $@ cli/cmd/main.go
	shasum -a 256 $@ > $@.sha256


#----------------------------------------------------------------------------------
# Release
#----------------------------------------------------------------------------------

RELEASE_BINARIES := \
	$(OUTPUT_DIR)/supergloo-linux-amd64 \
	$(OUTPUT_DIR)/supergloo-darwin-amd64 \
	$(OUTPUT_DIR)/supergloo-cli-linux-amd64 \
	$(OUTPUT_DIR)/supergloo-cli-darwin-amd64 \

.PHONY: release-binaries
release-binaries: $(RELEASE_BINARIES)

.PHONY: release-checksum
release-checksum:

.PHONY: release
release: release-binaries
	hack/create-release.sh github_api_token=$(GITHUB_TOKEN) owner=solo-io repo=$(REPOSITORY) tag=$(VERSION)
	@$(foreach BINARY,$(RELEASE_BINARIES),hack/upload-github-release-asset.sh github_api_token=$(GITHUB_TOKEN) owner=solo-io repo=$(REPOSITORY) tag=$(VERSION) filename=$(BINARY) && hack/upload-github-release-asset.sh github_api_token=$(GITHUB_TOKEN) owner=solo-io repo=$(REPOSITORY) tag=$(VERSION) filename=$(BINARY).sha256;)
