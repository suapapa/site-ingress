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
FROM alpine:latest

RUN apk add --no-cache \
	certbot \
	dcron \
	busybox-initscripts

# try to renew letsencrypt ssl cert in every 12 hours
RUN SLEEPTIME=$(awk 'BEGIN{srand(); print int(rand()*(3600+1))}'); \
	echo "0 0,12 * * * root sleep $SLEEPTIME && certbot renew -q" | \
	tee -a /etc/crontabs/root > /dev/null

COPY --from=builder /build/create_ssl_cert.sh /bin/create_ssl_cert.sh
COPY --from=builder /build/app /bin/app

ENV TELEGRAM_APITOKEN="secret"
ENV TELEGRAM_ROOM_ID="secret"

EXPOSE 9001
EXPOSE 443
EXPOSE 80

WORKDIR /bin

ENTRYPOINT ["./app"]
