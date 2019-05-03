#----------------------------------------------------------------------------------
# This portion is managed by github.com/solo-io/build
#----------------------------------------------------------------------------------
# NOTE! All make targets that use the computed values must depend on the "must"
# target to ensure the expected computed values were received
.PHONY: must
must: validate-computed-values

# Read computed values into variables that can be used by make
# Since both stdout and stderr are passed, our make targets validate the variables
RELEASE := $(shell go run build/main.go parse-env release)
VERSION := $(shell go run build/main.go parse-env version)
IMAGE_TAG := $(shell go run build/main.go parse-env image-tag)
CONTAINER_REPO_PREFIX := $(shell go run build/main.go parse-env container-prefix)
HELM_REPO := $(shell go run build/main.go parse-env helm-repo)

# use this, or the shorter alias "must", as a dependency for any target that uses
# values produced by the build tool
.PHONY: validate-computed-values
validate-computed-values:
	go run build/main.go validate-operating-parameters \
		$(RELEASE) \
		$(VERSION) \
		$(CONTAINER_REPO_PREFIX) \
		$(IMAGE_TAG) \
		$(HELM_REPO)


.PHONY: preview-computed-values
preview-computed-values: must
	echo summary of computed values - \
		release: $(RELEASE), \
		version: $(VERSION), \
		container-prefix: $(CONTAINER_REPO_PREFIX), \
		image-tag: $(IMAGE_TAG), \
		HELM-REPOSITORY: $(HELM_REPO)

#### END OF MANAGED PORTION


#----------------------------------------------------------------------------------
# Base variables
#----------------------------------------------------------------------------------

ROOTDIR := $(shell pwd)
OUTPUT_DIR ?= $(ROOTDIR)/_output
SOURCES := $(shell find . -name "*.go" | grep -v test | grep -v mock)
LDFLAGS := "-X github.com/solo-io/supergloo/pkg/version.Version=$(VERSION)"


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
	go get -u github.com/envoyproxy/protoc-gen-validate
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


#----------------------------------------------------------------------------------
# SuperGloo
#----------------------------------------------------------------------------------

$(OUTPUT_DIR)/supergloo-linux-amd64: $(SOURCES)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags=$(LDFLAGS) -o $@ cmd/supergloo/main.go
	shasum -a 256 $@ > $@.sha256

$(OUTPUT_DIR)/Dockerfile.supergloo: cmd/supergloo/Dockerfile
	cp $< $@

$(OUTPUT_DIR)/.supergloo-docker: $(OUTPUT_DIR)/supergloo-linux-amd64 $(OUTPUT_DIR)/Dockerfile.supergloo
	docker build -t $(CONTAINER_REPO_PREFIX)/supergloo:$(IMAGE_TAG) $(OUTPUT_DIR) -f $(OUTPUT_DIR)/Dockerfile.supergloo


#----------------------------------------------------------------------------------
# Admission Webhook (currently only for AWS App Mesh sidecar injection)
#----------------------------------------------------------------------------------

$(OUTPUT_DIR)/sidecar-injector-linux-amd64: $(SOURCES)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags=$(LDFLAGS) -o $@ cmd/admissionwebhook/main.go
	shasum -a 256 $@ > $@.sha256

$(OUTPUT_DIR)/Dockerfile.webhook: cmd/admissionwebhook/Dockerfile
	cp $< $@

$(OUTPUT_DIR)/.webhook-docker: $(OUTPUT_DIR)/sidecar-injector-linux-amd64 $(OUTPUT_DIR)/Dockerfile.webhook
	docker build -t $(CONTAINER_REPO_PREFIX)/sidecar-injector:$(IMAGE_TAG) $(OUTPUT_DIR) -f $(OUTPUT_DIR)/Dockerfile.webhook
	touch $@


#----------------------------------------------------------------------------------
# Mesh Discovery
#----------------------------------------------------------------------------------

$(OUTPUT_DIR)/mesh-discovery-linux-amd64: $(SOURCES)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags=$(LDFLAGS) -o $@ cmd/meshdiscovery/main.go
	shasum -a 256 $@ > $@.sha256

