FROM alpine

RUN apk upgrade --update-cache \
    && apk add ca-certificates \
    && rm -rf /var/cache/apk/*

COPY supergloo-linux-amd64 /usr/local/bin/supergloo

ENTRYPOINT ["/usr/local/bin/supergloo"]