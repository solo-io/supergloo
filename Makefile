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
	TAGGED_VERSION := $(shell git describe --tags --dirty)
	RELEASE := "false"
endif
VERSION ?= $(shell echo $(TAGGED_VERSION) | cut -c 2-)
# Kind of a hack to make sure _output exists
z := $(shell mkdir -p $(OUTPUT_DIR))

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

SUBDIRS:=services ci pkg cli test

.PHONY: generated-code
generated-code:
	CGO_ENABLED=0 go generate ./...
	goimports -w $(SUBDIRS)

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
# Deployment Manifests / Helm
#----------------------------------------------------------------------------------

HELM_SYNC_DIR := $(OUTPUT_DIR)/helm
HELM_DIR := install/helm
INSTALL_NAMESPACE ?= sm-marketplace

.PHONY: manifest
manifest: prepare-helm update-helm-chart install/mesh-projects.yaml

# creates Chart.yaml, values.yaml See install/helm/mesh-projects/README.md for more info.
.PHONY: prepare-helm
prepare-helm: $(OUTPUT_DIR)/.helm-prepared

$(OUTPUT_DIR)/.helm-prepared:
	go run install/helm/mesh-projects/generate.go $(VERSION)
	mkdir -p $(OUTPUT_DIR)/helm
	touch $@

update-helm-chart:
	mkdir -p $(HELM_SYNC_DIR)/charts
	helm package --destination $(HELM_SYNC_DIR)/charts $(HELM_DIR)/mesh-projects
	helm repo index $(HELM_SYNC_DIR)

HELMFLAGS ?= --namespace $(INSTALL_NAMESPACE) --set namespace.create=true

MANIFEST_OUTPUT = > /dev/null
ifneq ($(BUILD_ID),)
MANIFEST_OUTPUT =
endif

install/mesh-projects.yaml: prepare-helm
	helm template install/helm/mesh-projects $(HELMFLAGS) | tee $@ $(OUTPUT_YAML) $(MANIFEST_OUTPUT)

.PHONY: render-yaml
render-yaml: install/mesh-projects.yaml

.PHONY: save-helm
save-helm:
ifeq ($(RELEASE),"true")
	gsutil -m rsync -r './_output/helm' gs://mesh-projects-helm/
endif

.PHONY: fetch-helm
fetch-helm:
	mkdir -p $(OUTPUT_DIR)/helm
	gsutil -m rsync -r gs://mesh-projects-helm/ './_output/helm'

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
# Base Image
# Note: this is managed by a manual process
# To update the base image, run this step with the appropriate version, and update
# the version referenced in the code (pkg/version/version.go)
# cmd: TAGGED_VERSION=vX.Y.Z make push-base-image
#----------------------------------------------------------------------------------
.PHONY: push-base-image
push-base-image: dockerfile-generation
	cd build/base_image && docker build -t quay.io/solo-io/mc-base-image:$(VERSION) .
	docker push quay.io/solo-io/mc-base-image:$(VERSION)

# Note that this script can generate other resources too.
# At the moment, only Dockerfiles are managed by it.
.PHONY: dockerfile-generation
dockerfile-generation:
	go run hack/genbuild/main.go

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
docker: mesh-discovery-docker

# Depends on DOCKER_IMAGES, which is set to docker if RELEASE is "true", otherwise empty (making this a no-op).
# This prevents executing the dependent targets if RELEASE is not true, while still enabling `make docker`
# to be used for local testing.
# docker-push is intended to be run by CI
docker-push: $(DOCKER_IMAGES)
ifeq ($(RELEASE),"true")
	docker push quay.io/solo-io/mc-mesh-discovery:$(VERSION)
endif
