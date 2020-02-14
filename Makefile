#----------------------------------------------------------------------------------
# Base
#----------------------------------------------------------------------------------

ROOTDIR := $(shell pwd)
OUTPUT_DIR ?= $(ROOTDIR)/_output
SOURCES := $(shell find . -name "*.go" | grep -v test.go)
RELEASE := "true"
ifeq ($(TAGGED_VERSION),)
	TAGGED_VERSION := $(shell git describe --tags --dirty)
	RELEASE := "false"
endif
VERSION ?= $(shell echo $(TAGGED_VERSION) | cut -c 2-)
# Kind of a hack to make sure _output exists
z := $(shell mkdir -p $(OUTPUT_DIR))

LDFLAGS := "-X github.com/solo-io/mesh-projects/pkg/version.Version=$(VERSION)"
GCFLAGS := all="-N -l"

COMPONENTS := mesh-discovery mesh-group

# include helm makefile so it can be ran from the root
include install/helm/helm.mk

#----------------------------------------------------------------------------------
# Clean
#----------------------------------------------------------------------------------

# Important to clean before pushing new releases. Dockerfiles and binaries may not update properly
.PHONY: clean
clean: helm-clean

#----------------------------------------------------------------------------------
# Repo setup
#----------------------------------------------------------------------------------

# https://www.viget.com/articles/two-ways-to-share-git-hooks-with-your-team/
.PHONY: init
init: update-deps
	git config core.hooksPath .githooks

.PHONY: mod-download
mod-download:
	go mod download


.PHONY: update-deps
update-deps: mod-download
	cd $(shell go list -f '{{ .Dir }}' -m istio.io/tools) && \
	  go install ./cmd/protoc-gen-jsonshim
	GO111MODULE=off go get -u golang.org/x/tools/cmd/goimports
	GO111MODULE=off go get -u github.com/gogo/protobuf/gogoproto
	GO111MODULE=off go get -u github.com/gogo/protobuf/protoc-gen-gogo
	GO111MODULE=off go get -u github.com/solo-io/protoc-gen-ext
	GO111MODULE=off go get -u github.com/google/wire/cmd/wire
	GO111MODULE=off go get -u github.com/golang/mock/gomock
	GO111MODULE=off go install github.com/golang/mock/mockgen


.PHONY: fmt-changed
fmt-changed:
	git diff --name-only | grep '.*.go$$' | xargs goimports -w
	git diff --cached --name-only | grep '.*.go$$' | xargs goimports -w


# Enumerate the directories to validate with yamllint. If we just run "yamllint ." the command fails on a symlink in
# the vendor directory, even though we have excluded the directory in the .yamllint config file
.PHONY: check-format
check-format:
	NOT_FORMATTED=$$(gofmt -l ./services/ ./ci/) && if [ -n "$$NOT_FORMATTED" ]; then echo These files are not formatted: $$NOT_FORMATTED; exit 1; fi

.PHONY: check-spelling
check-spelling:
	./ci/spell.sh check

#----------------------------------------------------------------------------------
# Generated Code and Docs
#----------------------------------------------------------------------------------

.PHONY: generated-code
generated-code:
	rm -rf vendor_any
	CGO_ENABLED=0 go generate ./...
	goimports -w .

#----------------------------------------------------------------------------------
# Docker functions
#----------------------------------------------------------------------------------

# $(1) name of container
define build_container
docker build -t quay.io/solo-io/$(1):$(VERSION) $(ROOTDIR)/services/$(1)/_output -f $(ROOTDIR)/services/$(1)/cmd/Dockerfile;
endef

#----------------------------------------------------------------------------------
# Mesh Discovery
#----------------------------------------------------------------------------------
MESH_DISCOVERY=mesh-discovery
MESH_DISCOVERY_DIR=services/$(MESH_DISCOVERY)
MESH_DISCOVERY_OUTPUT_DIR=$(ROOTDIR)/$(MESH_DISCOVERY_DIR)/_output
MESH_DISCOVERY_SOURCES=$(shell find $(MESH_DISCOVERY_DIR) -name "*.go" | grep -v test | grep -v generated.go)

