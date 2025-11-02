# build stage
FROM golang:1.25-alpine AS builder

ARG PROGRAM_VER=dev-docker

ENV CGO_ENABLED=0

# RUN apt-get -qq update && \
# 	apt-get install -yqq upx

COPY . /build
WORKDIR /build

RUN go build -ldflags "-X main.programVer=${PROGRAM_VER}" -o app
# RUN strip /build/app
# RUN upx -q -9 /build/app

#---

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /app
COPY --from=builder /build/app .
COPY ./asset asset
COPY ./conf conf

EXPOSE 8080

ENTRYPOINT ["/app/app"]
