FROM scratch

ARG CI_PROJECT_URL

COPY --from=registry.synyx.cloud/gitlabci/golang:latest /etc/pki/tls/certs/ca-bundle.crt /etc/ssl/certs/ca-certificates.crt

WORKDIR /go
COPY ./gonagdash /go/gonagdash
EXPOSE 8988
ENTRYPOINT ["/go/gonagdash"]

LABEL org.opencontainers.image.authors="Jonathan Buch <jbuch@synyx.de>" \
      org.opencontainers.image.url=${CI_PROJECT_URL} \
      org.opencontainers.image.vendor="synyx GmbH & Co. KG" \
      org.opencontainers.image.title="Go Nagdash" \
      org.opencontainers.image.description="Nagios et al Dashboard"
