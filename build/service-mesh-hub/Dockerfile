FROM alpine:3.11.6

# installs public root certs
# needed to establish trust with third party sources
RUN apk upgrade --update-cache \
    && apk add ca-certificates \
    && rm -rf /var/cache/apk/*

COPY service-mesh-hub-linux-amd64 /usr/local/bin/service-mesh-hub

ENTRYPOINT ["/usr/local/bin/service-mesh-hub"]
