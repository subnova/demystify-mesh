FROM golang:alpine AS build-env
WORKDIR /go/src/colorui
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN go build -o /go/bin/colorui

FROM envoyproxy/envoy-alpine:v1.11.1 as envoy

FROM frolvlad/alpine-glibc
RUN apk update && apk add --no-cache ca-certificates su-exec iptables

RUN addgroup -g 1000 app && adduser -D -u 1000 -G app app
RUN addgroup -g 1337 envoy && adduser -D -u 1337 -G envoy envoy
RUN mkdir /etc/envoy

COPY docker-entrypoint.sh /usr/local/bin/
COPY --from=build-env /go/bin/colorui /app/colorui
COPY --from=envoy /usr/local/bin/envoy /usr/local/bin/envoy
COPY ./*.html /app

RUN chmod +x /usr/local/bin/docker-entrypoint.sh
RUN chown -R app:app /app

RUN mkdir -p /var/log/envoy && chown envoy:envoy /var/log/envoy

WORKDIR /app

ENTRYPOINT ["docker-entrypoint.sh"]

