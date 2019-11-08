
#----------------------------------------------------------------------------------
# some-binary
# Generated with args: {"BinaryNameBase":"some-binary","ImageName":"app-prefix-some-binary","OutputFile":"","BinaryDir":"services/some-binary/cmd","DockerOutputFilepath":""}
#----------------------------------------------------------------------------------
SOME_BINARY_DIR=services/some-binary/cmd
SOME_BINARY_SOURCES=$(shell find $(SOME_BINARY_DIR) -name "*.go" | grep -v test | grep -v generated.go)

$(OUTPUT_DIR)/some-binary-linux-amd64: $(SOME_BINARY_SOURCES)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) -o $@ $(SOME_BINARY_DIR)/main.go

.PHONY: some-binary
some-binary: $(OUTPUT_DIR)/some-binary-linux-amd64

$(OUTPUT_DIR)/Dockerfile.some-binary: $(SOME_BINARY_DIR)/Dockerfile
	cp $< $@

.PHONY: some-binary-docker
some-binary-docker: $(OUTPUT_DIR)/.some-binary-docker

$(OUTPUT_DIR)/.some-binary-docker: $(OUTPUT_DIR)/some-binary-linux-amd64 $(OUTPUT_DIR)/Dockerfile.some-binary
	docker build -t quay.io/solo-io/app-prefix-some-binary:$(VERSION) $(call get_test_tag_option,some-binary) $(OUTPUT_DIR) -f $(OUTPUT_DIR)/Dockerfile.some-binary
	touch $@
