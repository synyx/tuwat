FROM gcr.io/distroless/static-debian11

WORKDIR /go
COPY ./tuwat /tuwat
USER 1000
EXPOSE 8988
ENTRYPOINT ["/tuwat"]

LABEL org.opencontainers.image.authors="Jonathan Buch <jbuch@synyx.de>" \
      org.opencontainers.image.url="https://github.com/synyx/tuwat" \
      org.opencontainers.image.vendor="synyx GmbH & Co. KG" \
      org.opencontainers.image.title="Tuwat Dashboard" \
      org.opencontainers.image.description="Tuwat Operations Dashboard"
