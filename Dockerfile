# build stage
FROM golang:1.19 as builder

ENV CGO_ENABLED=0

RUN apt-get -qq update && \
	apt-get install -yqq upx

COPY . /build
WORKDIR /build

RUN go build -o app
RUN strip /build/app
RUN upx -q -9 /build/app

# ---
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/app /bin/app

ENV TELEGRAM_APITOKEN="secret"
ENV TELEGRAM_ROOM_ID="secret"

EXPOSE 8080

WORKDIR /bin

ENTRYPOINT ["./app"]
