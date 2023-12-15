FROM golang:1.20 as build

WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 go build -o tuwat ./cmd/tuwat

FROM gcr.io/distroless/static-debian11

COPY --from=build /go/src/app/tuwat /tuwat
EXPOSE 8988
ENTRYPOINT ["/tuwat"]

LABEL org.opencontainers.image.authors="Jonathan Buch <jbuch@synyx.de>" \
      org.opencontainers.image.url="https://github.com/synyx/tuwat" \
      org.opencontainers.image.vendor="synyx GmbH & Co. KG" \
      org.opencontainers.image.title="Tuwat Dashboard" \
      org.opencontainers.image.description="Tuwat Operations Dashboard"
