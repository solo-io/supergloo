# Note to developers/testers
# >> to force docs publication and deployment during testing, do the following:
# cd to repo root, connect to a cluster where you want to push your images, run the following (update SUFFIX as you like)
# SUFFIX=a PROJECT_ID=solo-public TAGGED_VERSION=vtest-docs-build$SUFFIX make publish-docs -B && kubectl apply -f docs/install/manifest-latest.yaml -n docs
# >> to run host the docs locally, do the following
# cd to the docs dir
# make serve-site -B
ifeq ($(TAGGED_VERSION),)
	TAGGED_VERSION := $(shell git describe --tags --dirty --always)
	RELEASE := "false"
endif
VERSION ?= $(shell echo $(TAGGED_VERSION) | cut -c 2-)
### REPLACE with product name
IMAGE_REGISTRY := quay.io/solo-io
IMAGE_LEAF_NAME := service-mesh-hub-docs
IMAGE_REPO := $(IMAGE_REGISTRY)/$(IMAGE_LEAF_NAME)

### REPLACE with product name
PRODUCT_SCOPE := service-mesh-hub
DOCS_VERSION := latest

#----------------------------------------------------------------------------------
# Docs
#----------------------------------------------------------------------------------

.PHONY: site-common
site-common: clean
	# this theme is crucial (has the nested scoped short codes: protobuf and versioned_link_path - see dev-portal docs for use demos)
	if [ ! -d themes/hugo-theme-soloio ]; then git clone https://github.com/solo-io/hugo-theme-soloio themes/hugo-theme-soloio; fi
	# style updates for putting docs in the dev-portal repo, see details here https://github.com/solo-io/hugo-theme-soloio/commit/e0c50784a92fb7f61c635ff9a6e3a010f636f550
	git -C themes/hugo-theme-soloio checkout 91507dc0e262b5009327a741bc9fd5cc2a949c38

.PHONY: site-test
site-test: site-common
	hugo --config docs.toml

.PHONY: site-release
site-release: site-common
	HUGO_PARAMS_Version=$(VERSION) hugo --config docs.toml
	mv site site-latest

.PHONY: serve-site
serve-site: site-test
	HUGO_PARAMS_Version=$(VERSION) hugo --config docs.toml --themesDir themes server -D

.PHONY: clean
clean:
	rm -rf ./site ./resources ./site-latest ./site-versioned ./themes


.PHONY: tagged
tagged: site-release
	docker build \
		--build-arg VERSION=latest \
		--build-arg PRODUCT_SCOPE=$(PRODUCT_SCOPE) \
		--build-arg FROM_DIR=./site-latest \
		-t $(IMAGE_REPO):$(VERSION) .
	docker push $(IMAGE_REPO):$(VERSION)

.PHONY: latest
latest: site-release
	docker build \
		--build-arg VERSION=latest \
		--build-arg PRODUCT_SCOPE=$(PRODUCT_SCOPE) \
		--build-arg FROM_DIR=./site-latest \
		-t $(IMAGE_REPO):latest .
	docker push $(IMAGE_REPO):latest
