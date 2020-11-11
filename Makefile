#----------------------------------------------------------------------------------
# Base
#----------------------------------------------------------------------------------
OUTDIR ?= _output
PROJECT ?= gloo-mesh

DOCKER_REPO ?= soloio
GLOOMESH_IMAGE ?= $(DOCKER_REPO)/gloo-mesh
CA_IMAGE ?= $(DOCKER_REPO)/cert-agent

SOURCES := $(shell find . -name "*.go" | grep -v test.go)
RELEASE := "true"
ifeq ($(TAGGED_VERSION),)
	TAGGED_VERSION := $(shell git describe --tags --dirty --always)
	RELEASE := "false"
endif

VERSION ?= $(shell echo $(TAGGED_VERSION) | cut -c 2-)
.PHONY: print-version
print-version:
ifeq ($(TAGGED_VERSION),)
	exit 1
endif
	echo $(VERSION)

LDFLAGS := "-X github.com/solo-io/$(PROJECT)/pkg/common/version.Version=$(VERSION)"
GCFLAGS := all="-N -l"

print-info:
	@echo RELEASE: $(RELEASE)
	@echo TAGGED_VERSION: $(TAGGED_VERSION)
	@echo VERSION: $(VERSION)

#----------------------------------------------------------------------------------
# Code Generation
#----------------------------------------------------------------------------------

DEPSGOBIN=$(shell pwd)/$(OUTDIR)/.bin
export PATH:=$(DEPSGOBIN):$(PATH)
export GOBIN:=$(DEPSGOBIN)

.PHONY: fmt
fmt:
	goimports -w $(shell ls -d */ | grep -v vendor)

.PHONY: mod-download
mod-download:
	go mod download

.PHONY: clear-vendor-any
clear-vendor-any:
	rm -rf vendor_any

# Dependencies for code generation
.PHONY: install-go-tools
install-go-tools: mod-download
	mkdir -p $(DEPSGOBIN)
	go install istio.io/tools/cmd/protoc-gen-jsonshim
	go install github.com/gogo/protobuf/protoc-gen-gogo
	go install github.com/golang/protobuf/protoc-gen-go
	go install github.com/solo-io/protoc-gen-ext
	go install github.com/golang/mock/mockgen
	go install golang.org/x/tools/cmd/goimports
	go install github.com/onsi/ginkgo/ginkgo
	go install github.com/gobuffalo/packr/packr

# Call all generated code targets
.PHONY: generated-code
generated-code: operator-gen \
				manifest-gen \
				go-generate \
				generated-reference-docs \
				fmt
	go mod tidy

#----------------------------------------------------------------------------------
# Go generate
#----------------------------------------------------------------------------------

# Run go-generate on all sub-packages
go-generate:
	go generate -v ./...

#----------------------------------------------------------------------------------
# Operator Code Generation
#----------------------------------------------------------------------------------

# Generate Operator Code
.PHONY: operator-gen
operator-gen: clear-vendor-any
	go run -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) codegen/generate.go

#----------------------------------------------------------------------------------
# Docs Code Generation
#----------------------------------------------------------------------------------

# Generate Reference documentation
.PHONY: generated-reference-docs
generated-reference-docs: clear-vendor-any
	go run codegen/docs/docsgen.go

#----------------------------------------------------------------------------------
# Build
#----------------------------------------------------------------------------------

.PHONY: build-all-images
build-all-images: gloo-mesh-image cert-agent-image

#----------------------------------------------------------------------------------
# Build gloo-mesh controller + image
#----------------------------------------------------------------------------------

# for local development only; to build docker image, use gloo-mesh-linux-amd-64
.PHONY: gloo-mesh
gloo-mesh: $(OUTDIR)/gloo-mesh
$(OUTDIR)/gloo-mesh: $(SOURCES)
	go build -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) -o $@ cmd/gloo-mesh/main.go

.PHONY: gloo-mesh-linux-amd64
gloo-mesh-linux-amd64: $(OUTDIR)/gloo-mesh-linux-amd64
$(OUTDIR)/gloo-mesh-linux-amd64: $(SOURCES)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) -o $@ cmd/gloo-mesh/main.go

# build image with gloo-mesh binary
# this is an alternative to using operator-gen to build the image
.PHONY: gloo-mesh-image
gloo-mesh-image: gloo-mesh-linux-amd64
	cp $(OUTDIR)/gloo-mesh-linux-amd64 build/gloo-mesh/ && \
	docker build -t $(GLOOMESH_IMAGE):$(VERSION) build/gloo-mesh/
	rm build/gloo-mesh/gloo-mesh-linux-amd64

.PHONY: gloo-mesh-image-push
gloo-mesh-image-push: gloo-mesh-image
ifeq ($(RELEASE),"true")
	docker push $(GLOOMESH_IMAGE):$(VERSION)
endif

.PHONY: gloo-mesh-image-load
gloo-mesh-image-load: gloo-mesh-image
    kind load docker-image --name mgmt-cluster $(GLOOMESH_IMAGE):$(VERSION)

