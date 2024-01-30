ARG BASE_IMAGE=debian:bookworm
ARG BUILDER_IMAGE=golang:1.21-bookworm

## Build Container ##
FROM ${BUILDER_IMAGE} AS builder
ARG PRIVATE_MODS
ARG VERSION
ARG GITKEY
WORKDIR /go/src/
ADD . /go/src/
RUN mkdir -p -m 0600 ~/.ssh; \
    echo "StrictHostKeyChecking no " > /root/.ssh/config; \
    ssh-keyscan github.com >> ~/.ssh/known_hosts;

RUN if [ ! -z "$GITKEY" -a -e $GITKEY ]; then \
      cp -fp /go/src/$GITKEY /root/github.key; \
      chmod 400 /root/github.key; \
      IFS=,; for mod in $PRIVATE_MODS; do \
          hostname=`echo "$mod"| awk -F/ '{print $1}'`; \
          echo "Host $hostname" >> ~/.ssh/config; \
          echo "              HostName $hostname" >> ~/.ssh/config; \
          echo "              User git" >> ~/.ssh/config; \
          echo "              IdentityFile /root/github.key" >> ~/.ssh/config; \
      done; \
    fi

RUN --mount=type=ssh \
    make init-once init-cli drop-replace && \
    go mod tidy && \
    make clean build PRIVATE_MODS="$PRIVATE_MODS" VERSION="$VERSION";

FROM ${BASE_IMAGE}
EXPOSE 8900
WORKDIR /service
# copy service
COPY --chown=10001:10001 --from=builder /go/src/dist/ /service/
COPY --chown=10001:10001 build/package/dockerlaunch.sh /service/dockerlaunch.sh
RUN chmod 755 /service/dockerlaunch.sh
USER 10001

ENTRYPOINT ["/service/dockerlaunch.sh"]

