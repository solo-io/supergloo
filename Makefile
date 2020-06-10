#----------------------------------------------------------------------------------
# Base
#----------------------------------------------------------------------------------

ROOTDIR := $(shell pwd)
OUTPUT_DIR ?= $(ROOTDIR)/_output
SOURCES := $(shell find . -name "*.go" | grep -v test.go)

# Kind of a hack to make sure _output exists
z := $(shell mkdir -p $(OUTPUT_DIR))

COMPONENTS := mesh-discovery mesh-networking
# include helm Makefile so it can be ran from the root
include install/helm/helm.mk

RELEASE := "true"
ifeq ($(TAGGED_VERSION),)
	TAGGED_VERSION := $(shell git describe --tags --dirty --always)
	RELEASE := "false"
endif
VERSION ?= $(shell echo $(TAGGED_VERSION) | cut -c 2-)

LDFLAGS := "-X github.com/solo-io/service-mesh-hub/pkg/common/version.Version=$(VERSION)"
GCFLAGS := all="-N -l"

GO_BUILD_FLAGS := GO111MODULE=on CGO_ENABLED=0 GOARCH=amd64

# If you just put your username, then that refers to your account at hub.docker.com
# To use quay images, set the IMAGE_REPO to "quay.io/solo-io"
# To use dockerhub images, set the IMAGE_REPO to "soloio"
# To use gcr images, set the IMAGE_REPO to "gcr.io"
IMAGE_REPO := soloio


#----------------------------------------------------------------------------------
# Clean
#----------------------------------------------------------------------------------

# Important to clean before pushing new releases. Dockerfiles and binaries may not update properly
.PHONY: clean
clean: helm-clean
	# delete all _output directories, even those in cmd/*
	# -prune prevents searching directories that we just deleted
	find . -name '*_output' -prune -exec rm -rf {} \;

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
	NOT_FORMATTED=$$(gofmt -l ./cmd/ ./ci/) && if [ -n "$$NOT_FORMATTED" ]; then echo These files are not formatted: $$NOT_FORMATTED; exit 1; fi

.PHONY: check-spelling
check-spelling:
	./ci/spell.sh check

mockgen:
	grep -r 'go:generate mockgen' ./ -l | grep -v Makefile | grep -v CONTRIBUTING.md | while read -r f; do echo $$f; go generate $$f; done
	goimports -w .

#----------------------------------------------------------------------------------
# Generated Code and Docs
#----------------------------------------------------------------------------------

.PHONY: generated-code
generated-code:
	rm -rf vendor_any
	CGO_ENABLED=0 go generate -v ./...
	goimports -w .

#----------------------------------------------------------------------------------
# Docker functions
#----------------------------------------------------------------------------------

# $(1) name of container
define build_container
docker build -t ${IMAGE_REPO}/$(1):$(VERSION) $(ROOTDIR)/cmd/$(1)/_output -f $(ROOTDIR)/cmd/$(1)/cmd/Dockerfile;
endef

#----------------------------------------------------------------------------------
# Mesh Discovery
#----------------------------------------------------------------------------------
MESH_DISCOVERY_DIR=cmd/mesh-discovery
MESH_DISCOVERY_OUTPUT_DIR=$(ROOTDIR)/$(MESH_DISCOVERY_DIR)/_output
MESH_DISCOVERY_SOURCES=$(shell find $(MESH_DISCOVERY_DIR) -name "*.go" | grep -v test | grep -v generated.go)

$(MESH_DISCOVERY_OUTPUT_DIR)/mesh-discovery-linux-amd64: $(MESH_DISCOVERY_SOURCES)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) -o $@ $(MESH_DISCOVERY_DIR)/cmd/main.go

.PHONY: mesh-discovery-docker
mesh-discovery-docker: $(MESH_DISCOVERY_OUTPUT_DIR)/mesh-discovery-linux-amd64
	$(call build_container,mesh-discovery)


#----------------------------------------------------------------------------------
# Mesh Networking
#----------------------------------------------------------------------------------
MESH_NETWORKING_DIR=cmd/mesh-networking
MESH_NETWORKING_OUTPUT_DIR=$(ROOTDIR)/$(MESH_NETWORKING_DIR)/_output
MESH_NETWORKING_SOURCES=$(shell find $(MESH_NETWORKING_DIR) -name "*.go" | grep -v test | grep -v generated.go)

$(MESH_NETWORKING_OUTPUT_DIR)/mesh-networking-linux-amd64: $(MESH_NETWORKING_SOURCES)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) -o $@ $(MESH_NETWORKING_DIR)/cmd/main.go

.PHONY: mesh-networking-docker
mesh-networking-docker: $(MESH_NETWORKING_OUTPUT_DIR)/mesh-networking-linux-amd64
	$(call build_container,mesh-networking)

#----------------------------------------------------------------------------------
# Csr Agent
#----------------------------------------------------------------------------------
CSR_AGENT_DIR=services/csr-agent
CSR_AGENT_OUTPUT_DIR=$(ROOTDIR)/$(CSR_AGENT_DIR)/_output
CSR_AGENT_SOURCES=$(shell find $(CSR_AGENT_DIR) -name "*.go" | grep -v test | grep -v generated.go)

$(CSR_AGENT_OUTPUT_DIR)/csr-agent-linux-amd64: $(CSR_AGENT_SOURCES)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) -o $@ $(CSR_AGENT_DIR)/cmd/main.go

.PHONY: csr-agent-docker
csr-agent-docker: $(CSR_AGENT_OUTPUT_DIR)/csr-agent-linux-amd64
	$(call build_container,csr-agent)

#----------------------------------------------------------------------------------
# meshctl
#----------------------------------------------------------------------------------
CLI_DIR=cli

$(OUTPUT_DIR)/meshctl: $(SOURCES)
	$(GO_BUILD_FLAGS) go build -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) -o $@ $(CLI_DIR)/cmd/main.go

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
docker: mesh-discovery-docker mesh-networking-docker csr-agent-docker

# $(1) name of component
define docker_push
docker push ${IMAGE_REPO}/$(1):$(VERSION);
endef

# Depends on DOCKER_IMAGES, which is set to docker if RELEASE is "true", otherwise empty (making this a no-op).
# This prevents executing the dependent targets if RELEASE is not true, while still enabling `make docker`
# to be used for local testing.
# docker-push is intended to be run by CI
docker-push: docker
	$(foreach component,$(COMPONENTS),$(call docker_push,$(component)))
	$(call docker_push,csr-agent)

CLUSTER_NAME := $(if $(CLUSTER_NAME),$(CLUSTER_NAME), $(shell kind get clusters | grep management-plane))
define kind_load
kind load docker-image ${IMAGE_REPO}/$(1):$(VERSION) --name $(CLUSTER_NAME);
endef

.PHONY: kind-load-images
kind-load-images:
	$(foreach component,$(COMPONENTS),$(call kind_load,$(component)))
	$(call kind_load,csr-agent)

.PHONY: start-local-env
start-local-env:
	./ci/setup-kind.sh

.PHONY: destroy-local-env
destroy-local-env:
	./ci/setup-kind.sh cleanup
