FROM golang:alpine AS build-env
WORKDIR /go/src/control
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN go build -o /go/bin/control

FROM frolvlad/alpine-glibc
RUN apk update && apk add --no-cache ca-certificates su-exec iptables

RUN addgroup -g 1000 app && adduser -D -u 1000 -G app app

COPY docker-entrypoint.sh /usr/local/bin/
COPY --from=build-env /go/bin/control /app/control

RUN chmod +x /usr/local/bin/docker-entrypoint.sh
RUN chown -R app:app /app

WORKDIR /app

ENTRYPOINT ["docker-entrypoint.sh"]

