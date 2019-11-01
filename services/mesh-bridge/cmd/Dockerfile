FROM alpine

RUN apk upgrade --update-cache \
    && apk add ca-certificates \
    && rm -rf /var/cache/apk/*

COPY mesh-bridge-linux-amd64 /usr/local/bin/mesh-bridge

ENTRYPOINT ["/usr/local/bin/mesh-bridge"]