$(MESH_DISCOVERY_OUTPUT_DIR)/mesh-discovery-linux-amd64: $(MESH_DISCOVERY_SOURCES)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) -o $@ $(MESH_DISCOVERY_DIR)/cmd/main.go

.PHONY: mesh-discovery-docker
mesh-discovery-docker: $(MESH_DISCOVERY_OUTPUT_DIR)/mesh-discovery-linux-amd64
	$(call build_container,$(MESH_DISCOVERY))


#----------------------------------------------------------------------------------
# Mesh Group
#----------------------------------------------------------------------------------
MESH_GROUP=mesh-group
MESH_GROUP_DIR=services/$(MESH_GROUP)
MESH_GROUP_OUTPUT_DIR=$(ROOTDIR)/$(MESH_GROUP_DIR)/_output
MESH_GROUP_SOURCES=$(shell find $(MESH_GROUP_DIR) -name "*.go" | grep -v test | grep -v generated.go)

$(MESH_GROUP_OUTPUT_DIR)/mesh-group-linux-amd64: $(MESH_GROUP_SOURCES)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) -o $@ $(MESH_GROUP_DIR)/cmd/main.go

.PHONY: mesh-group-docker
mesh-group-docker: $(MESH_GROUP_OUTPUT_DIR)/mesh-group-linux-amd64
	$(call build_container,$(MESH_GROUP))

#----------------------------------------------------------------------------------
# meshctl
#----------------------------------------------------------------------------------
CLI_DIR=cli

$(OUTPUT_DIR)/meshctl: $(SOURCES)
	GO111MODULE=on go build -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) -o $@ $(CLI_DIR)/cmd/main.go

$(OUTPUT_DIR)/meshctl-linux-amd64: $(SOURCES)
	$(GO_BUILD_FLAGS) GOOS=linux go build -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) -o $@ $(CLI_DIR)/cmd/main.go

$(OUTPUT_DIR)/meshctl-darwin-amd64: $(SOURCES)
	$(GO_BUILD_FLAGS) GOOS=darwin go build -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) -o $@ $(CLI_DIR)/cmd/main.go

$(OUTPUT_DIR)/meshctl-windows-amd64.exe: $(SOURCES)
	$(GO_BUILD_FLAGS) GOOS=windows go build -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) -o $@ $(CLI_DIR)/cmd/main.go


.PHONY: meshctl
meshctl: $(OUTPUT_DIR)/meshctl
.PHONY: meshctl-linux-amd64
meshctl-linux-amd64: $(OUTPUT_DIR)/meshctl-linux-amd64
.PHONY: meshctl-darwin-amd64
meshctl-darwin-amd64: $(OUTPUT_DIR)/meshctl-darwin-amd64
.PHONY: meshctl-windows-amd64
meshctl-windows-amd64: $(OUTPUT_DIR)/meshctl-windows-amd64.exe

.PHONY: build-cli
build-cli: meshctl-linux-amd64 meshctl-darwin-amd64 meshctl-windows-amd64

#----------------------------------------------------------------------------------
# Release
#----------------------------------------------------------------------------------

.PHONY: release
release: docker-push upload-github-release-assets

# The code does the proper checking for a TAGGED_VERSION
.PHONY: upload-github-release-assets
upload-github-release-assets: build-cli
	go run ci/upload_github_release_assets.go

#----------------------------------------------------------------------------------
# Docker
#----------------------------------------------------------------------------------
#
#---------
#--------- Push
#---------

.PHONY: docker docker-push
docker: mesh-discovery-docker mesh-group-docker

# $(1) name of component
define docker_push
docker push quay.io/solo-io/$(1):$(VERSION);
endef

# Depends on DOCKER_IMAGES, which is set to docker if RELEASE is "true", otherwise empty (making this a no-op).
# This prevents executing the dependent targets if RELEASE is not true, while still enabling `make docker`
# to be used for local testing.
# docker-push is intended to be run by CI
docker-push: docker
	$(foreach component,$(COMPONENTS),$(call docker_push,$(component)))

