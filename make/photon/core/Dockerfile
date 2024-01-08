FROM hub.byted.org/harbor/harbor-swagger:v0.25.0 as swagger
FROM hub.byted.org/infcplibrary/golang:1.17.10-security-auth as builder

COPY . /go/src/github.com/goharbor/harbor
WORKDIR /go/src/github.com/goharbor/harbor

COPY --from=swagger /usr/bin/swagger /usr/bin/swagger
# set gitconfig for pulling access
COPY gitconfig /etc/gitconfig
ARG CODEBASE
ENV GIT_SSH_COMMAND $CODEBASE

RUN make compile_core_binary && cp ./make/photon/core/harbor_core /tmp && \
    make versions_prepare && cp ./make/photon/prepare/versions /tmp

FROM hub.byted.org/base/security.photon40:latest

RUN tdnf install -y tzdata shadow >> /dev/null \
    && tdnf clean all \
    && groupadd -r -g 10000 harbor && useradd --no-log-init -r -m -g 10000 -u 10000 harbor \
    && mkdir /harbor/
RUN tdnf upgrade -y

HEALTHCHECK CMD curl --fail -s http://127.0.0.1:8080/api/v2.0/ping || curl -k --fail -s https://127.0.0.1:8443/api/v2.0/ping || exit 1
COPY ./make/photon/common/install_cert.sh /harbor/
COPY ./make/photon/core/entrypoint.sh /harbor/
COPY --from=builder /tmp/harbor_core /harbor/
COPY ./src/core/views /harbor/views
COPY ./make/migrations /harbor/migrations
COPY ./icons /harbor/icons

RUN chown -R harbor:harbor /etc/pki/tls/certs \
    && chown -R harbor:harbor /harbor/ \
    && chmod u+x /harbor/entrypoint.sh \
    && chmod u+x /harbor/install_cert.sh \
    && chmod u+x /harbor/harbor_core

WORKDIR /harbor/
USER harbor
ENTRYPOINT ["/harbor/entrypoint.sh"]
COPY --from=builder /tmp/versions /harbor/
