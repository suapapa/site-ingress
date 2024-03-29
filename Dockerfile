# build stage
FROM golang:1.19 as builder

ARG PROGRAM_VER=dev-docker

ENV CGO_ENABLED=0

RUN apt-get -qq update && \
	apt-get install -yqq upx

COPY . /build
WORKDIR /build

RUN go build -ldflags "-X main.programVer=${PROGRAM_VER}" -o app
RUN strip /build/app
RUN upx -q -9 /build/app

# ---
FROM alpine

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/app /bin/app

EXPOSE 8080

WORKDIR /bin

ENTRYPOINT ["./app"]
