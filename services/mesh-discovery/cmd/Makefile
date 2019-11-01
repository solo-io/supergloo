
# TODO - consider moving this to the root repo and standardizing with the Gloo build format
DOCKER_REPO ?= quay.io/solo-io
APP_NAME := mesh-discovery
VERSION ?= v0.0.0-dev
LDFLAGS := ""
GCFLAGS := all="-N -l"

docker-push:
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) -o $(APP_NAME)-linux-amd64 main.go
	docker build -t $(DOCKER_REPO)/$(APP_NAME):$(VERSION) .
	docker push $(DOCKER_REPO)/$(APP_NAME):$(VERSION)
