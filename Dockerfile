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
# The backend server-renders the SPA routes, so it needs the Vite-built index.html as its
# template (see render.go). Only index.html is copied — assets stay in the frontend image.
FROM node:20-alpine AS web

WORKDIR /web
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ .
# Same as frontend/Dockerfile: public/model is a dev symlink to ../../asset/.
RUN rm -rf public/model
RUN npm run build

#---

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /app
COPY --from=builder /build/app .
COPY ./asset asset
COPY ./conf conf
COPY --from=web /web/dist/index.html static/index.html

EXPOSE 8080

ENTRYPOINT ["/app/app"]
