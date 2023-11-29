FROM hub.byted.org/infcplibrary/golang:1.17.10-security-auth as builder

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
COPY ./make/photon/registry/entrypoint.sh /home/harbor
COPY --from=builder /registry/src/github.com/docker/distribution/bin/registry /usr/bin/registry_DO_NOT_USE_GC

RUN chown -R harbor:harbor /etc/pki/tls/certs \
    && chown harbor:harbor /home/harbor/entrypoint.sh && chmod u+x /home/harbor/entrypoint.sh \
    && chown harbor:harbor /home/harbor/install_cert.sh && chmod u+x /home/harbor/install_cert.sh \
    && chown harbor:harbor /usr/bin/registry_DO_NOT_USE_GC && chmod u+x /usr/bin/registry_DO_NOT_USE_GC

HEALTHCHECK CMD curl --fail -s http://127.0.0.1:5000 || curl -k --fail -s https://127.0.0.1:5443 || exit 1

USER harbor

ENTRYPOINT ["/home/harbor/entrypoint.sh"]

VOLUME ["/storage"]

