FROM hub.byted.org/infcplibrary/golang:1.17.10-security-auth as builder
COPY . /harbor
WORKDIR /harbor
# set gitconfig for pulling access
COPY gitconfig /etc/gitconfig
ARG CODEBASE
ENV GIT_SSH_COMMAND $CODEBASE

RUN CGO_ENABLED=0 make compile_registryctl_binary

FROM hub.byted.org/infcplibrary/golang:1.17.10-security-auth as rbuilder

WORKDIR /registry
COPY ./make/photon/registry/builder.new /registry/builder

# set gitconfig for pulling access
COPY gitconfig /etc/gitconfig
ARG CODEBASE
ENV GIT_SSH_COMMAND $CODEBASE

RUN chmod 755 /registry/builder && \
    /registry/builder release-v2.8

FROM hub.byted.org/base/security.photon40:latest

RUN tdnf install -y shadow >> /dev/null \
    && tdnf clean all \
    && groupadd -r -g 10000 harbor && useradd --no-log-init -m -g 10000 -u 10000 harbor \
    && mkdir -p /etc/registry
RUN tdnf upgrade -y

COPY ./make/photon/common/install_cert.sh /home/harbor
COPY --from=rbuilder /registry/src/github.com/docker/distribution/bin/registry /usr/bin/registry_DO_NOT_USE_GC
COPY ./make/photon/registryctl/start.sh /home/harbor
COPY --from=builder /harbor/make/photon/registryctl/harbor_registryctl /home/harbor

RUN chown -R harbor:harbor /etc/pki/tls/certs \
    && chown harbor:harbor /home/harbor/harbor_registryctl && chmod u+x /home/harbor/harbor_registryctl \
    && chown harbor:harbor /usr/bin/registry_DO_NOT_USE_GC && chmod u+x /usr/bin/registry_DO_NOT_USE_GC \
    && chown harbor:harbor /home/harbor/start.sh && chmod u+x /home/harbor/start.sh \
    && chown harbor:harbor /home/harbor/install_cert.sh && chmod u+x /home/harbor/install_cert.sh


HEALTHCHECK CMD curl --fail -s http://127.0.0.1:8080/api/health || curl -sk --fail --key /etc/harbor/ssl/registryctl.key --cert /etc/harbor/ssl/registryctl.crt https://127.0.0.1:8443/api/health || exit 1

VOLUME ["/var/lib/registry"]

ENTRYPOINT ["/home/harbor/start.sh"]

USER harbor
