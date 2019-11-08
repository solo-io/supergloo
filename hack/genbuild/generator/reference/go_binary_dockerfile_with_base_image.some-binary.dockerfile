# GENERATED FILE - DO NOT EDIT
FROM quay.io/solo-io/mc-base-image:some-version

COPY some-binary-linux-amd64 /usr/local/bin/some-binary

ENTRYPOINT ["/usr/local/bin/some-binary"]
