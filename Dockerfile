FROM gcr.io/distroless/static-debian11

WORKDIR /go
COPY ./tuwat /go/tuwat
EXPOSE 8988
ENTRYPOINT ["/go/tuwat"]

LABEL org.opencontainers.image.authors="Jonathan Buch <jbuch@synyx.de>" \
      org.opencontainers.image.url="https://github.com/synyx/tuwat" \
      org.opencontainers.image.vendor="synyx GmbH & Co. KG" \
      org.opencontainers.image.title="Tuwat Dashboard" \
      org.opencontainers.image.description="Tuwat Operations Dashboard"
