package generator

var gcloudDocker = `FROM docker:17.12.0-ce as static-docker-source

FROM alpine:3.10
ARG CLOUD_SDK_VERSION=270.0.0
ENV CLOUD_SDK_VERSION=$CLOUD_SDK_VERSION

ENV PATH /google-cloud-sdk/bin:$PATH
COPY --from=static-docker-source /usr/local/bin/docker /usr/local/bin/docker
RUN apk --no-cache add \
        curl \
        python \
        py-crcmod \
        bash \
        libc6-compat \
        openssh-client \
        git \
        gnupg \
    && curl -O https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-${CLOUD_SDK_VERSION}-linux-x86_64.tar.gz && \
    tar xzf google-cloud-sdk-${CLOUD_SDK_VERSION}-linux-x86_64.tar.gz && \
    rm google-cloud-sdk-${CLOUD_SDK_VERSION}-linux-x86_64.tar.gz && \
    gcloud config set core/disable_usage_reporting true && \
    gcloud config set component_manager/disable_update_check true && \
    gcloud config set metrics/environment github_docker_image && \
    gcloud --version
VOLUME ["/root/.config"]`