$(OUTPUT_DIR)/Dockerfile.mesh-discovery: cmd/meshdiscovery/Dockerfile
	cp $< $@

$(OUTPUT_DIR)/.mesh-discovery-docker: $(OUTPUT_DIR)/mesh-discovery-linux-amd64 $(OUTPUT_DIR)/Dockerfile.mesh-discovery
	docker build -t $(CONTAINER_REPO_PREFIX)/mesh-discovery:$(IMAGE_TAG) $(OUTPUT_DIR) -f $(OUTPUT_DIR)/Dockerfile.mesh-discovery


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

.PHONY: build-cli build-cli-local
build-cli: must $(OUTPUT_DIR)/supergloo-cli-linux-amd64 $(OUTPUT_DIR)/supergloo-cli-darwin-amd64 $(OUTPUT_DIR)/supergloo-cli-windows-amd64.exe

build-cli-local:
	go build -ldflags=$(LDFLAGS) -o supergloo cli/cmd/main.go


#----------------------------------------------------------------------------------
# Deployment Manifests / Helm
#----------------------------------------------------------------------------------

HELM_SYNC_DIR := $(OUTPUT_DIR)/helm
HELM_CHART_DIR := install/helm/supergloo
MANIFEST_DIR := install/manifest

.PHONY: manifest
manifest: fetch-helm-repo helm-generate render-manifest helm-package-chart push-helm-repo

### Generates Chart.yaml, values.yaml, and requirements.yaml
.PHONY: helm-generate
helm-generate:
	go run $(HELM_CHART_DIR)/generate-values.go $(VERSION) $(IMAGE_TAG)

### Packages the chart and regenerates the repo index file
helm-package-chart: helm-dirs helm-generate
	helm package --destination $(HELM_SYNC_DIR)/charts $(HELM_CHART_DIR)
	helm repo index $(HELM_SYNC_DIR)

### Render manifest for release assets
.PHONY: render-manifest
render-manifest: $(MANIFEST_DIR)/supergloo.yaml

$(MANIFEST_DIR)/supergloo.yaml: helm-dirs helm-generate
	helm template $(HELM_CHART_DIR) --namespace supergloo-system --name=supergloo > $@

### Recipes to pull/push from rempote chart repository
.PHONY: fetch-helm-repo
fetch-helm-repo: helm-dirs
	gsutil -m rsync -r $(HELM_REPO) $(HELM_SYNC_DIR)

.PHONY: push-helm-repo
push-helm-repo: helm-dirs
	gsutil -m rsync -r $(HELM_SYNC_DIR) $(HELM_REPO)

### Creates required directories
.PHONY: helm-dirs
helm-dirs:
	mkdir -p $(MANIFEST_DIR)
	mkdir -p $(HELM_SYNC_DIR)/charts


#----------------------------------------------------------------------------------
# Upload release assets and push docs to the solo-docs repo
#----------------------------------------------------------------------------------

# If this is not a release, this target is a no-op
.PHONY: upload-github-release-assets
upload-github-release-assets: build-cli render-manifest
	go run ci/upload_github_release_assets.go


# If the docs have not changed, this target is a no-op
.PHONY: push-docs
push-docs: must
	go run ci/push_docs.go


#----------------------------------------------------------------------------------
# Docker
#----------------------------------------------------------------------------------

.PHONY: docker
docker: must $(OUTPUT_DIR)/.supergloo-docker $(OUTPUT_DIR)/.webhook-docker $(OUTPUT_DIR)/.mesh-discovery-docker

.PHONY: docker-push
docker-push: must docker
	docker push $(CONTAINER_REPO_PREFIX)/supergloo:$(IMAGE_TAG)
	docker push $(CONTAINER_REPO_PREFIX)/sidecar-injector:$(IMAGE_TAG)
	docker push $(CONTAINER_REPO_PREFIX)/mesh-discovery:$(IMAGE_TAG)


