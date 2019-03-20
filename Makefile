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

LDFLAGS := "-X github.com/solo-io/supergloo/pkg/version.Version=$(VERSION)"
GCFLAGS := all="-N -l"

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
	go get -u github.com/lyft/protoc-gen-validate
	go get -u github.com/paulvollmer/2gobytes
	go get -u github.com/golang/mock/gomock
    go install github.com/golang/mock/mockgen

.PHONY: pin-repos
pin-repos:
	go run ci/pin_repos.go

.PHONY: check-format
check-format:
	NOT_FORMATTED=$$(gofmt -l ./pkg/ ./test/) && if [ -n "$$NOT_FORMATTED" ]; then echo These files are not formatted: $$NOT_FORMATTED; exit 1; fi

check-spelling:
	./ci/spell.sh check

.PHONY: generated-code
generated-code: $(OUTPUT_DIR)/.generated-code

SUBDIRS:=cli pkg cmd test
$(OUTPUT_DIR)/.generated-code:
	go generate ./...
	(rm -f docs/cli/supergloo*; go run cli/cmd/docs/main.go)
	gofmt -w $(SUBDIRS)
	goimports -w $(SUBDIRS)
	mkdir -p $(OUTPUT_DIR)
	touch $@

#----------------------------------------------------------------------------------
# Clean
#----------------------------------------------------------------------------------

# Important to clean before pushing new releases. Dockerfiles and binaries may not update properly
.PHONY: clean
clean:
	rm -rf _output
	rm -fr site


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

### Controller

$(OUTPUT_DIR)/supergloo-linux-amd64: $(SOURCES)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags=$(LDFLAGS) -o $@ cmd/main.go
	shasum -a 256 $@ > $@.sha256

$(OUTPUT_DIR)/supergloo-darwin-amd64: $(SOURCES)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=darwin go build -ldflags=$(LDFLAGS) -o $@ cmd/main.go
	shasum -a 256 $@ > $@.sha256

$(OUTPUT_DIR)/Dockerfile.supergloo: cmd/Dockerfile
	cp $< $@

supergloo-docker: $(OUTPUT_DIR)/supergloo-linux-amd64 $(OUTPUT_DIR)/Dockerfile.supergloo
	docker build -t quay.io/solo-io/supergloo:$(VERSION)  $(OUTPUT_DIR) -f $(OUTPUT_DIR)/Dockerfile.supergloo

supergloo-docker-push: supergloo-docker
	docker push quay.io/solo-io/supergloo:$(VERSION)

#----------------------------------------------------------------------------------
# SuperGloo CLI
#----------------------------------------------------------------------------------

SOURCES=$(shell find . -name "*.go" | grep -v test | grep -v mock)

.PHONY: install-cli
install-cli:
	cd cli/cmd && go build -ldflags=$(LDFLAGS) -o $(GOPATH)/bin/supergloo

$(OUTPUT_DIR)/supergloo-cli-linux-amd64: $(SOURCES)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags=$(LDFLAGS) -o $@ cli/cmd/main.go

$(OUTPUT_DIR)/supergloo-cli-darwin-amd64: $(SOURCES)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=darwin go build -ldflags=$(LDFLAGS) -o $@ cli/cmd/main.go

$(OUTPUT_DIR)/supergloo-cli-windows-amd64.exe: $(SOURCES)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=windows go build -ldflags=$(LDFLAGS) -o $@ cli/cmd/main.go

.PHONY: build-cli
ifeq ($(RELEASE),"true")
build-cli: $(OUTPUT_DIR)/supergloo-cli-linux-amd64 $(OUTPUT_DIR)/supergloo-cli-darwin-amd64 $(OUTPUT_DIR)/supergloo-cli-windows-amd64.exe
else
build-cli:
	go build -ldflags=$(LDFLAGS) -o supergloo cli/cmd/main.go
endif

#----------------------------------------------------------------------------------
# Deployment Manifests / Helm
#----------------------------------------------------------------------------------


HELM_SYNC_DIR := $(OUTPUT_DIR)/helm
HELM_DIR := install/helm
MANIFEST_DIR := install/manifest

.PHONY: manifest
manifest: helm-template install/manifest/supergloo.yaml update-helm-chart

# creates Chart.yaml, values.yaml, and requirements.yaml
.PHONY: helm-template
helm-template:
	mkdir -p $(MANIFEST_DIR)
	go run install/helm/supergloo/generate-values.go $(VERSION)

update-helm-chart: helm-template
ifeq ($(RELEASE),"true")
	mkdir -p $(HELM_SYNC_DIR)/charts
	helm package --destination $(HELM_SYNC_DIR)/charts $(HELM_DIR)/supergloo
	helm repo index $(HELM_SYNC_DIR)
endif

install/manifest/supergloo.yaml: helm-template
	helm template install/helm/supergloo --namespace supergloo-system --name=supergloo > $@

.PHONY: render-yaml
render-yaml: install/manifest/supergloo.yaml

#----------------------------------------------------------------------------------
# Release
#----------------------------------------------------------------------------------

.PHONY: upload-github-release-assets
upload-github-release-assets: build-cli render-yaml
	go run ci/upload_github_release_assets.go

.PHONY: push-docs
push-docs:
	go run ci/push_docs.go

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
docker: supergloo-docker

# Depends on DOCKER_IMAGES, which is set to docker if RELEASE is "true", otherwise empty (making this a no-op).
# This prevents executing the dependent targets if RELEASE is not true, while still enabling `make docker`
# to be used for local testing.
# docker-push is intended to be run by CI
docker-push: $(DOCKER_IMAGES)
ifeq ($(RELEASE),"true")
	docker push quay.io/solo-io/supergloo:$(VERSION)
endif


