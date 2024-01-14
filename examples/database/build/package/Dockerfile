ARG BASE_IMAGE=dockerhub.cisco.com/vms-platform-dev-docker/go-scratch:1.20
## Build Container ##
FROM dockerhub.cisco.com/vms-platform-dev-docker/golang:1.20 AS builder
ARG PRIVATE_MODS
ARG VERSION
ARG GITKEY
WORKDIR /go/src/
ADD . /go/src/
RUN mkdir -p -m 0600 ~/.ssh && \
    echo "StrictHostKeyChecking no " > /root/.ssh/config && \
    echo "Host cto-github.cisco.com" >> ~/.ssh/config  && \
    echo "              HostName cto-github.cisco.com" >> ~/.ssh/config  && \
    echo "              User git" >> ~/.ssh/config  && \
    echo "             IdentityFile /root/github.key" >> ~/.ssh/config
COPY ${GITKEY} /root/github.key
RUN chmod 600 /root/github.key
RUN make init-once init-cli drop-replace; \
    make clean build PRIVATE_MODS="$PRIVATE_MODS" VERSION="$VERSION"

FROM ${BASE_IMAGE}
EXPOSE 9898
WORKDIR /service
# copy service
COPY --chown=10001:10001 --from=builder /go/src/dist/ /service/
COPY --chown=10001:10001 build/package/dockerlaunch.sh /service/dockerlaunch.sh
RUN chmod 755 /service/dockerlaunch.sh
USER 10001

ENTRYPOINT ["/service/dockerlaunch.sh"]

