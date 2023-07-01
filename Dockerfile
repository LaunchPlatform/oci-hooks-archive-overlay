# ref: https://github.com/containers/podman/blob/main/contrib/podmanimage/stable/Containerfile
FROM python:3.11.4-alpine3.18
RUN apk update && apk add --no-cache \
        fuse \
        fuse-overlayfs \
        podman \
        s6-overlay \
        s6-overlay-syslogd

ARG _REPO_URL="https://raw.githubusercontent.com/containers/podman/v4.5.1/contrib/podmanimage/stable"
ADD $_REPO_URL/containers.conf /etc/containers/containers.conf
ADD $_REPO_URL/podman-containers.conf /home/podman/.config/containers/containers.conf

RUN mkdir /project
WORKDIR /project

COPY . /project/
RUN pip install poetry && \
    poetry config virtualenvs.create false && \
    poetry install --no-interaction --no-ansi

# Note VOLUME options must always happen after the chown call above
# RUN commands can not modify existing volumes
VOLUME /var/lib/containers

RUN mkdir -p /var/lib/shared/overlay-images \
             /var/lib/shared/overlay-layers \
             /var/lib/shared/vfs-images \
             /var/lib/shared/vfs-layers && \
    touch /var/lib/shared/overlay-images/images.lock && \
    touch /var/lib/shared/overlay-layers/layers.lock && \
    touch /var/lib/shared/vfs-images/images.lock && \
    touch /var/lib/shared/vfs-layers/layers.lock

ENTRYPOINT ["/init"]
