FROM alpine

RUN apk upgrade --update-cache \
    && apk add ca-certificates \
    && rm -rf /var/cache/apk/*

COPY test-service /usr/local/bin/test-service

ENTRYPOINT ["/usr/local/bin/test-service"]