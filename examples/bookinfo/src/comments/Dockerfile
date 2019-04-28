FROM alpine

RUN apk upgrade --update-cache \
    && apk add ca-certificates \
    && rm -rf /var/cache/apk/*

COPY comments /usr/local/bin/comments

ENTRYPOINT ["/usr/local/bin/comments"]