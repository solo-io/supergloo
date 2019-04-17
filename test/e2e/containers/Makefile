.PHONY: test-service

TAG?=:latest

test-service/test-service: test-service/main.go
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o $@ ./test-service

test-service: test-service/test-service
	docker build -t quay.io/solo-io/test-service$(TAG) ./test-service

test-service-push: test-service
	docker push quay.io/solo-io/test-service$(TAG)