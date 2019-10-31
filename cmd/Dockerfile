FROM alpine

RUN apk upgrade --update-cache \
    && apk add ca-certificates \
    && rm -rf /var/cache/apk/*

# Install aws-cli
RUN apk -Uuv add groff less python py-pip
RUN pip install awscli
RUN apk --purge -v del py-pip
RUN rm /var/cache/apk/*

COPY mesh-discovery-linux-amd64 /usr/local/bin/mesh-discovery

ENTRYPOINT ["/usr/local/bin/mesh-discovery"]