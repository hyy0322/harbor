FROM hub.byted.org/infcplibrary/golang:1.17.10-security-auth as builder
COPY . /go/src/github.com/goharbor/harbor
WORKDIR /go/src/github.com/goharbor/harbor

# set gitconfig for pulling access
COPY gitconfig /etc/gitconfig
ARG CODEBASE
ENV GIT_SSH_COMMAND $CODEBASE

RUN make compile_jobservice_binary && cp ./make/photon/jobservice/harbor_jobservice /tmp

FROM hub.byted.org/base/security.photon40:latest

RUN tdnf install -y tzdata shadow >> /dev/null \
    && tdnf clean all \
    && groupadd -r -g 10000 harbor && useradd --no-log-init -r -m -g 10000 -u 10000 harbor \
    && mkdir /harbor/
RUN tdnf upgrade -y

COPY ./make/photon/common/install_cert.sh /harbor/
COPY ./make/photon/jobservice/entrypoint.sh /harbor/
COPY --from=builder /tmp/harbor_jobservice /harbor/


RUN chown -R harbor:harbor /etc/pki/tls/certs \
    && chown -R harbor:harbor /harbor/ \
    && chmod u+x /harbor/entrypoint.sh \
    && chmod u+x /harbor/install_cert.sh \
    && chmod u+x /harbor/harbor_jobservice

WORKDIR /harbor/

USER harbor

VOLUME ["/var/log/jobs/"]

HEALTHCHECK CMD curl --fail -s http://localhost:8080/api/v1/stats || curl -sk --fail --key /etc/harbor/ssl/job_service.key --cert /etc/harbor/ssl/job_service.crt https://localhost:8443/api/v1/stats || exit 1

ENTRYPOINT ["/harbor/entrypoint.sh"]
