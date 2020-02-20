FROM alpine:3.10

# installs public root certs
# needed to establish trust with third party sources
RUN apk upgrade --update-cache \
    && apk add ca-certificates \
    && rm -rf /var/cache/apk/*

COPY mesh-networking-linux-amd64 /usr/local/bin/mesh-networking

ENTRYPOINT ["/usr/local/bin/mesh-networking"]
