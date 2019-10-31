.PHONY: pin-repos
pin-repos:
	go run ci/pin_repos.go

.PHONY: update-deps
update-deps:
	go get -u golang.org/x/tools/cmd/goimports
	go get -u github.com/gogo/protobuf/gogoproto
	go get -u github.com/gogo/protobuf/protoc-gen-gogo
#	go get -u github.com/envoyproxy/protoc-gen-validate
	go get -u github.com/paulvollmer/2gobytes
	go get -u github.com/golang/mock/gomock
	go install github.com/golang/mock/mockgen


SUBDIRS:=$(shell ls -d -- */ | grep -v vendor)
.PHONY: generated-code
generated-code:
	rm -rf docs/api
	go generate ./...
	gofmt -w $(SUBDIRS)
	goimports -w $(SUBDIRS)
