#----------------------------------------------------------------------------------
# Base
#----------------------------------------------------------------------------------

ROOTDIR := $(shell pwd)
OUTPUT_DIR ?= $(ROOTDIR)/_output
SOURCES := $(shell find . -name "*.go" | grep -v test.go | grep -v '\.\#*')
RELEASE := "true"
ifeq ($(TAGGED_VERSION),)
	# TAGGED_VERSION := $(shell git describe --tags)
	# This doesn't work in CI, need to find another way...
	TAGGED_VERSION := vdev
	RELEASE := "false"
endif
VERSION ?= $(shell echo $(TAGGED_VERSION) | cut -c 2-)

LDFLAGS := "-X github.com/solo-io/mesh-projects/pkg/version.Version=$(VERSION)"
GCFLAGS := all="-N -l"


#----------------------------------------------------------------------------------
# Clean
#----------------------------------------------------------------------------------

# Important to clean before pushing new releases. Dockerfiles and binaries may not update properly
.PHONY: clean
clean:
	rm -rf _output
	git clean -xdf install ui/src/proto/github.com

#----------------------------------------------------------------------------------
# Repo setup
#----------------------------------------------------------------------------------

# https://www.viget.com/articles/two-ways-to-share-git-hooks-with-your-team/
.PHONY: init
init:
	git config core.hooksPath .githooks

.PHONY: update-deps
update-deps:
	go get -u golang.org/x/tools/cmd/goimports
	go get -u github.com/gogo/protobuf/gogoproto
	go get -u github.com/gogo/protobuf/protoc-gen-gogo
	go get -u github.com/paulvollmer/2gobytes
	go get -u github.com/google/wire/cmd/wire
	go get -u github.com/golang/mock/gomock
	go install github.com/golang/mock/mockgen


.PHONY: pin-repos
pin-repos:
	go run ci/pin_repos.go


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

SUBDIRS:=services ci

.PHONY: generated-code
generated-code:
	CGO_ENABLED=0 go generate ./...
	gofmt -w $(SUBDIRS)
	goimports -w $(SUBDIRS)
	mkdir -p $(OUTPUT_DIR) # TODO: find a better place for this

#----------------------------------------------------------------------------------
# Apiserver
#----------------------------------------------------------------------------------

MESH_DISCOVERY_DIR=services/mesh-discovery
MESH_DISCOVERY_SOURCES=$(shell find $(MESH_DISCOVERY_DIR) -name "*.go" | grep -v test | grep -v generated.go)

$(OUTPUT_DIR)/mesh-discovery-linux-amd64: $(MESH_DISCOVERY_SOURCES)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) -o $@ $(MESH_DISCOVERY_DIR)/cmd/main.go

.PHONY: apiserver
apiserver: $(OUTPUT_DIR)/mesh-discovery-linux-amd64

$(OUTPUT_DIR)/Dockerfile.mesh-discovery: $(MESH_DISCOVERY_DIR)/cmd/Dockerfile
	cp $< $@

.PHONY: mesh-discovery-docker
mesh-discovery-docker: $(OUTPUT_DIR)/.mesh-discovery-docker

$(OUTPUT_DIR)/.mesh-discovery-docker: $(OUTPUT_DIR)/mesh-discovery-linux-amd64 $(OUTPUT_DIR)/Dockerfile.mesh-discovery
	docker build -t quay.io/solo-io/mc-mesh-discovery:$(VERSION) $(call get_test_tag_option,mesh-discovery) $(OUTPUT_DIR) -f $(OUTPUT_DIR)/Dockerfile.mesh-discovery
	touch $@

#----------------------------------------------------------------------------------
# Mesh Bridge
#----------------------------------------------------------------------------------

MESH_BRIDGE_DIR=services/mesh-bridge
MESH_BRIDGE_SOURCES=$(shell find $(MESH_BRIDGE_DIR) -name "*.go" | grep -v test | grep -v generated.go)

$(OUTPUT_DIR)/mesh-bridge-linux-amd64: $(MESH_BRIDGE_SOURCES)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) -o $@ $(MESH_BRIDGE_DIR)/cmd/main.go

.PHONY: mesh-bridge
mesh-bridge: $(OUTPUT_DIR)/mesh-bridge-linux-amd64

$(OUTPUT_DIR)/Dockerfile.mesh-bridge: $(MESH_BRIDGE_DIR)/cmd/Dockerfile
	cp $< $@

.PHONY: mesh-bridge-docker
mesh-bridge-docker: $(OUTPUT_DIR)/.mesh-bridge-docker

$(OUTPUT_DIR)/.mesh-bridge-docker: $(OUTPUT_DIR)/mesh-bridge-linux-amd64 $(OUTPUT_DIR)/Dockerfile.mesh-bridge
	docker build -t quay.io/solo-io/mc-mesh-bridge:$(VERSION) $(call get_test_tag_option,mesh-bridge) $(OUTPUT_DIR) -f $(OUTPUT_DIR)/Dockerfile.mesh-bridge
	touch $@
#----------------------------------------------------------------------------------
# Release
#----------------------------------------------------------------------------------

.PHONY: upload-github-release-assets
upload-github-release-assets:
	go run ci/upload_github_release_assets.go

.PHONY: release
release: docker-push upload-github-release-assets

#----------------------------------------------------------------------------------
# Docker
#----------------------------------------------------------------------------------
#
#---------
#--------- Push
#---------

DOCKER_IMAGES :=
ifeq ($(RELEASE),"true")
	DOCKER_IMAGES := docker
endif

.PHONY: docker docker-push
docker: mesh-discovery-docker mesh-bridge-docker

# Depends on DOCKER_IMAGES, which is set to docker if RELEASE is "true", otherwise empty (making this a no-op).
# This prevents executing the dependent targets if RELEASE is not true, while still enabling `make docker`
# to be used for local testing.
# docker-push is intended to be run by CI
docker-push: $(DOCKER_IMAGES)
ifeq ($(RELEASE),"true")
	docker push quay.io/solo-io/mc-mesh-bridge:$(VERSION) && \
	docker push quay.io/solo-io/mc-mesh-discovery:$(VERSION)
endif