#----------------------------------------------------------------------------------
# Build cert-agent + image
#----------------------------------------------------------------------------------

# for local development only; to build docker image, use cert-agent-linux-amd-64
.PHONY: cert-agent
cert-agent: $(OUTDIR)/cert-agent
$(OUTDIR)/cert-agent: $(SOURCES)
	go build -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) -o $@ cmd/cert-agent/main.go

.PHONY: cert-agent-linux-amd64
cert-agent-linux-amd64: $(OUTDIR)/cert-agent-linux-amd64
$(OUTDIR)/cert-agent-linux-amd64: $(SOURCES)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) -o $@ cmd/cert-agent/main.go

# build image with cert-agent binary
# this is an alternative to using operator-gen to build the image
.PHONY: cert-agent-image
cert-agent-image: cert-agent-linux-amd64
	cp $(OUTDIR)/cert-agent-linux-amd64 build/cert-agent/ && \
	docker build -t $(CA_IMAGE):$(VERSION) build/cert-agent/
	rm build/cert-agent/cert-agent-linux-amd64

.PHONY: cert-agent-image-push
cert-agent-image-push: cert-agent-image
ifeq ($(RELEASE),"true")
	docker push $(CA_IMAGE):$(VERSION)
endif

.PHONY: cert-agent-image-load
cert-agent-image-load: cert-agent-image
    kind load docker-image --name mgmt-cluster $(CA_IMAGE):$(VERSION)


#----------------------------------------------------------------------------------
# Build gloo-mesh cli (meshctl)
#----------------------------------------------------------------------------------

.PHONY: meshctl-linux-amd64
meshctl-linux-amd64: $(OUTDIR)/meshctl-linux-amd64
$(OUTDIR)/meshctl-linux-amd64: $(SOURCES)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux packr build -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) -o $@ cmd/meshctl/main.go

.PHONY: meshctl-darwin-amd64
meshctl-darwin-amd64: $(OUTDIR)/meshctl-darwin-amd64
$(OUTDIR)/meshctl-darwin-amd64: $(SOURCES)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=darwin packr build -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) -o $@ cmd/meshctl/main.go

.PHONY: meshctl-windows-amd64
meshctl-windows-amd64: $(OUTDIR)/meshctl-windows-amd64.exe
$(OUTDIR)/meshctl-windows-amd64.exe: $(SOURCES)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=windows packr build -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) -o $@ cmd/meshctl/main.go

.PHONY: build-cli
build-cli: install-go-tools meshctl-linux-amd64 meshctl-darwin-amd64 meshctl-windows-amd64

.PHONY: install-cli
install-cli:
	packr build -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) -o ${GOPATH}/bin/meshctl cmd/meshctl/main.go

#----------------------------------------------------------------------------------
# Push images
#----------------------------------------------------------------------------------

.PHONY: push-all-images
push-all-images: gloo-mesh-image-push cert-agent-image-push

#----------------------------------------------------------------------------------
# Helm chart
#----------------------------------------------------------------------------------
HELM_ROOTDIR := install/helm
# Include helm makefile so its targets can be ran from the root of this repo
include install/helm/helm.mk

# Generate Manifests from Helm Chart
.PHONY: chart-gen
chart-gen: clear-vendor-any
	go run -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) codegen/generate.go -chart

.PHONY: manifest-gen
manifest-gen: install/gloo-mesh-default.yaml
install/gloo-mesh-default.yaml: chart-gen
	helm template --include-crds --namespace gloo-mesh install/helm/gloo-mesh > $@

#----------------------------------------------------------------------------------
# Test
#----------------------------------------------------------------------------------

# run all tests
# set TEST_PKG to run a specific test package
.PHONY: run-tests
run-tests:
	ginkgo -r -failFast -trace $(GINKGOFLAGS) \
		-ldflags=$(LDFLAGS) \
		-gcflags=$(GCFLAGS) \
		-progress \
		-race \
		-compilers=4 \
		-skipPackage=$(SKIP_PACKAGES) $(TEST_PKG)

# regen code+manifests, image build+push, and run all tests
# convenience for local testing
.PHONY: test-everything
test-everything: clean-generated-code generated-code manifest-gen run-tests

##----------------------------------------------------------------------------------
## Release
##----------------------------------------------------------------------------------

.PHONY: upload-github-release-assets
upload-github-release-assets: build-cli
ifeq ($(RELEASE),"true")
	go run ci/upload_github_release_assets.go
endif

#----------------------------------------------------------------------------------
# Clean
#----------------------------------------------------------------------------------

.PHONY: clean
clean: clean-helm
	rm -f install/gloo-mesh-default.yaml
	rm -rf  _output/ vendor_any/

.PHONY: clean-generated-code
clean-generated-code:
	find pkg -name "*.pb.go" -type f -delete
	find pkg -name "*.hash.go" -type f -delete
	find pkg -name "*.gen.go" -type f -delete
	find pkg -name "*deepcopy.go" -type f -delete
