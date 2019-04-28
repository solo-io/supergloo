.PHONY: comments comments-docker

TAG?=:latest

comments: main.go
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o $@ .

comments-docker: comments
	docker build -t docker.io/ilackarms/comments$(TAG) .

comments-push: comments-docker
	docker push docker.io/ilackarms/comments$(TAG)